package parser

type Service struct {
	Name    string
	Methods []*HTTPMethod
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
	RefPackage         string // handler import dir
	RefPackageAlias    string // handler import alias
}
