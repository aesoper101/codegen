package parser

type Package struct {
	ImportPackage string
	ImportPath    string

	Namespace string

	Files []*File

	Deps map[string]string // import path -> alias
}
