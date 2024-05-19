package hzclient

type CliGeneratorOption func(generator *Generator)

func WithBaseDomain(baseDomain string) CliGeneratorOption {
	return func(generator *Generator) {
		generator.baseDomain = baseDomain
	}
}

func WithModelMod(modelMod string) CliGeneratorOption {
	return func(generator *Generator) {
		generator.modelMod = modelMod
	}
}

func WithOutPath(outPath string) CliGeneratorOption {
	return func(generator *Generator) {
		generator.outPath = outPath
	}
}

func WithIdlPath(idlPath string) CliGeneratorOption {
	return func(generator *Generator) {
		generator.idlPath = idlPath
	}
}
