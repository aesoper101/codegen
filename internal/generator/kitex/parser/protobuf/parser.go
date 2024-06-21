package protobuf

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/generator/kitex/parser"
	"github.com/aesoper101/codegen/pkg/utils"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"strings"
)

var _ parser.Parser = (*Parser)(nil)

type Parser struct {
	opts         *Options
	packageCache map[string]*parser.Package
}

func newParser(opts *Options) (*Parser, error) {
	return &Parser{
		opts:         opts,
		packageCache: make(map[string]*parser.Package),
	}, nil
}

func (p *Parser) Parse(files ...string) ([]*parser.Package, error) {
	files = utils.GetProtoFiles(files)

	protoParser := protoparse.Parser{
		ImportPaths: []string{},
	}
	parseFiles, err := protoParser.ParseFiles(files...)
	if err != nil {
		return nil, err
	}

	for _, file := range parseFiles {
		if _, err := p.makePackage(file); err != nil {
			return nil, fmt.Errorf("failed to make package: %w", err)
		}
	}

	var packages []*parser.Package
	for _, pkg := range p.packageCache {
		packages = append(packages, pkg)
	}

	return packages, nil
}

func (p *Parser) makePackage(file *desc.FileDescriptor) (*parser.Package, error) {
	//protoOptions := file.GetOptions()
	return nil, nil
}

// https://www.cloudwego.io/zh/docs/kitex/tutorials/code-gen/code_generation/#%E4%BD%BF%E7%94%A8-protobuf-idl-%E7%9A%84%E6%B3%A8%E6%84%8F%E4%BA%8B%E9%A1%B9
func (p *Parser) findOrCreatePackage(file *desc.FileDescriptor) (*parser.Package, error) {
	namespace := file.GetPackage()
	namespace = strings.ReplaceAll(namespace, "/", ".")
	if pkg, ok := p.packageCache[namespace]; ok {
		return pkg, nil
	}

	pkg := &parser.Package{
		Namespace: namespace,
	}

	return pkg, nil
}
