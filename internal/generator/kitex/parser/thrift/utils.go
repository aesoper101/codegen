package thrift

import (
	"github.com/cloudwego/thriftgo/generator/golang"
	"path/filepath"
)

func getPthBy(root *golang.Scope, pkgName string) (pth string, alisa string) {
	deps, _ := root.ResolveImports()
	for path, name := range deps {
		if name == pkgName {
			return path, pkgName
		}
	}

	for path, name := range deps {
		if name == "" && filepath.Base(path) == pkgName {
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
