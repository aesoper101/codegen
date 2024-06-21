package hertz

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aesoper101/codegen/internal/generator"
	"github.com/aesoper101/codegen/internal/generator/hertz/parser"
	"github.com/aesoper101/codegen/pkg/utils"
	"github.com/aesoper101/x/filex"
	"log/slog"
	"path/filepath"
	"strings"
)

var _ generator.Generator = (*Generator)(nil)

type Generator struct {
	ctx            context.Context
	generateConfig *GenerateConfig
	parser         parser.Parser
}

func NewGenerator(ctx context.Context, configFiles []string) (*Generator, error) {
	generateConfig, err := loadConfig(ctx, configFiles...)
	if err != nil {
		return nil, err
	}

	idlParser, err := generateConfig.Options.NewParser()
	if err != nil {
		return nil, err
	}

	return &Generator{
		ctx:            ctx,
		parser:         idlParser,
		generateConfig: generateConfig,
	}, nil
}

func (g *Generator) Name() string {
	return "hertz"
}

func (g *Generator) Generate() error {
	packages, err := g.parser.Parse(g.generateConfig.Files...)
	if err != nil {
		return err
	}

	if len(packages) == 0 {
		return fmt.Errorf("no package found in the idl files")
	}

	templates := g.generateConfig.Templates
	var files []*generator.File
	for _, tpl := range templates {
		tpl.Preprocess()

		fs, err := g.generateVirtualFiles(tpl, packages)
		if err != nil {
			return err
		}
		files = append(files, fs...)
	}

	if err := g.generateFiles(files); err != nil {
		return g.rollbackFiles(files)
	}
	return nil
}

// 生成实际文件
func (g *Generator) generateFiles(files []*generator.File) error {
	for _, file := range files {
		if err := filex.CreateFileFromString(file.Name, true, file.Content); err != nil {
			return err
		}
	}
	return nil
}

// 回滚文件
func (g *Generator) rollbackFiles(files []*generator.File) error {
	for _, file := range files {
		if file.IsNew {
			if err := filex.DeleteFile(file.Name); err != nil {
				return err
			}
		} else {
			if err := filex.CreateFileFromString(file.Name, true, file.BackupContent); err != nil {
				return err
			}
		}
	}
	return nil
}

// 生成虚拟文件
func (g *Generator) generateVirtualFiles(tpl *Template, packages parser.Packages) ([]*generator.File, error) {
	switch tpl.Type {
	case Custom:
		return g.generateForCustom(tpl, packages)
	case File:
		return g.generateLoopFile(tpl, packages)
	case Package:
		return g.generateLoopPackage(tpl, packages)
	case Service:
		return g.generateLoopService(tpl, packages)
	case Method:
		return g.generateLoopMethod(tpl, packages)
	default:
		return nil, fmt.Errorf("unsupported generate mode: %s", tpl.Type.String())
	}
}

func (g *Generator) generateForCustom(tpl *Template, packages parser.Packages) (files []*generator.File, err error) {
	renderInfo := FilePathRenderInfo{
		GoModule:  g.generateConfig.GoModule,
		ModuleDir: g.generateConfig.ModuleDir,
	}

	filePath, err := renderFilePath(tpl, renderInfo)
	if err != nil {
		return nil, err
	}

	data := &CustomizedFileForCustom{
		FilePath:           filePath,
		FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
		Packages:           packages,
		Extras:             tpl.Extras,
		FilePathRenderInfo: renderInfo,
	}

	fileAlreadyExists := filex.IsExists(filePath)
	var content, oldContent string
	if fileAlreadyExists {
		oldContent, err = readFileContent(filePath)
		if err != nil {
			return nil, err
		}
	}
	switch tpl.UpdateBehavior.Type {
	case Skip:
		if fileAlreadyExists {
			return
		}
		fallthrough
	case Cover:
		content, err = renderFileContent(tpl, data)
		if err != nil {
			return nil, err
		}

		if content, err = removeUnUseImportAndFormatIfGoFile(filePath, content); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("in mode %s, the custom template does not support the update behavior %s", tpl.Type.String(), tpl.UpdateBehavior.Type.String())
	}

	files = append(files, &generator.File{
		Name:          filePath,
		Content:       content,
		IsNew:         !fileAlreadyExists,
		BackupContent: oldContent,
	})
	return files, nil
}

