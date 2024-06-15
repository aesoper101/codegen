package custom

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/aesoper101/codegen/internal/generator"
	"github.com/aesoper101/codegen/internal/generator/hertz/protobuf"
	"github.com/aesoper101/codegen/internal/generator/hertz/thrift"
	"github.com/aesoper101/codegen/internal/generator/hertz/types"
	"github.com/aesoper101/codegen/pkg/utils"
	cond "github.com/aesoper101/x/condition"
	"github.com/aesoper101/x/filex"
	"github.com/spf13/afero"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var _ generator.Generator = (*Generator)(nil)

var (
	ErrNoSupportIdlType = errors.New("not support idl type")
)

type Generator struct {
	idlType          IdlType
	packagePrefix    string
	idlPaths         []string
	tplPaths         []string
	defaultTemplates []*File
}

func New(opts ...Option) (*Generator, error) {
	g := &Generator{
		idlType:  Thrift,
		idlPaths: []string{"./idl"},
		tplPaths: []string{"./templates"},
	}
	for _, opt := range opts {
		if err := opt(g); err != nil {
			return nil, err
		}
	}

	return g, nil
}

func (g *Generator) Name() string {
	return "HertzCustom"
}

func (g *Generator) Generate() error {
	var convertor types.Convertor
	switch g.idlType {
	case Proto:
		convertor = protobuf.NewConvertor(g.packagePrefix)
	case Thrift:
		convertor = thrift.NewConvertor(g.packagePrefix)
	default:
		return ErrNoSupportIdlType
	}

	httpPackages, err := convertor.Convert(g.idlPaths...)
	if err != nil {
		return err
	}

	return g.generate(httpPackages)
}

func (g *Generator) generate(packages []*types.HttpPackage) error {
	tplConfigs, err := loadTemplates(g.tplPaths...)
	if err != nil {
		return err
	}

	files := g.defaultTemplates

	for _, c := range tplConfigs {
		extras := make(Extras)
		if c.Extras != nil {
			extras = c.Extras
		}

		for _, tpl := range c.Layouts {
			var err error
			var tempFiles []*File
			preTemplate(tpl)
			switch tpl.Type {
			case Custom:
				tempFiles, err = g.generateByCustomFile(extras, tpl, packages)
			case Package:
				tempFiles, err = g.generateByLoopPackage(extras, tpl, packages)
			case Service:
				tempFiles, err = g.generateByLoopService(extras, tpl, packages)
			case Method:
				tempFiles, err = g.generateByLoopMethod(extras, tpl, packages)
			default:
				err = errors.New("not support template type")
			}

			if err != nil {
				return err
			}

			files = append(files, tempFiles...)
		}
	}

	if err := g.generateFiles(files); err != nil {
		return g.rollback(files)
	}
	return nil
}

func (g *Generator) generateFiles(files []*File) error {
	for _, file := range files {
		if err := filex.CreateFileFromString(file.Name, true, file.Content); err != nil {
			slog.Warn(fmt.Sprintf("create file %s failed, err: %v", file.Name, err))
			return err
		}
	}

	return nil
}

func (g *Generator) generateBackupFiles(files []*File) error {
	memFs := afero.NewMemMapFs()
	for _, file := range files {
		if file.IsNew {
			continue
		}
		dir := filepath.Dir(file.Name)
		if err := memFs.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}

		if err := afero.WriteFile(memFs, file.Name, []byte(file.BackupContent), 0666); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) rollback(files []*File) error {
	for _, file := range files {
		if file.IsNew {
			if err := filex.DeleteFile(file.Name); err != nil {
				slog.Warn(fmt.Sprintf("remove file %s failed, err: %v", file.Name, err))
			}
		} else {
			if err := filex.CreateFileFromString(file.Name, true, file.BackupContent); err != nil {
				slog.Warn(fmt.Sprintf("create file %s failed, err: %v", file.Name, err))
			}
		}
	}

	return nil
}

