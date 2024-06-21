package protobuf

import "github.com/aesoper101/codegen/internal/generator/hertz/parser"

var _ parser.Options = (*Options)(nil)

type Options struct {
}

func DefaultOptions() Options {
	return Options{}
}

func (o *Options) NewParser() (parser.Parser, error) {
	return newParser(o)
}
