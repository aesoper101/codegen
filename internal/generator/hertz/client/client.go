package client

import "github.com/aesoper101/codegen/internal/generator"

var _ generator.Generator = (*Generator)(nil)

type Generator struct {
	idlPath string
	outPath string
	modName string
}

func (g *Generator) Name() string {
	return "HertzClient"
}

func (g *Generator) Generate() error {
	// TODO implement me
	panic("implement me")
}
