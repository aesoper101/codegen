package thrift

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/generator/kitex/parser"
	"github.com/aesoper101/codegen/pkg/utils"
	"github.com/aesoper101/x/copierx"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/generator/golang/streaming"
	parser2 "github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
	"path/filepath"
)

var defaultFeatures = golang.Features{
	MarshalEnumToText:           false,
	MarshalEnum:                 false,
	UnmarshalEnum:               false,
	GenerateSetter:              false,
	GenDatabaseTag:              false,
	GenOmitEmptyTag:             true,
	TypedefAsTypeAlias:          true,
	ValidateSet:                 true,
	ValueTypeForSIC:             false,
	ScanValueForEnum:            true,
	ReorderFields:               false,
	TypedEnumString:             false,
	KeepUnknownFields:           false,
	GenDeepEqual:                false,
	CompatibleNames:             false,
	ReserveComments:             true,
	NilSafe:                     false,
	FrugalTag:                   false,
	EscapeDoubleInTag:           true,
	GenerateTypeMeta:            false,
	GenerateJSONTag:             true,
	AlwaysGenerateJSONTag:       false,
	SnakeTyleJSONTag:            false,
	LowerCamelCaseJSONTag:       true,
	GenerateReflectionInfo:      false,
	WithReflection:              false,
	EnumAsINT32:                 false,
	CodeRefSlim:                 false,
	CodeRef:                     false,
	KeepCodeRefName:             false,
	TrimIDL:                     false,
	EnableNestedStruct:          true,
	JSONStringer:                false,
	WithFieldMask:               false,
	FieldMaskHalfway:            false,
	FieldMaskZeroRequired:       false,
	ThriftStreaming:             false,
	NoDefaultSerdes:             false,
	NoAliasTypeReflectionMethod: false,
	EnableRefInterface:          true,
	UseOption:                   false,
	NoFmt:                       false,
	SkipEmpty:                   false,
	NoProcessor:                 false,
}

var _ parser.Parser = (*Parser)(nil)

type Parser struct {
	opts         *Options
	cu           *golang.CodeUtils
	cache        *Cache
	packageCache map[string]*parser.Package // namespace -> package
}

func newParser(opts *Options) (*Parser, error) {
	cu := golang.NewCodeUtils(backend.DummyLogFunc())
	cu.SetFeatures(opts.Features)
	cu.SetPackagePrefix(opts.PackagePrefix)
	if len(opts.ImportReplace) > 0 {
		for path, repl := range opts.ImportReplace {
			cu.UsePackage(path, repl)
		}
	}
	return &Parser{
		opts:         opts,
		cu:           cu,
		cache:        newThriftIDLCache(cu),
		packageCache: make(map[string]*parser.Package),
	}, nil
}

func (p *Parser) Parse(files ...string) ([]*parser.Package, error) {
	files = utils.GetThriftFiles(files)
	for _, file := range files {
		if _, err := p.cache.AddFile(file); err != nil {
			return nil, err
		}
	}

	return p.parse()
}

