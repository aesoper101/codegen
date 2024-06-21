package parser

type Package struct {
	ImportPackage string
	ImportPath    string

	Namespace string

	Files []*File

	Deps map[string]string // import path -> alias
}

type Packages []*Package

func (p Packages) Files() []*File {
	var files []*File
	for _, pkg := range p {
		files = append(files, pkg.Files...)
	}
	return files
}

func (p Packages) Services() []*Service {
	var services []*Service
	for _, file := range p.Files() {
		services = append(services, file.Services...)
	}
	return services
}

func (p Packages) Methods() []*Method {
	var methods []*Method
	for _, si := range p.Services() {
		methods = append(methods, si.Methods...)
	}
	return methods
}
