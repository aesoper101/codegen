package utils

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"path/filepath"
)

func AddImport(file, alias, impt string) (string, error) {
	fset := token.NewFileSet()
	path, _ := filepath.Abs(file)
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("can not parse ast for file: %s, err: %v", path, err)
	}

	return addImport(fset, f, alias, impt)
}

func AddImportForContent(fileContent string, alias, impt string) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", fileContent, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("can not parse ast for file: %s, err: %v", fileContent, err)
	}

	return addImport(fset, f, alias, impt)
}

func addImport(fset *token.FileSet, f *ast.File, alias, impt string) (string, error) {
	added := astutil.AddNamedImport(fset, f, alias, impt)
	if !added {
		return "", fmt.Errorf("can not add import \"%s\" for file: %s", impt, f.Name.Name)
	}
	var output []byte
	buffer := bytes.NewBuffer(output)
	err := format.Node(buffer, fset, f)
	if err != nil {
		return "", fmt.Errorf("can not add import for file: %s, err: %v", f.Name.Name, err)
	}

	return buffer.String(), nil
}
