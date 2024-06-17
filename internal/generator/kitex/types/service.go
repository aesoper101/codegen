package types

import (
	"fmt"
	"github.com/cloudwego/thriftgo/generator/golang/streaming"
)

// ServiceInfo .
type ServiceInfo struct {
	PkgInfo
	ServiceName           string
	RawServiceName        string
	ServiceTypeName       func() string `json:"-"`
	Methods               []*MethodInfo
	HasStreaming          bool
	ServiceFilePath       string
	Protocol              string
	HandlerReturnKeepResp bool
	UseThriftReflection   bool

	from *PackageInfo
}

// Deps .
func (s *ServiceInfo) Deps() map[string]string {
	deps := make(map[string]string)
	for _, m := range s.Methods {
		for _, a := range m.Args {
			for _, d := range a.Deps {
				deps[d.PkgName] = d.ImportPath
			}
		}

		if m.Resp != nil {
			for _, d := range m.Resp.Deps {
				deps[d.PkgName] = d.ImportPath
			}
		}

		for _, e := range m.Exceptions {
			for _, d := range e.Deps {
				deps[d.PkgName] = d.ImportPath
			}
		}
	}
	return deps
}

func (s *ServiceInfo) From() *PackageInfo {
	return s.from
}

func (s *ServiceInfo) AddMethods(methods ...*MethodInfo) error {
	for _, m := range methods {
		if m == nil {
			continue
		}
		if s.hasSameMethod(m) {
			return fmt.Errorf("duplicate method %s in service %s", m.Name, s.ServiceName)
		}
		m.from = s
		s.Methods = append(s.Methods, m)
	}

	return nil
}

func (s *ServiceInfo) hasSameMethod(m *MethodInfo) bool {
	for _, v := range s.Methods {
		if v.Name == m.Name {
			return true
		}
	}
	return false
}

// MethodInfo .
type MethodInfo struct {
	PkgInfo
	ServiceName            string
	Name                   string
	RawName                string
	Oneway                 bool
	Void                   bool
	Args                   []*Parameter
	ArgsLength             int
	Resp                   *Parameter
	Exceptions             []*Parameter
	ArgStructName          string
	ResStructName          string
	IsResponseNeedRedirect bool // int -> int*
	GenArgResultStruct     bool
	ClientStreaming        bool
	ServerStreaming        bool
	Streaming              *streaming.Streaming

	from *ServiceInfo
}

func (m *MethodInfo) Service() *ServiceInfo {
	return m.from
}

// Parameter .
type Parameter struct {
	Deps    []PkgInfo
	Name    string
	RawName string
	Type    string // *PkgA.StructB
}
