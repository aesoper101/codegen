package parser

type HttpMethod struct {
	Name               string
	HTTPMethod         string
	Comment            string
	RequestArgName     string
	RequestTypeName    string
	RequestTypePackage string
	RequestTypeRawName string
	ReturnTypeName     string
	ReturnTypePackage  string
	ReturnTypeRawName  string
	Path               string
	Serializer         string

	Deps map[string]string // import path -> alias

	Annotations Annotations

	Service *Service `json:"-"`
}
