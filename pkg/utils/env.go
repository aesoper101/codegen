package utils

import (
	"github.com/aesoper101/x/filepathx"
	"github.com/mitchellh/go-homedir"
)

func CodegenHome() string {
	dir, _ := homedir.Dir()
	return filepathx.JoinPath(dir, ".codegen")
}

func CodegenCacheHome() string {
	return filepathx.JoinPath(CodegenHome(), "cache")
}

func CodegenTemplateHome() string {
	return filepathx.JoinPath(CodegenCacheHome(), "templates")
}

func CodegenConfigHome() string {
	return filepathx.JoinPath(CodegenHome(), "config")
}
