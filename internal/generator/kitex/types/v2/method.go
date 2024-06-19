package v2

import "github.com/cloudwego/thriftgo/generator/golang/streaming"

type Method struct {
	Service *Service

	Name                   string
	RawName                string
	Oneway                 bool
	Void                   bool
	Args                   []*Parameter
	ArgsLength             int
	Resp                   *Parameter
	Exceptions             []*Parameter
	ArgStructName          string
	ResStructName          string
	IsResponseNeedRedirect bool // int -> int*
	GenArgResultStruct     bool
	ClientStreaming        bool
	ServerStreaming        bool
	Streaming              *streaming.Streaming
}
