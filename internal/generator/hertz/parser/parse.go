package parser

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/generator/hertz/config"
	"github.com/cloudwego/hertz/cmd/hz/generator"
	"github.com/cloudwego/hertz/cmd/hz/generator/model"
	"github.com/cloudwego/hertz/cmd/hz/util"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/plugin"
	"io"
	"os"
)

type Parser struct {
	args *config.Argument
}

type Option func(parser *Parser) error

func WithArgument(args *config.Argument) Option {
	return func(parser *Parser) error {
		parser.args = args
		return nil
	}
}

func New(opts ...Option) (*Parser, error) {
	p := &Parser{
		args: config.NewArgument(),
	}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (p *Parser) Parse(args *config.Argument) ([]*generator.Service, error) {
	return p.parseFile(args)
}

func (p *Parser) parseFile(args *config.Argument) ([]*generator.Service, error) {
	req, err := p.handleRequest()
	if err != nil {
		return nil, err
	}
	thriftgoUtil = golang.NewCodeUtils(backend.DummyLogFunc())
	if err := thriftgoUtil.HandleOptions(req.GeneratorParameters); err != nil {
		return nil, err
	}

	ast := req.GetAST()

	pkgMap := args.OptPkgMap
	pkg := getGoPackage(ast, pkgMap)
	main := &model.Model{
		FilePath:    ast.Filename,
		Package:     pkg,
		PackageName: util.SplitPackageName(pkg, ""),
	}
	fmt.Printf("%+v\n", main)
	rs, err := NewResolver(ast, main, pkgMap)
	if err != nil {
		return nil, err
	}
	return astToService(ast, rs, args)
}

func (p *Parser) handleRequest() (*plugin.Request, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	return plugin.UnmarshalRequest(data)
}
