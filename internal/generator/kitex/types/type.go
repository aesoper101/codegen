package types

import "fmt"

type File struct {
	Name          string
	Content       string
	IsNew         bool
	BackupContent string
}

type PackageInfo struct {
	Namespace    string            // a dot-separated string for generating service package under kitex_gen
	Dependencies map[string]string // import path => alias
	Services     []*ServiceInfo    // the target service

	Imports map[string]map[string]bool // import path => alias

	IdlName string
	Package string
}

// AddServices .
func (p *PackageInfo) AddServices(ss ...*ServiceInfo) error {
	for _, s := range ss {
		if s == nil {
			continue
		}

		if p.checkServiceName(s.ServiceName) {
			return fmt.Errorf("duplicate service %s in package %s", s.ServiceName, p.Package)
		}

		s.from = p
		p.Services = append(p.Services, s)
	}

	p.calculateImports()
	p.calculateDependencies()
	return nil
}

func (p *PackageInfo) checkServiceName(name string) bool {
	for _, s := range p.Services {
		if s.ServiceName == name {
			return true
		}
	}
	return false
}

func (p *PackageInfo) calculateImports() {
	p.Imports = make(map[string]map[string]bool)
	for _, s := range p.Services {
		deps := s.Deps()
		for k, v := range deps {
			if _, ok := p.Imports[v]; !ok {
				p.Imports[v] = make(map[string]bool)
			}
			p.Imports[v][k] = true
		}
	}
}

func (p *PackageInfo) calculateDependencies() {
	p.Dependencies = make(map[string]string)
	for _, s := range p.Services {
		for k, v := range s.Deps() {
			p.Dependencies[v] = k
		}
	}
}

// PkgInfo .
type PkgInfo struct {
	PkgName    string
	PkgRefName string
	ImportPath string
}
