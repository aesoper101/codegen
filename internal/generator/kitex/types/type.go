package types

type File struct {
	Name          string
	Content       string
	IsNew         bool
	BackupContent string
}

type PackageInfo struct {
	Namespace    string            // a dot-separated string for generating service package under kitex_gen
	Dependencies map[string]string // package name => import path, used for searching imports
	*ServiceInfo                   // the target service

	// the following fields will be filled and used by the generator
	Codec            string
	NoFastAPI        bool
	Version          string
	RealServiceName  string
	Imports          map[string]map[string]bool // import path => alias
	ExternalKitexGen string
	Features         []feature
	FrugalPretouch   bool
	Module           string
	Protocol         Protocol
	IDLName          string
	ServerPkg        string
}

// AddImport .
func (p *PackageInfo) AddImport(pkg, path string) {
	if p.Imports == nil {
		p.Imports = make(map[string]map[string]bool)
	}
	if pkg != "" {
		//if p.ExternalKitexGen != "" && strings.Contains(path, KitexGenPath) {
		//	parts := strings.Split(path, KitexGenPath)
		//	path = filepathx.JoinPath(p.ExternalKitexGen, parts[len(parts)-1])
		//}
		if path == pkg {
			p.Imports[path] = nil
		} else {
			if p.Imports[path] == nil {
				p.Imports[path] = make(map[string]bool)
			}
			p.Imports[path][pkg] = true
		}
	}
}

// AddImports .
func (p *PackageInfo) AddImports(pkgs ...string) {
	for _, pkg := range pkgs {
		if path, ok := p.Dependencies[pkg]; ok {
			p.AddImport(pkg, path)
		} else {
			p.AddImport(pkg, pkg)
		}
	}
}

// PkgInfo .
type PkgInfo struct {
	PkgName    string
	PkgRefName string
	ImportPath string
}
