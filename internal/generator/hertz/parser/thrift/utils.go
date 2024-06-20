package thrift

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/generator/hertz/parser"
	"github.com/cloudwego/thriftgo/generator/golang"
	"path/filepath"
)

func checkHttpMethodSamePath(ms []*parser.HttpMethod, m *parser.HttpMethod) error {
	for _, method := range ms {
		if method.Path == m.Path {
			return fmt.Errorf("http method %s has same path %s with method %s", m.Name, m.Path, method.Name)
		}
	}
	return nil
}

func getPthBy(root *golang.Scope, pkgRefName string) (pth string, alisa string) {
	deps, _ := root.ResolveImports()
	for path, name := range deps {
		if name == pkgRefName {
			return path, pkgRefName
		}
	}

	for path, name := range deps {
		if name == "" && filepath.Base(path) == pkgRefName {
			return path, ""
		}
	}

	return "", ""
}

func mergeImports(imports map[string]string, newImports map[string]string) map[string]string {
	for path, alias := range newImports {
		if _, ok := imports[path]; !ok {
			imports[path] = alias
		}
	}

	return imports
}
