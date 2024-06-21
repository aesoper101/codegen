package parser

import (
	"path/filepath"
	"strings"
)

type File struct {
	Package *Package `json:"-"`

	IDLPath string

	Services []*Service

	Deps map[string]string // import path -> alias

	FileDescriptor interface{} // parsed file descriptor from IDL, such as desc.FileDescriptor Or parser.Thrift
}

func (f *File) ReferenceName() string {
	return strings.ToLower(strings.TrimSuffix(filepath.Base(f.IDLPath), filepath.Ext(f.IDLPath)))
}

func (f *File) Filename() string {
	return f.ReferenceName()
}