func (g *Generator) generateByLoopPackage(extras Extras, tpl *Template, packages []*types.HttpPackage) ([]*File, error) {
	var files []*File

	for _, pkg := range packages {
		renderInfo := FilePathRenderInfo{
			Extras:      extras,
			PackageName: pkg.Package,
		}
		filePath, err := renderFilePath(tpl, renderInfo)
		if err != nil {
			return nil, err
		}

		data := CustomizedFileForIDL{
			FilePath:           filePath,
			FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
			IdlInfo:            pkg,
			Extras:             cond.Ternary(tpl.Extras != nil, tpl.Extras, Extras{}),
			FilePathRenderInfo: renderInfo,
		}

		fileAlreadyExists := filex.IsExists(filePath)
		var content, oldContent string
		if fileAlreadyExists {
			oldContent, err = readFileContent(filePath)
			if err != nil {
				return nil, err
			}

			content = oldContent
		}
		switch tpl.UpdateBehavior.Type {
		case Cover:
			content, err = renderFileContent(tpl, data)
			if err != nil {
				return nil, err
			}
		case Skip:
			if fileAlreadyExists {
				slog.Warn(fmt.Sprintf("file %s already exists, skip it", filePath))
				return nil, nil
			}
		case Append:
			switch tpl.UpdateBehavior.AppendKey {
			case AppendKeyService:
				var appendContent string
				for _, service := range pkg.Services {
					renderInfo.ServiceName = service.Name
					data := CustomizedFileForService{
						FilePath:           filePath,
						FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
						Service:            service,
						Extras:             cond.Ternary(tpl.Extras != nil, tpl.Extras, Extras{}),
						FilePathRenderInfo: renderInfo,
					}
					insertKey, err := renderInsertKey(tpl, data)
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

					if len(appendContent) > 0 {
						buf := bytes.NewBufferString(content)
						buf.WriteString(appendContent)
						content = buf.String()
					}
				}
			default:
				return nil, fmt.Errorf("in mode 'package', only support 'service' append key")
			}
		}

		if content, err = removeUnUseImportAndFormatIfGoFile(filePath, content); err != nil {
			return nil, err
		}

		files = append(files, &File{
			Name:          filePath,
			Content:       content,
			IsNew:         !fileAlreadyExists,
			BackupContent: oldContent,
		})
	}
	return files, nil
}

func (g *Generator) generateByLoopService(extras Extras, tpl *Template, packages []*types.HttpPackage) ([]*File, error) {
	var files []*File
	var services []*types.Service
	for _, pkg := range packages {
		services = append(services, pkg.Services...)
	}

	for _, service := range services {
		renderInfo := FilePathRenderInfo{
			Extras:      extras,
			PackageName: service.From().Package,
			ServiceName: service.Name,
		}
		filePath, err := renderFilePath(tpl, renderInfo)
		if err != nil {
			return nil, err
		}

		exists := filex.IsExists(filePath)

		var content, oldContent string
		if exists {
			oldContent, err = readFileContent(filePath)
			if err != nil {
				return nil, err
			}
			content = oldContent
		}

		data := CustomizedFileForService{
			FilePath:           filePath,
			FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
			Service:            service,
			Extras:             cond.Ternary(tpl.Extras != nil, tpl.Extras, Extras{}),
			FilePathRenderInfo: renderInfo,
		}

		switch tpl.UpdateBehavior.Type {
		case Skip:
			if exists {
				slog.Warn(fmt.Sprintf("file %s already exists, skip it", filePath))
				continue
			}
			fallthrough
		case Cover:
			content, err = renderFileContent(tpl, data)
			if err != nil {
				return nil, err
			}
		case Append:
			switch tpl.UpdateBehavior.AppendKey {
			case AppendKeyMethod:
				var appendContent string
				for _, method := range service.Methods {
					renderInfo.MethodName = method.Name
					data := CustomizedFileForMethod{
						FilePath:           filePath,
						FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
						Method:             method,
						Extras:             cond.Ternary(tpl.Extras != nil, tpl.Extras, Extras{}),
						FilePathRenderInfo: renderInfo,
					}
					insertKey, err := renderInsertKey(tpl, data)
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

					if len(appendContent) > 0 {
						buf := bytes.NewBufferString(content)
						buf.WriteString(appendContent)
						content = buf.String()
					}
				}
			default:
				return nil, fmt.Errorf("in mode 'service', only support 'method' append key")
			}
		}

		if content, err = removeUnUseImportAndFormatIfGoFile(filePath, content); err != nil {
			return nil, err
		}

		files = append(files, &File{
			Name:          filePath,
			Content:       content,
			IsNew:         !exists,
			BackupContent: oldContent,
		})
	}
	return files, nil
}

