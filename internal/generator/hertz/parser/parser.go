package parser

import (
	"fmt"
	"github.com/aesoper101/x/filex"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
	"os"
	"path/filepath"
)

type Parser struct {
	cu *golang.CodeUtils
}

func New() *Parser {
	cu := golang.NewCodeUtils(backend.DummyLogFunc())
	return &Parser{cu: cu}
}

func (p *Parser) Parse(paths ...string) error {
	if err := p.parseFiles(paths...); err != nil {
		return err
	}
	p.cu.RootScope()
	return nil
}

func (p *Parser) parseFiles(paths ...string) error {
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if filex.IsExists(absPath) {
			return fmt.Errorf("file %s not exists", absPath)
		}

		if filex.IsDir(absPath) {
			err := filepath.Walk(absPath, func(pp string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				return p.parseFile(pp)
			})
			if err != nil {
				return err
			}
		}
		if err = p.parseFile(absPath); err != nil {
			return err
		}
	}

	return nil
}
func (p *Parser) parseFile(path string) error {
	ext := filepath.Ext(path)
	if ext != ".thrift" {
		return fmt.Errorf("file %s is not thrift file", path)
	}
	thrift, err := parser.ParseFile(path, []string{}, true)
	if err != nil {
		return err
	}
	_, err = golang.BuildScope(p.cu, thrift)
	if err != nil {
		return err
	}
	return nil
}
