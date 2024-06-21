package parser

type Service struct {
	File *File `json:"-"`

	Name          string
	Comment       string
	Methods       []*HttpMethod
	ClientMethods []*ClientMethod
	BaseDomain    string // base domain for client code
	ServiceGroup  string // service level router group

	Deps map[string]string // import path -> alias

	Annotations Annotations
}

func (s *Service) GetMethod(name string) *HttpMethod {
	for _, m := range s.Methods {
		if m.Name == name {
			return m
		}
	}
	return nil
}

func (s *Service) GetClientMethod(name string) *ClientMethod {
	for _, m := range s.ClientMethods {
		if m.Name == name {
			return m
		}
	}
	return nil
}
