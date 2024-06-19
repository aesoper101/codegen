package v2

import (
	"path/filepath"
	"strings"
)

type File struct {
	Package *Package

	FilePath    string
	FilePackage string

	Services []*Service
}

func (f *File) ReferenceName() string {
	return strings.ToLower(strings.TrimSuffix(filepath.Base(f.FilePath), filepath.Ext(f.FilePackage)))
}

// TODO: 在render时，计算依赖map
