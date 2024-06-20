package thrift

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/generator/hertz/parser"
)

func checkHttpMethodSamePath(ms []*parser.HttpMethod, m *parser.HttpMethod) error {
	for _, method := range ms {
		if method.Path == m.Path {
			return fmt.Errorf("http method %s has same path %s with method %s", m.Name, m.Path, method.Name)
		}
	}
	return nil
}
