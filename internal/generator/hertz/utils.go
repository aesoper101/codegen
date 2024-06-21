package hertz

import (
	"bytes"
	"fmt"
	"github.com/aesoper101/x/templatex"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var funcMap = func() template.FuncMap {
	m := template.FuncMap{
		"Split":     strings.Split,
		"Trim":      strings.Trim,
		"EqualFold": strings.EqualFold,
	}
	for key, f := range templatex.TxtFuncMap() {
		m[key] = f
	}
	return m
}()

func renderFilePath(tplInfo *Template, filePathRenderInfo FilePathRenderInfo) (string, error) {
	tpl, err := template.New(tplInfo.Path).Funcs(funcMap).Parse(tplInfo.Path)
	if err != nil {
		return "", fmt.Errorf("parse file path template(%s) failed, err: %v", tplInfo.Path, err)
	}
	filePath := bytes.NewBuffer(nil)
	err = tpl.Execute(filePath, filePathRenderInfo)
	if err != nil {
		return "", fmt.Errorf("render file path template(%s) failed, err: %v", tplInfo.Path, err)
	}

	return filePath.String(), nil
}

func renderInsertKey(tplInfo *Template, data interface{}) (string, error) {
	tpl, err := template.New(tplInfo.UpdateBehavior.InsertKey).Funcs(funcMap).Parse(tplInfo.UpdateBehavior.InsertKey)
	if err != nil {
		return "", fmt.Errorf("parse insert key template(%s) failed, err: %v", tplInfo.UpdateBehavior.InsertKey, err)
	}
	insertKey := bytes.NewBuffer(nil)
	err = tpl.Execute(insertKey, data)
	if err != nil {
		return "", fmt.Errorf("render insert key template(%s) failed, err: %v", tplInfo.UpdateBehavior.InsertKey, err)
	}

	return insertKey.String(), nil
}

func renderFileContent(tplInfo *Template, data interface{}) (string, error) {
	tpl, err := template.New(tplInfo.Body).Funcs(funcMap).Parse(tplInfo.Body)
	if err != nil {
		return "", fmt.Errorf("parse file content template(%s) failed, err: %v", tplInfo.Body, err)
	}
	fileContent := bytes.NewBuffer(nil)
	err = tpl.Execute(fileContent, data)
	if err != nil {
		return "", fmt.Errorf("render file content template(%s) failed, err: %v", tplInfo.Body, err)
	}

	return fileContent.String(), nil
}

func isGoFile(filePath string) bool {
	return filepath.Ext(filePath) == ".go"
}

func removeUnUseImportAndFormatIfGoFile(filePath, content string) (string, error) {
	if !isGoFile(filePath) {
		return content, nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", []byte(content), parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parse file failed, err: %v", err)
	}

	unused := getUnusedImports(f)
	for ipath, name := range unused {
		astutil.DeleteNamedImport(fset, f, name, ipath)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return "", fmt.Errorf("format go file failed, err: %v", err)
	}

	return buf.String(), nil
}

func getUnusedImports(file *ast.File) map[string]string {
	imported := map[string]*ast.ImportSpec{}
	used := map[string]bool{}

	ast.Walk(visitFn(func(node ast.Node) {
		if node == nil {
			return
		}
		switch v := node.(type) {
		case *ast.ImportSpec:
			if v.Name != nil {
				imported[v.Name.Name] = v
				break
			}
			ipath := strings.Trim(v.Path.Value, `"`)
			used[ipath] = true
		}
	}), file)

	unused := map[string]string{}
	for name, spec := range imported {
		ipath := strings.Trim(spec.Path.Value, `"`)
		if !used[ipath] {
			unused[ipath] = name
		}
	}

	return unused
}

type visitFn func(node ast.Node)

func (fn visitFn) Visit(node ast.Node) ast.Visitor {
	fn(node)
	return fn
}

