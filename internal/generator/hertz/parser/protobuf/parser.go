package protobuf

import "github.com/aesoper101/codegen/internal/generator/hertz/parser"

var _ parser.Parser = (*Parser)(nil)

type Parser struct {
	opt *Options
}

func newParser(opt *Options) (*Parser, error) {
	return &Parser{opt: opt}, nil
}

func (p *Parser) Parse(files ...string) (parser.Packages, error) {
	//TODO implement me
	panic("implement me")
}
