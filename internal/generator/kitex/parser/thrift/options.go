package thrift

import (
	"github.com/aesoper101/codegen/internal/generator/kitex/parser"
	"github.com/cloudwego/thriftgo/generator/golang"
)

var _ parser.Options = (*Options)(nil)

type Options struct {
	PackagePrefix string            `json:"package_prefix"`
	Features      golang.Features   `json:"features"`
	ImportReplace map[string]string `json:"import_replace"`
}

func DefaultOptions() Options {
	return Options{
		Features: defaultFeatures,
	}
}

func (o *Options) NewParser() (parser.Parser, error) {
	return newParser(o)
}