func readFileContent(filePath string) (string, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func getInsertImportContent(tplInfo *Template, renderInfo interface{}, fileContent []byte) ([][2]string, error) {
	importContent, err := renderImportTpl(tplInfo, renderInfo)
	if err != nil {
		return nil, err
	}
	var imptSlice [][2]string
	for _, impt := range importContent {
		// import has to format
		// 1. alias "import"
		// 2. "import"
		// 3. import (can not contain '"')
		impt = strings.TrimSpace(impt)
		if !strings.Contains(impt, "\"") { // 3. import (can not contain '"')
			if bytes.Contains(fileContent, []byte(impt)) {
				continue
			}
			imptSlice = append(imptSlice, [2]string{"", impt})
		} else {
			if !strings.HasSuffix(impt, "\"") {
				return nil, fmt.Errorf("import can not has suffix \"\"\", for file: %s", tplInfo.Path)
			}
			if strings.HasPrefix(impt, "\"") { // 2. "import"
				if bytes.Contains(fileContent, []byte(impt[1:len(impt)-1])) {
					continue
				}
				imptSlice = append(imptSlice, [2]string{"", impt[1 : len(impt)-1]})
			} else { // 3. alias "import"
				idx := strings.Index(impt, "\"")
				if idx == -1 {
					return nil, fmt.Errorf("error import format for file: %s", tplInfo.Path)
				}
				if bytes.Contains(fileContent, []byte(impt[idx+1:len(impt)-1])) {
					continue
				}
				imptSlice = append(imptSlice, [2]string{impt[:idx], impt[idx+1 : len(impt)-1]})
			}
		}
	}

	return imptSlice, nil
}

// renderImportTpl will render import template
// it will return the []string, like blow:
// ["import", alias "import", import]
// other format will be error
func renderImportTpl(tplInfo *Template, data interface{}) ([]string, error) {
	var importList []string
	for _, impt := range tplInfo.UpdateBehavior.ImportTpl {
		tpl, err := template.New(impt).Funcs(funcMap).Parse(impt)
		if err != nil {
			return nil, fmt.Errorf("parse import template(%s) failed, err: %v", impt, err)
		}
		imptContent := bytes.NewBuffer(nil)
		err = tpl.Execute(imptContent, data)
		if err != nil {
			return nil, fmt.Errorf("render import template(%s) failed, err: %v", impt, err)
		}
		importList = append(importList, imptContent.String())
	}
	var ret []string
	for _, impts := range importList {
		// 'import render result' may have multiple imports
		if strings.Contains(impts, "\n") {
			for _, impt := range strings.Split(impts, "\n") {
				ret = append(ret, impt)
			}
		} else {
			ret = append(ret, impts)
		}
	}

	return ret, nil
}

// renderAppendContent used to render append content for 'update' command
func renderAppendContent(tplInfo *Template, renderInfo interface{}) (string, error) {
	tpl, err := template.New(tplInfo.Path).Funcs(funcMap).Parse(tplInfo.UpdateBehavior.AppendTpl)
	if err != nil {
		return "", fmt.Errorf("parse append content template(%s) failed, err: %v", tplInfo.Path, err)
	}
	appendContent := bytes.NewBuffer(nil)
	err = tpl.Execute(appendContent, renderInfo)
	if err != nil {
		return "", fmt.Errorf("render append content template(%s) failed, err: %v", tplInfo.Path, err)
	}

	return appendContent.String(), nil
}

// appendUpdateFile used to append content to file for 'update' command
func appendUpdateFile(tplInfo *Template, renderInfo interface{}, fileContent string) (string, error) {
	// render insert content
	appendContent, err := renderAppendContent(tplInfo, renderInfo)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(nil)
	_, err = buf.WriteString(fileContent)
	if err != nil {
		return "", fmt.Errorf("write file(%s) failed, err: %v", tplInfo.Path, err)
	}
	// "\r\n" && "\n" has the same suffix
	if !bytes.HasSuffix(buf.Bytes(), []byte("\n")) {
		_, err = buf.WriteString("\n")
		if err != nil {
			return "", fmt.Errorf("write file(%s) line break failed, err: %v", tplInfo.Path, err)
		}
	}
	_, err = buf.WriteString(appendContent)
	if err != nil {
		return "", fmt.Errorf("append file(%s) failed, err: %v", tplInfo.Path, err)
	}

	return buf.String(), nil
}

func writeString(buf *bytes.Buffer, strs ...string) error {
	for _, b := range strs {
		_, err := buf.WriteString(b)
		if err != nil {
			return err
		}
	}

	return nil
}
