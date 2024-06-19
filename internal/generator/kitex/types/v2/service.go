package v2

type Service struct {
	File *File

	ServiceName           string
	RawServiceName        string
	ServiceTypeName       func() string `json:"-"`
	Methods               []*Method
	HasStreaming          bool
	ServiceFilePath       string
	Protocol              string
	HandlerReturnKeepResp bool
	UseThriftReflection   bool
}
