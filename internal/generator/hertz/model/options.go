package hzmodel

type ModelGeneratorOption func(*Generator)

func WithIDLPath(path string) ModelGeneratorOption {
	return func(g *Generator) {
		g.idlPath = path
	}
}

func WithOutPath(path string) ModelGeneratorOption {
	return func(g *Generator) {
		g.outPath = path
	}
}

func WithModName(name string) ModelGeneratorOption {
	return func(g *Generator) {
		g.modName = name
	}
}