func (g *Generator) generateByLoopMethod(extras Extras, tpl *Template, packages []*types.HttpPackage) ([]*File, error) {
	var files []*File

	var services []*types.Service
	for _, pkg := range packages {
		services = append(services, pkg.Services...)
	}

	var methods []*types.HttpMethod
	for _, service := range services {
		methods = append(methods, service.Methods...)
	}

	for _, method := range methods {
		renderInfo := FilePathRenderInfo{
			Extras:      extras,
			PackageName: method.From().From().Package,
			ServiceName: method.From().Name,
			MethodName:  method.Name,
		}
		filePath, err := renderFilePath(tpl, renderInfo)
		if err != nil {
			return nil, err
		}

		exists := filex.IsExists(filePath)

		var content, oldContent string
		if exists {
			oldContent, err = readFileContent(filePath)
			if err != nil {
				return nil, err
			}
			content = oldContent
		}

		data := CustomizedFileForMethod{
			FilePath:           filePath,
			FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
			Method:             method,
			Extras:             cond.Ternary(tpl.Extras != nil, tpl.Extras, Extras{}),
			FilePathRenderInfo: renderInfo,
		}

		switch tpl.UpdateBehavior.Type {
		case Skip:
			if exists {
				slog.Warn(fmt.Sprintf("file %s already exists, skip it", filePath))
				continue
			}
			fallthrough
		case Cover:
			content, err = renderFileContent(tpl, data)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("in mode 'method', only support 'cover' and 'skip' update behavior")
		}

		if content, err = removeUnUseImportAndFormatIfGoFile(filePath, content); err != nil {
			return nil, err
		}

		files = append(files, &File{
			Name:          filePath,
			Content:       content,
			IsNew:         !exists,
			BackupContent: oldContent,
		})

	}
	return files, nil
}

func (g *Generator) generateByCustomFile(extras Extras, tpl *Template, packages []*types.HttpPackage) (files []*File, err error) {
	renderInfo := FilePathRenderInfo{
		Extras: extras,
	}

	filePath, err := renderFilePath(tpl, renderInfo)
	if err != nil {
		return nil, err
	}

	data := CustomizedFileForCustom{
		FilePath:           filePath,
		FilePackage:        utils.SplitPackage(filepath.Dir(filePath), ""),
		Data:               packages,
		Extras:             cond.Ternary(tpl.Extras != nil, tpl.Extras, Extras{}),
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
		return nil, fmt.Errorf("in mode 'custom', only support 'cover' and 'skip' update behavior")
	}

	files = append(files, &File{
		Name:          filePath,
		Content:       content,
		IsNew:         !fileAlreadyExists,
		BackupContent: oldContent,
	})
	return files, nil
}

type File struct {
	Name          string
	Content       string
	IsNew         bool
	BackupContent string
}

func (f *File) Clone() *File {
	return &File{
		Name:          f.Name,
		Content:       f.Content,
		IsNew:         f.IsNew,
		BackupContent: f.BackupContent,
	}
}

type FilePathRenderInfo struct {
	Extras
	PackageName string // package name, when template type is custom
	ServiceName string // service name, when template type is service or method
	MethodName  string // method name, when template type is method
}

type CustomizedFileForService struct {
	FilePath    string
	FilePackage string
	Service     *types.Service
	Extras      Extras
	FilePathRenderInfo
}

type CustomizedFileForMethod struct {
	FilePath    string
	FilePackage string
	Method      *types.HttpMethod
	Extras      Extras
	FilePathRenderInfo
}

type CustomizedFileForIDL struct {
	FilePath    string
	FilePackage string
	IdlInfo     *types.HttpPackage
	Extras      Extras
	FilePathRenderInfo
}

type CustomizedFileForCustom struct {
	FilePath    string
	FilePackage string
	Data        []*types.HttpPackage
	Extras      Extras
	FilePathRenderInfo
}

func preTemplate(tpl *Template) {
	if tpl.Extras != nil {
		tpl.Extras = Extras{}
	}
	if tpl.Type == "" {
		tpl.Type = Custom
	}
	if tpl.UpdateBehavior.Type == "" {
		tpl.UpdateBehavior.Type = Skip
	}
}
