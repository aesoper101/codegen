package v2

type ClientMethod struct {
	*HttpMethod
	BodyParamsCode   string
	QueryParamsCode  string
	PathParamsCode   string
	HeaderParamsCode string
	FormValueCode    string
	FormFileCode     string
}
