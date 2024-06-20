package parser

type HttpMethod struct {
	Name               string
	HTTPMethod         string
	Comment            string
	RequestArgName     string
	RequestTypeName    string
	RequestTypePackage string
	RequestTypeRawName string
	RequestDeps        []PkgInfo
	ReturnTypeName     string
	ReturnTypePackage  string
	ReturnTypeRawName  string
	ReturnDeps         []PkgInfo
	Path               string
	Serializer         string

	Service *Service `json:"-"`
}
