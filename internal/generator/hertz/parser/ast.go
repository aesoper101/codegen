package parser

import (
	"github.com/cloudwego/thriftgo/parser"
)

/*---------------------------Service-----------------------------*/
func astToService(ast *parser.Thrift) ([]*Service, error) {
	ss := ast.GetServices()
	out := make([]*Service, 0, len(ss))

	for _, s := range ss {
		service := &Service{
			Name: s.GetName(),
		}

		//funcs := s.GetFunctions()

		out = append(out, service)
	}

	return out, nil
}
