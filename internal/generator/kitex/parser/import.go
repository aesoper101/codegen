package parser

type PkgInfo struct {
	PkgName    string // PkgA
	Alias      string // A
	ImportPath string
}

// Parameter .
type Parameter struct {
	Deps    []PkgInfo
	Name    string
	RawName string
	Type    string // *PkgA.StructB
}
