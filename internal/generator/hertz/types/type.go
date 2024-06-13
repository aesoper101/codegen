package types

import (
	"fmt"
)

type HttpPackage struct {
	PackageName string
	Package     string
	Namespace   string
	//RouterInfo *Router
	Files []*File
}

type PkgInfo struct {
	PkgName    string
	PkgRefName string
	ImportPath string
}

type File struct {
	IdlName     string
	FilePath    string
	GoFilePath  string
	PackageName string
	Package     string
	//Thrift      *parser.Thrift
	Services []*Service
}

type Service struct {
	Name        string
	PackageName string
	Package     string
	//RawService    *parser.Service
	Methods       []*HttpMethod
	ClientMethods []*ClientMethod
	BaseDomain    string // base domain for client code
	ServiceGroup  string // service level router group
	Imports       []PkgInfo
}

type HttpMethod struct {
	Name string
	//RawFunction        *parser.Function
	HTTPMethod         string
	Comment            string
	RequestArgName     string
	RequestTypeName    string
	RequestTypePackage string
	RequestTypeRawName string
	ReturnTypeName     string
	ReturnTypePackage  string
	ReturnTypeRawName  string
	Path               string
	Serializer         string
}

func (p *HttpPackage) AddFile(files ...*File) {
	var tempFiles []*File
	for _, file := range files {
		if p.checkHasSameFile(file, tempFiles) {
			continue
		}
		if p.checkServiceExists(file.Services) {
			continue
		}
		tempFiles = append(tempFiles, file)
	}

	p.Files = append(p.Files, tempFiles...)
}

func (p *HttpPackage) checkServiceExists(services []*Service) bool {
	for _, s := range p.Services() {
		for _, ss := range services {
			if s.Name == ss.Name {
				return true
			}
		}
	}
	return false
}

func (p *HttpPackage) checkHasSameFile(file *File, tempFiles []*File) bool {
	for _, f := range p.Files {
		if f.Equal(file) {
			return true
		}
	}
	return false
}

func (p *HttpPackage) Services() []*Service {
	var services []*Service
	for _, file := range p.Files {
		services = append(services, file.Services...)
	}
	return services
}

func (p *HttpPackage) IsSame(pp *HttpPackage) bool {
	if p.PackageName != pp.PackageName {
		return false
	}
	if len(p.Files) != len(pp.Files) {
		return false
	}
	for i, file := range p.Files {
		if !file.Equal(pp.Files[i]) {
			return false
		}
	}
	return true
}

func (p *HttpPackage) MergeSame(ps ...*HttpPackage) error {
	for _, pp := range ps {
		if !p.IsSame(pp) {
			return fmt.Errorf("package %s is not same with %s", p.PackageName, pp.PackageName)
		}
	}

	for _, pp := range ps {
		p.AddFile(pp.Files...)
	}
	return nil
}

func (f *File) Equal(file *File) bool {
	return f.IdlName == file.IdlName && f.FilePath == file.FilePath
}

func (f *File) AddService(services ...*Service) {
	for _, service := range services {
		if !f.checkServiceExists(service) {
			f.Services = append(f.Services, service)
		}
	}
}

func (f *File) checkServiceExists(service *Service) bool {
	for _, s := range f.Services {
		if s.Name == service.Name {
			return true
		}
	}
	return false
}

func (s *Service) AddImports(imports ...PkgInfo) {
	for _, imp := range imports {
		if imp.ImportPath != "" && !s.checkImportExists(imp) {
			s.Imports = append(s.Imports, imp)
		}
	}
}

func (s *Service) checkImportExists(imp PkgInfo) bool {
	for _, i := range s.Imports {
		if i.ImportPath == imp.ImportPath {
			return true
		}
	}
	return false
}