func (g *Generator) generateLoopFile(tpl *Template, packages parser.Packages) (files []*generator.File, err error) {
	for _, idl := range packages.Files() {
		file, err := g.generateForFile(tpl, idl)
		if err != nil {
			return nil, err
		}

		files = append(files, file...)
	}

	return files, nil
}

func (g *Generator) generateForFile(tpl *Template, idl *parser.File) (files []*generator.File, err error) {
	renderInfo := FilePathRenderInfo{
		GoModule:    g.generateConfig.GoModule,
		ModuleDir:   g.generateConfig.ModuleDir,
		Namespace:   idl.Package.Namespace,
		PackageName: idl.Package.Namespace,
		IDLName:     idl.Filename(),
	}

	filePath, err := renderFilePath(tpl, renderInfo)
	if err != nil {
		return nil, err
	}

	data := &CustomizedFileForIDL{
		FilePath:           filePath,
		FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
		IdlInfo:            idl,
		Extras:             tpl.Extras,
		FilePathRenderInfo: renderInfo,
	}

	fileAlreadyExists := filex.IsExists(filePath)
	var content, oldContent string
	if fileAlreadyExists {
		oldContent, err = readFileContent(filePath)
		if err != nil {
			return nil, err
		}
	}

	switch tpl.UpdateBehavior.Type {
	case Skip:
		if fileAlreadyExists {
			return
		}
		fallthrough
	case Cover:
		content, err = renderFileContent(tpl, data)
		if err != nil {
			return nil, err
		}
	case Append:
		var appendContent string
		for _, service := range idl.Services {
			renderInfo.ServiceName = service.Name
			insertData := &CustomizedFileForService{
				FilePath:           filePath,
				FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
				Service:            service,
				Extras:             tpl.Extras,
				FilePathRenderInfo: renderInfo,
			}

			insertKey, err := renderInsertKey(tpl, insertData)
			if err != nil {
				return nil, err
			}
			if strings.Contains(content, insertKey) {
				continue
			}

			imptSlice, err := getInsertImportContent(tpl, data, []byte(content))
			if err != nil {
				return nil, err
			}

			for _, impt := range imptSlice {
				if strings.Contains(content, impt[1]) {
					continue
				}
				content, err = utils.AddImportForContent(content, impt[0], impt[1])
				// insert import error do not influence the generated file
				if err != nil {
					slog.Warn(fmt.Sprintf("can not add import(%s) for file(%s), err: %v\n", impt[1], filePath, err))
				}
			}
			appendContent, err = appendUpdateFile(tpl, data, appendContent)
			if err != nil {
				return nil, err
			}
		}

		if len(appendContent) > 0 {
			buf := bytes.NewBufferString(content)
			buf.WriteString(appendContent)
			content = buf.String()
		}
	default:
		return nil, fmt.Errorf("in mode %s, the idl template does not support the update behavior %s", tpl.Type.String(), tpl.UpdateBehavior.Type.String())
	}

	if content, err = removeUnUseImportAndFormatIfGoFile(filePath, content); err != nil {
		return nil, err
	}

	files = append(files, &generator.File{
		Name:          filePath,
		Content:       content,
		IsNew:         !fileAlreadyExists,
		BackupContent: oldContent,
	})

	return files, nil
}

func (g *Generator) generateLoopPackage(tpl *Template, packages parser.Packages) (files []*generator.File,
	err error) {
	for _, pkg := range packages {
		file, err := g.generateForPackage(tpl, pkg)
		if err != nil {
			return nil, err
		}
		files = append(files, file...)
	}
	return files, nil
}

