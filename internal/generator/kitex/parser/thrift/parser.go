package thrift

import (
	"github.com/aesoper101/codegen/internal/generator/kitex/parser"
	"github.com/aesoper101/codegen/pkg/utils"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/generator/golang/streaming"
	"github.com/cloudwego/thriftgo/semantic"

	parser2 "github.com/cloudwego/thriftgo/parser"
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
	opts         options
	cu           *golang.CodeUtils
	cache        *Cache
	packageCache map[string]*parser.Package // namespace -> package
}

func NewParser(opts ...Option) (*Parser, error) {
	o := options{
		features: defaultFeatures,
	}
	if err := o.apply(opts...); err != nil {
		return nil, err
	}

	cu := golang.NewCodeUtils(backend.DummyLogFunc())
	cu.SetFeatures(o.features)
	cu.SetPackagePrefix(o.packagePrefix)
	if len(o.importReplace) > 0 {
		for path, repl := range o.importReplace {
			cu.UsePackage(path, repl)
		}
	}
	return &Parser{
		opts:         o,
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

		if pkg, err := p.makePackage(scope); err != nil {
			return nil, err
		} else {
			pkgs = append(pkgs, pkg)
		}
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
	}
	p.packageCache[namespace] = pkg
	return pkg
}

func (p *Parser) makeFile(pkg *parser.Package, scope *golang.Scope, resolver *golang.Resolver) (*parser.File, error) {
	file := &parser.File{
		Package: pkg,
		IDLPath: scope.AST().GetFilename(),
	}

	for _, service := range scope.Services() {
		svc, err := p.makeService(file, service, resolver)
		if err != nil {
			return nil, err
		}
		file.Services = append(file.Services, svc)
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

	si.UseThriftReflection = p.cu.Features().WithReflection

	return si, nil
}

func (p *Parser) makeMethod(si *parser.Service, f *golang.Function, resolver *golang.Resolver) (*parser.Method, error) {
	st, err := streaming.ParseStreaming(f.Function)
	if err != nil {
		return nil, err
	}

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
	}

	if st.IsStreaming {
		si.HasStreaming = true
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
	}
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
		typName := golang.TypeName(t.Name).Deref()
		if typName.IsForeign() {
			parts := semantic.SplitType(typName.String())
			if len(parts) == 2 {
				inc := scope.Includes().ByPackage(parts[0])
				if inc != nil {
					res = append(res, parser.PkgInfo{
						PkgRefName: inc.FilePackage(),
						ImportPath: golang.GetImportPath(p.cu, inc.AST()),
					})
				}
			}
			return
		}
		if scope != p.cu.RootScope() {
			res = append(res, parser.PkgInfo{
				PkgRefName: p.cu.GetPackageName(scope.AST()),
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
	parts := semantic.SplitType(svc.Extends)

	switch len(parts) {
	case 1:
		// find the service in the same file
		extendSvc := svc.From().Service(parts[0])
		if extendSvc != nil {
			funcs := extendSvc.Functions()
			extendFuncs, err := p.getAllExtendFunction(extendSvc, resolver)
			if err != nil {
				return nil, err
			}
			res = append(res, append(funcs, extendFuncs...)...)
		}
	case 2:
		inc := svc.From().Includes().ByPackage(parts[0])
		if inc != nil {
			extendSvc := inc.Service(parts[1])
			if extendSvc != nil {
				funcs := extendSvc.Functions()
				extendFuncs, err := p.getAllExtendFunction(extendSvc, resolver)
				if err != nil {
					return nil, err
				}
				res = append(res, append(funcs, extendFuncs...)...)
			}
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
	scope.Includes()
	if t.Category.IsStructLike() {
		return typeName.Pointerize(), nil
	}

	return typeName, nil
}
