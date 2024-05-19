package parser

import (
	"github.com/aesoper101/x/filex"
	"github.com/cloudwego/thriftgo/parser"
	"io/fs"
	"path/filepath"
)

type Parser struct {
	idlPath string
}

type Option func(parser *Parser) error

func New(opts ...Option) (*Parser, error) {
	p := &Parser{idlPath: "idl"}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (p *Parser) Parse() ([]*Service, error) {
	files, err := p.getAllThriftFile()
	if err != nil {
		return nil, err
	}

	var out []*Service
	for _, file := range files {
		service, err := p.parseFile(file)
		if err != nil {
			return nil, err
		}

		out = append(out, service...)
	}

	return out, err
}

func (p *Parser) parseFile(path string) ([]*Service, error) {
	astThrift, err := parser.ParseFile(path, []string{}, true)
	if err != nil {
		return nil, err
	}
	return astToService(astThrift)
}

func (p *Parser) getAllThriftFile() ([]string, error) {
	var paths []string
	if filex.IsDir(p.idlPath) {
		err := filepath.WalkDir(p.idlPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				if ext := filepath.Ext(path); ext == ".thrift" {
					paths = append(paths, path)
				}
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return paths, nil
}