func (g *Generator) generateForPackage(tpl *Template, pkgInfo *parser.Package) (files []*generator.File, err error) {
	renderInfo := FilePathRenderInfo{
		GoModule:    g.generateConfig.GoModule,
		ModuleDir:   g.generateConfig.ModuleDir,
		PackageName: pkgInfo.ImportPackage,
		Namespace:   pkgInfo.Namespace,
	}

	filePath, err := renderFilePath(tpl, renderInfo)
	if err != nil {
		return nil, err
	}

	data := &CustomizedFileForPackage{
		FilePath:           filePath,
		FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
		Package:            pkgInfo,
		Extras:             tpl.Extras,
		FilePathRenderInfo: renderInfo,
	}

	fileAlreadyExists := filex.IsExists(filePath)
	var content, oldContent string
	if fileAlreadyExists {
		oldContent, err = readFileContent(filePath)
		if err != nil {
			return nil, err
		}
	}

	switch tpl.UpdateBehavior.Type {
	case Skip:
		if fileAlreadyExists {
			return
		}
		fallthrough
	case Cover:
		content, err = renderFileContent(tpl, data)
		if err != nil {
			return nil, err
		}
	case Append:
		var appendContent string
		for _, file := range pkgInfo.Files {
			renderInfo.IDLName = file.Filename()
			insertData := &CustomizedFileForIDL{
				FilePath:           filePath,
				FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
				IdlInfo:            file,
				Extras:             tpl.Extras,
				FilePathRenderInfo: renderInfo,
			}

			insertKey, err := renderInsertKey(tpl, insertData)
			if err != nil {
				return nil, err
			}
			if strings.Contains(content, insertKey) {
				continue
			}

			imptSlice, err := getInsertImportContent(tpl, data, []byte(content))
			if err != nil {
				return nil, err
			}

			for _, impt := range imptSlice {
				if strings.Contains(content, impt[1]) {
					continue
				}
				content, err = utils.AddImportForContent(content, impt[0], impt[1])
				// insert import error do not influence the generated file
				if err != nil {
					slog.Warn(fmt.Sprintf("can not add import(%s) for file(%s), err: %v\n", impt[1], filePath, err))
				}
			}
			appendContent, err = appendUpdateFile(tpl, data, appendContent)
			if err != nil {
				return nil, err
			}
		}
		if len(appendContent) > 0 {
			buf := bytes.NewBufferString(content)
			buf.WriteString(appendContent)
			content = buf.String()
		}
	default:
		return nil, fmt.Errorf("in mode %s, the package template does not support the update behavior %s", tpl.Type.String(), tpl.UpdateBehavior.Type.String())
	}

	if content, err = removeUnUseImportAndFormatIfGoFile(filePath, content); err != nil {
		return nil, err
	}

	files = append(files, &generator.File{
		Name:          filePath,
		Content:       content,
		IsNew:         !fileAlreadyExists,
		BackupContent: oldContent,
	})

	return nil, nil
}

func (g *Generator) generateLoopService(tpl *Template, packages parser.Packages) (files []*generator.File, err error) {
	for _, pkg := range packages.Services() {
		file, err := g.generateForService(tpl, pkg)
		if err != nil {
			return nil, err
		}
		files = append(files, file...)
	}
	return files, nil
}

