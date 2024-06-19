package v2

type Package struct {
	ImportPackage string
	ImportPath    string

	Namespace string

	Files []*File
}
