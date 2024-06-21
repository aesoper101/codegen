package kitex

import (
	"github.com/aesoper101/codegen/internal/generator/kitex/parser"
)

type FilePathRenderInfo struct {
	ModuleDir   string // module dir
	GoModule    string // go module name
	IDLName     string // idl name, when template type is idl
	Namespace   string // namespace, when template type is package
	PackageName string // package name, when template type is  package
	ServiceName string // service name, when template type is service or method
	MethodName  string // method name, when template type is method
}

type CustomizedFileForPackage struct {
	FilePath    string
	FilePackage string
	Package     *parser.Package
	Extras      Extras // from template
	FilePathRenderInfo
}

type CustomizedFileForService struct {
	FilePath    string
	FilePackage string
	Service     *parser.Service
	Extras      Extras // from template
	FilePathRenderInfo
}

type CustomizedFileForMethod struct {
	FilePath    string
	FilePackage string
	Method      *parser.Method
	Extras      Extras // from template
	FilePathRenderInfo
}

type CustomizedFileForIDL struct {
	FilePath    string
	FilePackage string
	IdlInfo     *parser.File
	Extras      Extras // from template
	FilePathRenderInfo
}

type CustomizedFileForCustom struct {
	FilePath    string
	FilePackage string
	Packages    parser.Packages
	Extras      Extras // from template
	FilePathRenderInfo
}

type Extras map[string]string

func (e Extras) Get(key string) string {
	v, _ := e[key]
	return v
}
