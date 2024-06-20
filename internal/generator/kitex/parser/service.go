package parser

type Service struct {
	File *File `json:"-"`

	ServiceName    string
	RawServiceName string
	//ServiceTypeName       func() string `json:"-"`
	Methods               []*Method
	HasStreaming          bool
	ServiceFilePath       string
	Protocol              string
	HandlerReturnKeepResp bool
	UseThriftReflection   bool

	Annotations Annotations

	Deps map[string]string // import path -> alias
}
