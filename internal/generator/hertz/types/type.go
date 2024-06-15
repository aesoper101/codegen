package types

type HttpPackage struct {
	IdlName   string
	Package   string
	Namespace string
	Services  []*Service
}

func (p *HttpPackage) AddServices(s ...*Service) {
	for _, v := range s {
		v.from = p
		p.Services = append(p.Services, v)
	}
}

type PkgInfo struct {
	PkgName    string
	PkgRefName string
	ImportPath string
}

type Service struct {
	Name          string
	Methods       []*HttpMethod
	ClientMethods []*ClientMethod
	BaseDomain    string            // base domain for client code
	ServiceGroup  string            // service level router group
	Imports       map[string]string // import path -> alias

	from *HttpPackage
}

func (s *Service) AddHttpMethods(m ...*HttpMethod) {
	for _, v := range m {
		v.from = s
		s.Methods = append(s.Methods, v)
	}
}

func (s *Service) From() *HttpPackage {
	return s.from
}

type HttpMethod struct {
	Name               string
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

	from *Service
}

func (m *HttpMethod) GetRequestArgName() string {
	if m.RequestArgName != "" {
		return m.RequestArgName
	}
	return "req"
}

func (m *HttpMethod) From() *Service {
	return m.from
}
