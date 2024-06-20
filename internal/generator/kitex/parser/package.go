package parser

type Package struct {
	ImportPackage string
	ImportPath    string

	Namespace string

	Files []*File
}

func (p *Package) Clone() *Package {
	return &Package{
		ImportPackage: p.ImportPackage,
		ImportPath:    p.ImportPath,
		Namespace:     p.Namespace,
		Files:         p.Files,
	}
}