func (g *Generator) generateForService(tpl *Template, service *parser.Service) (files []*generator.File, err error) {
	renderInfo := FilePathRenderInfo{
		GoModule:    g.generateConfig.GoModule,
		ModuleDir:   g.generateConfig.ModuleDir,
		PackageName: service.File.Package.ImportPackage,
		Namespace:   service.File.Package.Namespace,
		ServiceName: service.Name,
	}

	filePath, err := renderFilePath(tpl, renderInfo)
	if err != nil {
		return nil, err
	}

	data := &CustomizedFileForService{
		FilePath:           filePath,
		FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
		Service:            service,
		Extras:             tpl.Extras,
		FilePathRenderInfo: renderInfo,
	}

	fileAlreadyExists := filex.IsExists(filePath)
	var content, oldContent string
	if fileAlreadyExists {
		oldContent, err = readFileContent(filePath)
		if err != nil {
			return nil, err
		}
	}

	switch tpl.UpdateBehavior.Type {
	case Skip:
		if fileAlreadyExists {
			return
		}
		fallthrough
	case Cover:
		content, err = renderFileContent(tpl, data)
		if err != nil {
			return nil, err
		}
	case Append:
		var appendContent string
		for _, method := range service.Methods {
			renderInfo.MethodName = method.Name
			insertData := &CustomizedFileForMethod{
				FilePath:           filePath,
				FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
				Method:             method,
				ClientMethod:       method.Service.GetClientMethod(method.Name),
				Extras:             tpl.Extras,
				FilePathRenderInfo: renderInfo,
			}

			insertKey, err := renderInsertKey(tpl, insertData)
			if err != nil {
				return nil, err
			}
			if strings.Contains(content, insertKey) {
				continue
			}

			imptSlice, err := getInsertImportContent(tpl, data, []byte(content))
			if err != nil {
				return nil, err
			}

			for _, impt := range imptSlice {
				if strings.Contains(content, impt[1]) {
					continue
				}
				content, err = utils.AddImportForContent(content, impt[0], impt[1])
				// insert import error do not influence the generated file
				if err != nil {
					slog.Warn(fmt.Sprintf("can not add import(%s) for file(%s), err: %v\n", impt[1], filePath, err))
				}
			}
			appendContent, err = appendUpdateFile(tpl, data, appendContent)
			if err != nil {
				return nil, err
			}
		}

		if len(appendContent) > 0 {
			buf := bytes.NewBufferString(content)
			buf.WriteString(appendContent)
			content = buf.String()
		}

	default:
		return nil, fmt.Errorf("in mode %s, the service template does not support the update behavior %s", tpl.Type.String(), tpl.UpdateBehavior.Type.String())
	}

	if content, err = removeUnUseImportAndFormatIfGoFile(filePath, content); err != nil {
		return nil, err
	}

	files = append(files, &generator.File{
		Name:          filePath,
		Content:       content,
		IsNew:         !fileAlreadyExists,
		BackupContent: oldContent,
	})

	return files, nil
}

func (g *Generator) generateLoopMethod(tpl *Template, packages parser.Packages) (files []*generator.File, err error) {
	for _, pkg := range packages.Methods() {
		file, err := g.generateForMethod(tpl, pkg)
		if err != nil {
			return nil, err
		}
		files = append(files, file...)
	}
	return files, nil
}

func (g *Generator) generateForMethod(tpl *Template, method *parser.HttpMethod) (files []*generator.File, err error) {
	renderInfo := FilePathRenderInfo{
		GoModule:    g.generateConfig.GoModule,
		ModuleDir:   g.generateConfig.ModuleDir,
		PackageName: method.Service.File.Package.ImportPackage,
		Namespace:   method.Service.File.Package.Namespace,
		ServiceName: method.Service.Name,
		MethodName:  method.Name,
	}

	filePath, err := renderFilePath(tpl, renderInfo)
	if err != nil {
		return nil, err
	}

	data := &CustomizedFileForMethod{
		FilePath:           filePath,
		FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
		Method:             method,
		ClientMethod:       method.Service.GetClientMethod(method.Name),
		Extras:             tpl.Extras,
		FilePathRenderInfo: renderInfo,
	}

	fileAlreadyExists := filex.IsExists(filePath)
	var content, oldContent string
	if fileAlreadyExists {
		oldContent, err = readFileContent(filePath)
		if err != nil {
			return nil, err
		}
	}

	switch tpl.UpdateBehavior.Type {
	case Skip:
		if fileAlreadyExists {
			return
		}
		fallthrough
	case Cover:
		content, err = renderFileContent(tpl, data)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("in mode %s, the method template does not support the update behavior %s", tpl.Type.String(), tpl.UpdateBehavior.Type.String())
	}

	if content, err = removeUnUseImportAndFormatIfGoFile(filePath, content); err != nil {
		return nil, err
	}

	files = append(files, &generator.File{
		Name:          filePath,
		Content:       content,
		IsNew:         !fileAlreadyExists,
		BackupContent: oldContent,
	})

	return files, nil
}
