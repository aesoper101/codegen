package parser

type PkgInfo struct {
	PkgName    string
	PkgRefName string
	ImportPath string
}

type Service struct {
	PkgInfo
	Name         string
	BaseDomain   string
	ServiceGroup string
	ServicePath  string
	Comment      string
	Methods      []*HTTPMethod
}

type HTTPMethod struct {
	Name               string
	HTTPMethod         string
	Comment            string
	RequestTypeName    string
	RequestTypePackage string
	RequestTypeRawName string
	ReturnTypeName     string
	ReturnTypePackage  string
	ReturnTypeRawName  string
	Path               string
	Serializer         string
}
