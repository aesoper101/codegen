package parser

import (
	"errors"
	"fmt"
	"github.com/aesoper101/x/filex"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
)

type Converter struct {
	Utils    *golang.CodeUtils
	thrifts  map[string]*parser.Thrift
	svc2ast  map[*Service]*parser.Thrift
	services []*Service
}

func NewConverter() *Converter {
	cu := golang.NewCodeUtils(backend.DummyLogFunc())
	cu.SetFeatures(golang.Features{
		GenOmitEmptyTag:       true,
		TypedefAsTypeAlias:    true,
		ValidateSet:           true,
		GenerateJSONTag:       true,
		LowerCamelCaseJSONTag: true,
		ScanValueForEnum:      true,
		EscapeDoubleInTag:     true,
		ReserveComments:       true,
	})

	return &Converter{
		Utils:   cu,
		thrifts: make(map[string]*parser.Thrift),
		svc2ast: make(map[*Service]*parser.Thrift),
	}
}

func (c *Converter) Services() []*Service {
	return slices.Clone(c.services)
}

func (c *Converter) Convert(dirOrFilePaths ...string) error {
	var files []string
	for _, p := range dirOrFilePaths {
		if fs, err := c.getFiles(p); err != nil {
			return err
		} else {
			files = append(files, fs...)
		}
	}

	if err := c.parseFiles(files); err != nil {
		return err
	}

	if err := c.convertTypes(); err != nil {
		return err
	}

	return nil
}

func (c *Converter) convertTypes() error {
	for _, astThrift := range c.thrifts {
		for ast := range astThrift.DepthFirstSearch() {
			if err := semantic.ResolveSymbols(ast); err != nil {
				return fmt.Errorf("resolve fakse ast '%s': %w", ast.Filename, err)
			}

			scope, err := golang.BuildScope(c.Utils, ast)
			if err != nil {
				return fmt.Errorf("build scope for fake ast '%s': %w", ast.Filename, err)
			}

			c.Utils.SetRootScope(scope)
			for _, svc := range scope.Services() {
				if err := c.makeService(ast, svc); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *Converter) makeService(ast *parser.Thrift, srv *golang.Service) error {
	service := &Service{
		Name:    srv.GetName(),
		Comment: srv.GetReservedComments(),
	}

	baseDomainAnno := getAnnotation(srv.Annotations, ApiBaseDomain)
	if len(baseDomainAnno) == 1 {
		service.BaseDomain = baseDomainAnno[0]
	}

	groupAnno := getAnnotation(srv.Annotations, ApiServiceGroup)
	if len(groupAnno) == 1 {
		service.ServiceGroup = groupAnno[0]
	}

	pathAnno := getAnnotation(srv.Annotations, ApiServicePath)
	if len(pathAnno) == 1 {
		service.ServicePath = pathAnno[0]
	}

	for _, f := range srv.Functions() {
		method, err := c.makeMethod(service, f)
		if err != nil {
			return err
		}

		service.Methods = append(service.Methods, method)
	}

	for _, s := range c.services {
		if reflect.DeepEqual(s, service) {
			return nil
		}
	}

	c.services = append(c.services, service)
	c.svc2ast[service] = ast
	return nil
}

func (c *Converter) makeMethod(s *Service, f *golang.Function) (*HTTPMethod, error) {
	method := &HTTPMethod{
		Name:    f.GetName(),
		Comment: f.GetReservedComments(),
	}

	rs := getAnnotations(f.Annotations, HttpMethodAnnotations)

	httpAnnos := httpAnnotations{}
	for k, v := range rs {
		httpAnnos = append(httpAnnos, httpAnnotation{
			method: k,
			path:   v,
		})
	}

	sort.Sort(httpAnnos)

	if len(httpAnnos) > 0 {
		hmethod, path := httpAnnos[0].method, httpAnnos[0].path
		if len(path) == 0 || path[0] == "" {
			return nil, fmt.Errorf("invalid api.%s  for %s.%s: %s", hmethod, s.Name, method.Name, path)
		}
		method.Path = path[0]
		method.HTTPMethod = hmethod
	}
	return method, nil
}

func (c *Converter) parseFiles(files []string) error {
	c.thrifts = make(map[string]*parser.Thrift)
	for _, file := range files {
		ast, err := parser.ParseFile(file, []string{}, true)
		if err != nil {
			return err
		}

		c.thrifts[file] = ast
	}
	return nil
}

func (c *Converter) getFiles(dirOrFilePath string) ([]string, error) {
	if isThriftFile(dirOrFilePath) {
		return []string{dirOrFilePath}, nil
	}

	var files []string
	if filex.IsDir(dirOrFilePath) {
		err := filepath.Walk(dirOrFilePath, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && isThriftFile(path) {
				files = append(files, path)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	if len(files) == 0 {
		return nil, errors.New("no thrift file found ")
	}

	return files, nil
}

func isThriftFile(path string) bool {
	return filepath.Ext(path) == ".thrift"
}