func (p *Parser) parse() (pkgs []*parser.Package, err error) {
	scopes := p.cache.GetWaitProcessScopes()
	for _, scope := range scopes {
		p.cu.SetRootScope(scope)

		if _, err := p.makePackage(scope); err != nil {
			return nil, err
		}
	}

	for _, pkg := range p.packageCache {
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

func (p *Parser) makePackage(scope *golang.Scope) (*parser.Package, error) {
	pkg := p.findOrCreatePackage(scope)
	resolver := golang.NewResolver(scope, p.cu)

	file, err := p.makeFile(pkg, scope, resolver)
	if err != nil {
		return nil, err
	}

	pkg.Files = append(pkg.Files, file)
	pkg.Deps = mergeImports(pkg.Deps, file.Deps)
	return pkg, nil
}

func (p *Parser) findOrCreatePackage(scope *golang.Scope) *parser.Package {
	namespace := scope.AST().GetNamespaceOrReferenceName("go")
	if pkg, ok := p.packageCache[namespace]; ok {
		return pkg
	}
	pkg := &parser.Package{
		ImportPackage: scope.FilePackage(),
		ImportPath:    golang.GetImportPath(p.cu, scope.AST()),
		Namespace:     namespace,
		Deps:          make(map[string]string),
	}
	p.packageCache[namespace] = pkg
	return pkg
}

func (p *Parser) makeFile(pkg *parser.Package, scope *golang.Scope, resolver *golang.Resolver) (*parser.File, error) {
	file := &parser.File{
		Package:        pkg,
		IDLPath:        scope.AST().GetFilename(),
		Deps:           make(map[string]string),
		FileDescriptor: scope.AST(),
	}

	for _, service := range scope.Services() {
		svc, err := p.makeService(file, service, resolver)
		if err != nil {
			return nil, err
		}
		file.Services = append(file.Services, svc)
		file.Deps = mergeImports(file.Deps, svc.Deps)
	}

	return file, nil
}

func (p *Parser) makeService(file *parser.File, service *golang.Service, resolver *golang.Resolver) (*parser.Service, error) {
	ast := service.From().AST()

	si := &parser.Service{
		File:            file,
		ServiceFilePath: ast.GetFilename(),
		ServiceName:     service.GetName(),
		RawServiceName:  service.GetName(),
	}

	methods := service.Functions()
	extendMethods, err := p.getAllExtendFunction(service, resolver)
	if err != nil {
		return nil, err
	}

	allMethods := append(methods, extendMethods...)

	for _, method := range allMethods {
		mi, err := p.makeMethod(si, method, resolver)
		if err != nil {
			return nil, err
		}
		si.Methods = append(si.Methods, mi)
	}

	var annotations parser.Annotations
	_ = copierx.DeepCopy(service.GetAnnotations(), &annotations)

	deps, _ := service.From().ResolveImports()

	si.Deps = deps
	si.UseThriftReflection = p.cu.Features().WithReflection
	si.Annotations = annotations

	return si, nil
}

func (p *Parser) makeMethod(si *parser.Service, f *golang.Function, resolver *golang.Resolver) (*parser.Method, error) {
	st, err := streaming.ParseStreaming(f.Function)
	if err != nil {
		return nil, err
	}

	var annotations parser.Annotations
	_ = copierx.DeepCopy(f.GetAnnotations(), &annotations)

	mi := &parser.Method{
		Service:            si,
		Name:               f.GoName().String(),
		RawName:            f.Name,
		Oneway:             f.Oneway,
		Void:               f.Void,
		ArgStructName:      f.ArgType().GoName().String(),
		GenArgResultStruct: false,
		Streaming:          st,
		ClientStreaming:    st.ClientStreaming,
		ServerStreaming:    st.ServerStreaming,
		ArgsLength:         len(f.Arguments()),

		Annotations: annotations,
	}

	if st.IsStreaming {
		si.HasStreaming = true
	}

	deps := make(map[string]string)
	calcDeps := func(pkgInfo []parser.PkgInfo) {
		for _, dep := range pkgInfo {
			if _, ok := deps[dep.ImportPath]; !ok {
				deps[dep.ImportPath] = dep.Alias
			}
		}
	}

	scope := f.Service().From() // scope of the service, maybe not the root scope

	if !f.Oneway {
		mi.ResStructName = f.ResType().GoName().String()
	}
	if !f.Void {
		typeName := f.ResponseGoTypeName().String()
		mi.Resp = &parser.Parameter{
			Deps: p.getImports(resolver, scope, f.FunctionType),
			Type: typeName,
		}

		calcDeps(mi.Resp.Deps)

		mi.IsResponseNeedRedirect = "*"+typeName == f.ResType().Fields()[0].GoTypeName().String()
	}

	for _, a := range f.Arguments() {
		typeName, err := p.resolverTypeName(resolver, scope, a.GetType())
		if err != nil {
			return nil, err
		}
		arg := &parser.Parameter{
			Deps:    p.getImports(resolver, scope, a.GetType()),
			Name:    f.ArgType().Field(a.Name).GoName().String(),
			RawName: a.GoName().String(),
			Type:    typeName.String(),
		}
		mi.Args = append(mi.Args, arg)

		calcDeps(arg.Deps)
	}
	for _, t := range f.Throws() {
		typeName, err := p.resolverTypeName(resolver, scope, t.GetType())
		if err != nil {
			return nil, err
		}

		ex := &parser.Parameter{
			Deps:    p.getImports(resolver, scope, t.Type),
			Name:    f.ResType().Field(t.Name).GoName().String(),
			RawName: t.GoName().String(),
			Type:    typeName.String(),
		}
		mi.Exceptions = append(mi.Exceptions, ex)
		calcDeps(ex.Deps)
	}

	mi.Deps = deps
	return mi, nil
}

func (p *Parser) getImports(resolver *golang.Resolver, scope *golang.Scope, t *parser2.Type) (res []parser.PkgInfo) {
	switch t.Name {
	case "void":
		return nil
	case "bool", "byte", "i8", "i16", "i32", "i64", "double", "string", "binary":
		return nil
	case "map":
		res = append(res, p.getImports(resolver, scope, t.KeyType)...)
		fallthrough
	case "set", "list":
		res = append(res, p.getImports(resolver, scope, t.ValueType)...)
		return res
	default:
		typeName, err := p.resolverTypeName(resolver, scope, t)
		if err != nil {
			return nil
		}

		typName := typeName.Deref()
		if typName.IsForeign() {
			parts := semantic.SplitType(typName.String())
			if len(parts) == 2 {
				pth, alisa := getPthBy(p.cu.RootScope(), parts[0])
				if pth != "" {
					res = append(res, parser.PkgInfo{
						PkgName:    filepath.Base(pth),
						Alias:      alisa,
						ImportPath: pth,
					})
				}
			}
			return
		}
		if scope != p.cu.RootScope() {
			res = append(res, parser.PkgInfo{
				PkgName:    p.cu.GetPackageName(scope.AST()),
				ImportPath: golang.GetImportPath(p.cu, scope.AST()),
			})
		}

		return
	}
}

func (p *Parser) getAllExtendFunction(svc *golang.Service, resolver *golang.Resolver) (res []*golang.Function, err error) {
	if len(svc.Extends) == 0 {
		return
	}

	scope := svc.From() // scope of the service, maybe not the root scope
	ast := scope.AST()

	parts := semantic.SplitType(svc.Extends)

	switch len(parts) {
	case 1:
		// find the service in the same file
		extendSvc := scope.Service(parts[0])
		if extendSvc != nil {
			funcs := extendSvc.Functions()
			extendFuncs, err := p.getAllExtendFunction(extendSvc, resolver)
			if err != nil {
				return nil, err
			}
			res = append(res, append(funcs, extendFuncs...)...)
		}
	case 2:
		refAst, found := ast.GetReference(parts[0])
		if !found {
			return nil, fmt.Errorf("not found reference %s", parts[0])
		}

		refScope, found := p.cache.GetScopeByAst(refAst)
		if !found {
			return nil, fmt.Errorf("extend service %s not found", svc.Extends)
		}

		extendSvc := refScope.Service(parts[1])
		if extendSvc != nil {
			funcs := extendSvc.Functions()
			extendFuncs, err := p.getAllExtendFunction(extendSvc, resolver)
			if err != nil {
				return nil, err
			}
			res = append(res, append(funcs, extendFuncs...)...)
		}
		return res, nil
	}
	return res, nil
}

func (p *Parser) resolverTypeName(resolver *golang.Resolver, scope *golang.Scope, t *parser2.Type) (golang.TypeName, error) {
	typeName, err := resolver.GetTypeName(scope, t)
	if err != nil {
		return "", err
	}
	if t.Category.IsBaseType() {
		return typeName, nil
	}

	if t.Category.IsStructLike() {
		return typeName.Pointerize(), nil
	}

	return typeName, nil
}
