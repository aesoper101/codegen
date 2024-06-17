package utils

import (
	"cmp"
	"fmt"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
	"log/slog"
)

type ThriftIDLCache struct {
	// cache for thrift file. key is file path, value is thrift object
	cache map[string]*golang.Scope

	waitProcessScopes map[string]*golang.Scope

	cu      *golang.CodeUtils
	checker semantic.Checker
}

func NewThriftIDLCache(cu *golang.CodeUtils) *ThriftIDLCache {
	return &ThriftIDLCache{
		cache:             make(map[string]*golang.Scope),
		cu:                cu,
		checker:           semantic.NewChecker(semantic.Options{FixWarnings: true}),
		waitProcessScopes: make(map[string]*golang.Scope),
	}
}

func (c *ThriftIDLCache) GetScope(path string) (*golang.Scope, bool) {
	t, ok := c.cache[path]
	return t, ok
}

func (c *ThriftIDLCache) GetScopeByAst(ast *parser.Thrift) (*golang.Scope, bool) {
	path := ast.GetFilename()
	return c.GetScope(path)
}

func (c *ThriftIDLCache) AddFile(filename string) (*golang.Scope, error) {
	ast, err := parser.ParseFile(filename, nil, true)
	if err != nil {
		return nil, err
	}

	if path := parser.CircleDetect(ast); len(path) > 0 {
		return nil, fmt.Errorf("found include circle:\n\t%s", path)
	}

	if warns := parser.DetectKeyword(ast); len(warns) > 0 {
		slog.Warn("detect keyword", slog.Any("warns", warns))
	}

	if scope, ok := c.waitProcessScopes[ast.GetFilename()]; ok {
		slog.Warn("already in process, skip.", slog.String("path", ast.GetFilename()))
		return scope, nil
	}

	if scope, ok := c.cache[ast.GetFilename()]; ok {
		c.waitProcessScopes[ast.GetFilename()] = scope
		return scope, nil
	}

	scope, err := c.deepAdd(ast)
	if err != nil {
		return nil, err
	}

	c.waitProcessScopes[ast.GetFilename()] = scope
	return scope, nil
}

func (c *ThriftIDLCache) deepAdd(astFile *parser.Thrift) (*golang.Scope, error) {
	for ast := range astFile.DepthFirstSearch() {
		if _, err := c.add(ast); err != nil {
			return nil, err
		}
	}

	path := astFile.GetFilename()
	return c.cache[path], nil
}

func (c *ThriftIDLCache) add(ast *parser.Thrift) (*golang.Scope, error) {
	path := ast.GetFilename()
	if scope, ok := c.cache[path]; ok {
		return scope, nil
	}

	if warns, err := c.checker.CheckAll(ast); err != nil {
		slog.Warn("semantic check failed", slog.Any("warns", warns), slog.String("path", path))
		return nil, err
	}

	if err := semantic.ResolveSymbols(ast); err != nil {
		return nil, err
	}

	s1, s2, err := golang.BuildRefScope(c.cu, ast)
	if err != nil {
		return nil, err
	}

	scope := cmp.Or(s1, s2)

	c.cache[path] = scope
	return scope, nil
}

func (c *ThriftIDLCache) GetWaitProcessScopes() []*golang.Scope {
	var scopes []*golang.Scope
	for _, scope := range c.waitProcessScopes {
		scopes = append(scopes, scope)
	}
	return scopes
}
