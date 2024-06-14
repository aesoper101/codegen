package types

type ClientMethod struct {
	*HttpMethod
	BodyParamsCode   string
	QueryParamsCode  string
	PathParamsCode   string
	HeaderParamsCode string
	FormValueCode    string
	FormFileCode     string
}

type ClientConfig struct {
	QueryEnumAsInt bool
}

type ClientFile struct {
	Config        ClientConfig
	FilePath      string
	PackageName   string
	ServiceName   string
	BaseDomain    string
	Imports       map[string]string
	ClientMethods []*ClientMethod
}
