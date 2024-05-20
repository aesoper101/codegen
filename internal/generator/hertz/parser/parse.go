package parser

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/generator/hertz/config"
	"github.com/cloudwego/hertz/cmd/hz/generator"
	"github.com/cloudwego/hertz/cmd/hz/generator/model"
	"github.com/cloudwego/hertz/cmd/hz/util"
	"github.com/cloudwego/hertz/cmd/hz/util/logs"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
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

func (p *Parser) Parse() ([]*generator.Service, error) {
	return p.parseFile()
}

func (p *Parser) parseFile() ([]*generator.Service, error) {
	//req, err := p.handleRequest()
	//if err != nil {
	//	return nil, err
	//}
	//
	//args, err := p.parseArgs(req)
	//if err != nil {
	//	return nil, err
	//}

	//ast := req.GetAST()
	//if ast == nil {
	//	return nil, fmt.Errorf("no ast")
	//}

	ast, err := parser.ParseFile(p.args.IdlPaths[0], p.args.Includes, true)
	if err != nil {
		return nil, err
	}
	args := p.args
	//modelDir,_ := args.GetModelDir()
	//req := &plugin.Request{
	//	Version:    version.ThriftgoVersion,
	//	OutputPath: modelDir,
	//	Recursive:  !args.NoRecurse,
	//	AST:        ast,
	//}

	// TODO 应该写个thriftgo插件
	pkgMap := p.args.OptPkgMap
	pkg := getGoPackage(ast, pkgMap)
	main := &model.Model{
		FilePath:    ast.Filename,
		Package:     pkg,
		PackageName: util.SplitPackageName(pkg, ""),
	}

	rs, err := NewResolver(ast, main, pkgMap)
	if err != nil {
		return nil, err
	}
	err = rs.LoadAll(ast)
	if err != nil {
		return nil, err
	}
	return astToService(ast, rs, args)
}

func (p *Parser) handleRequest() (*plugin.Request, error) {
	// TODO 这里是读不到数据的关键
	//data, err := io.ReadAll(os.Stdin)
	//if err != nil {
	//	return nil, err
	//}
	//data := []byte(`{"Version":"0.0.1","GeneratorParameters":["--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports"],"PluginParameters":["--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports","--thrift_imports","github.com/cloudwego/thriftgo/thrift_imports"],"Language":"go","OutputPath":"./","Recursive":true}`)
	//req, err := plugin.UnmarshalRequest(data)
	//if err != nil {
	//	return nil, err
	//}
	req := plugin.NewRequest()
	thriftgoUtil = golang.NewCodeUtils(backend.DummyLogFunc())
	if err := thriftgoUtil.HandleOptions(req.GeneratorParameters); err != nil {
		return nil, err
	}
	return req, nil
}

func (p *Parser) parseArgs(req *plugin.Request) (*config.Argument, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	args := new(config.Argument)
	err := args.Unpack(req.PluginParameters)
	if err != nil {
		logs.Errorf("unpack args failed: %s", err.Error())
	}
	p.args = args
	return args, nil
}
