package utils

import (
	"github.com/aesoper101/x/filepathext"
	"github.com/mitchellh/go-homedir"
)

func CodegenHome() string {
	dir, _ := homedir.Dir()
	return filepathext.JoinPath(dir, ".codegen")
}

func CodegenCacheHome() string {
	return filepathext.JoinPath(CodegenHome(), "cache")
}

func CodegenTemplateHome() string {
	return filepathext.JoinPath(CodegenCacheHome(), "templates")
}

func CodegenConfigHome() string {
	return filepathext.JoinPath(CodegenHome(), "config")
}
