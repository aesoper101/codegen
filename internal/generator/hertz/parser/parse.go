package parser

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/generator/hertz/config"
	"github.com/cloudwego/hertz/cmd/hz/generator"
	"github.com/cloudwego/hertz/cmd/hz/generator/model"
	"github.com/cloudwego/hertz/cmd/hz/util"
	"github.com/cloudwego/thriftgo/parser"
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

func (p *Parser) Parse() ([]*generator.HttpPackage, error) {
	args := p.args.Fork()
	idlPaths := args.IdlPaths

	var packages []*generator.HttpPackage
	for _, path := range idlPaths {
		args.IdlPaths = []string{path}
		pkg, err := p.parseFile(args)
		if err != nil {
			return nil, err
		}

		packages = append(packages, pkg)
	}
	return packages, nil
}

func (p *Parser) parseFile(args *config.Argument) (*generator.HttpPackage, error) {
	ast, err := parser.ParseFile(args.IdlPaths[0], p.args.Includes, true)
	if err != nil {
		return nil, err
	}
	// TODO 应该写个thriftgo插件
	pkgMap := args.OptPkgMap
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

	idlPackage := getGoPackage(ast, pkgMap)
	if idlPackage == "" {
		return nil, fmt.Errorf("go package for '%s' is not defined", ast.GetFilename())
	}

	services, err := astToService(ast, rs, args)
	if err != nil {
		return nil, err
	}
	//var models model.Models
	//for _, s := range services {
	//	models.MergeArray(s.Models)
	//}

	return &generator.HttpPackage{
		Services: services,
		IdlName:  ast.GetFilename(),
		Package:  idlPackage,
		//Models:   models,
	}, nil
}
