package parser

type Package struct {
	ImportPackage string
	ImportPath    string

	Namespace string

	Files []*File
}
