package thrift

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/generator/hertz/consts"
	"github.com/aesoper101/codegen/internal/generator/hertz/parser"
	"github.com/aesoper101/codegen/pkg/utils"
	"github.com/aesoper101/x/copierx"
	"github.com/aesoper101/x/str"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	parser2 "github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
	"log/slog"
	"sort"
	"strings"
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
	cu           *golang.CodeUtils
	opts         options
	cache        *IDLCache
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

		if _, err := p.makePackage(scope); err != nil {
			return nil, err
		}
	}

	for _, pkg := range p.packageCache {
		pkgs = append(pkgs, pkg)
	}
	return
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
	ast := scope.AST()
	namespace := ast.GetNamespaceOrReferenceName("go")
	if pkg, ok := p.packageCache[namespace]; ok {
		return pkg
	}

	pkg := &parser.Package{
		ImportPackage: scope.FilePackage(),
		ImportPath:    golang.GetImportPath(p.cu, ast),
		Namespace:     namespace,
	}
	p.packageCache[namespace] = pkg
	return pkg
}

func (p *Parser) makeFile(pkg *parser.Package, scope *golang.Scope, resolver *golang.Resolver) (*parser.File, error) {
	file := &parser.File{
		Package: pkg,
		IDLPath: scope.AST().GetFilename(),
		Deps:    make(map[string]string),
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
	deps, _ := service.From().ResolveImports()
	si := &parser.Service{
		File:          file,
		Name:          service.GetName(),
		Comment:       service.GetReservedComments(),
		Methods:       nil,
		ClientMethods: nil,
		BaseDomain:    "",
		ServiceGroup:  "",
		Deps:          deps,
	}

	baseDomainAnnos := utils.GetAnnotation(service.Annotations, consts.ApiBaseDomain)
	if len(baseDomainAnnos) > 0 {
		si.BaseDomain = baseDomainAnnos[0]
	}
	serviceGroupAnnos := utils.GetAnnotation(service.Annotations, consts.ApiServiceGroup)
	if len(serviceGroupAnnos) > 0 {
		si.ServiceGroup = serviceGroupAnnos[0]
	}

	ms := service.Functions()

	extendSvcs, err := p.getAllExtendFunction(service, resolver)
	if err != nil {
		return nil, err
	}

	ms = append(ms, extendSvcs...)

	httpMethods := make([]*parser.HttpMethod, 0)
	clientMethods := make([]*parser.ClientMethod, 0)
	for _, m := range ms {
		httpMethod, err := p.makeHttpMethod(si, m, resolver)
		if err != nil {
			return nil, err
		}

		if err := checkHttpMethodSamePath(httpMethods, httpMethod); err != nil {
			return nil, err
		}
		httpMethods = append(httpMethods, httpMethod)

		clientMethod := &parser.ClientMethod{
			HttpMethod: httpMethod,
		}
		if err := p.parseAnnotationToClient(clientMethod, m); err != nil {
			return nil, err
		}
		clientMethods = append(clientMethods, clientMethod)
	}

	var annotations parser.Annotations
	_ = copierx.DeepCopy(service.GetAnnotations(), &annotations)

	si.Methods = httpMethods
	si.ClientMethods = clientMethods
	si.Annotations = annotations
	return si, nil
}

func (p *Parser) makeHttpMethod(si *parser.Service, m *golang.Function, resolver *golang.Resolver) (*parser.HttpMethod, error) {
	sr, _ := utils.GetFirstKV(utils.GetAnnotations(m.GetAnnotations(), consts.SerializerTags))
	rs := utils.GetAnnotations(m.GetAnnotations(), consts.HttpMethodAnnotations)
	if len(rs) == 0 {
		return nil, fmt.Errorf("no http method annotations found for method %s.%s", m.Service().GetName(), m.GetName())
	}
	httpAnnos := httpAnnotations{}
	for k, v := range rs {
		httpAnnos = append(httpAnnos, httpAnnotation{
			method: k,
			path:   v,
		})
	}

	sort.Sort(httpAnnos)

	hmethod, path := httpAnnos[0].method, httpAnnos[0].path
	if len(path) == 0 || path[0] == "" {
		return nil, fmt.Errorf("invalid api.%s  for %s.%s: %s", hmethod, m.Service().GetName(), m.GetName(), path)
	}

	deps := make(map[string]string)
	scope := m.Service().From() // scope of the service, maybe not the root scope

	var reqName, reqRawName, reqPackage, reqArgName string
	if len(m.GetArguments()) >= 1 {
		reqArgName = m.GetArguments()[0].GetName()
		if len(m.GetArguments()) > 1 {
			slog.Warn(fmt.Sprintf("function '%s' has more than one argument, but only the first can be used in hertz now", m.GetName()))
		}

		reqName = m.Arguments()[0].GetType().GetName()

		//tyName, err := resolver.GetTypeName(m.Service().From(), m.Arguments()[0].GetType())
		tyName, err := p.resolverTypeName(resolver, scope, m.GetArguments()[0].GetType())
		if err != nil {
			return nil, err
		}
		reqName = tyName.Deref().String()

		if strings.Contains(reqName, ".") && !m.GetArguments()[0].GetType().Category.IsContainerType() {
			// If reqName contains "." , then it must be of the form "pkg.name".
			// so reqRawName='name', reqPackage='pkg'
			names := strings.Split(reqName, ".")
			if len(names) != 2 {
				return nil, fmt.Errorf("request name: %s is wrong", reqName)
			}
			reqRawName = names[1]
			reqPackage = names[0]
			reqPth, alisa := getPthBy(p.cu.RootScope(), reqPackage)
			if reqPth != "" {
				deps[reqPth] = alisa
			}
		}
	}
	var respName, respRawName, respPackage string
	if !m.Oneway {
		//tyName, err := resolver.GetTypeName(m.Service().From(), m.GetFunctionType())
		tyName, err := p.resolverTypeName(resolver, scope, m.GetFunctionType())
		if err != nil {
			return nil, err
		}

		respName = tyName.Deref().String()

		if strings.Contains(respName, ".") && !m.GetFunctionType().Category.IsContainerType() {
			names := strings.Split(respName, ".")
			if len(names) != 2 {
				return nil, fmt.Errorf("response name: %s is wrong", respName)
			}
			// If respName contains "." , then it must be of the form "pkg.name".
			// so respRawName='name', respPackage='pkg'
			respRawName = names[1]
			respPackage = names[0]
			respPth, alisa := getPthBy(p.cu.RootScope(), respPackage)
			if respPth != "" {
				deps[respPth] = alisa
			}
		}
	}

	var annotations parser.Annotations
	_ = copierx.DeepCopy(m.GetAnnotations(), &annotations)

	return &parser.HttpMethod{
		Service:            si,
		Name:               m.GetName(),
		Comment:            m.GetReservedComments(),
		HTTPMethod:         hmethod,
		RequestArgName:     reqArgName,
		RequestTypeName:    reqName,
		RequestTypeRawName: reqRawName,
		RequestTypePackage: reqPackage,
		ReturnTypeName:     respName,
		ReturnTypeRawName:  respRawName,
		ReturnTypePackage:  respPackage,
		Path:               path[0],
		Serializer:         sr,
		Deps:               deps,
		Annotations:        annotations,
	}, nil
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
		fmt.Println(p.cu.RootScope().ResolveImports())
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

func (p *Parser) parseAnnotationToClient(clientMethod *parser.ClientMethod, m *golang.Function) error {
	var (
		hasBodyAnnotation bool
		hasFormAnnotation bool
	)
	for _, field := range m.Arguments() {
		hasAnnotation := false
		isStringFieldType := false
		if field.GetType().String() == "string" {
			isStringFieldType = true
		}
		if anno := utils.GetAnnotation(field.Annotations, consts.AnnotationQuery); len(anno) > 0 {
			hasAnnotation = true
			//query := str.ToSnake(anno[0])
			query := anno[0]
			clientMethod.QueryParamsCode += fmt.Sprintf("%q: req.Get%s(),\n", query, str.ToCamel(field.GoName().String()))
		}

		if anno := utils.GetAnnotation(field.Annotations, consts.AnnotationPath); len(anno) > 0 {
			hasAnnotation = true
			path := anno[0]
			if isStringFieldType {
				clientMethod.PathParamsCode += fmt.Sprintf("%q: req.Get%s(),\n", path, field.GoName().String())
			} else {
				clientMethod.PathParamsCode += fmt.Sprintf("%q: fmt.Sprint(req.Get%s()),\n", path, field.GoName().String())
			}
		}

		if anno := utils.GetAnnotation(field.Annotations, consts.AnnotationHeader); len(anno) > 0 {
			hasAnnotation = true
			header := anno[0]
			if isStringFieldType {
				clientMethod.HeaderParamsCode += fmt.Sprintf("%q: req.Get%s(),\n", header, field.GoName().String())
			} else {
				clientMethod.HeaderParamsCode += fmt.Sprintf("%q: fmt.Sprint(req.Get%s()),\n", header, field.GoName().String())
			}
		}

		if anno := utils.GetAnnotation(field.Annotations, consts.AnnotationForm); len(anno) > 0 {
			hasAnnotation = true
			//form := str.ToSnake(anno[0])
			form := anno[0]
			hasFormAnnotation = true
			if isStringFieldType {
				clientMethod.FormValueCode += fmt.Sprintf("%q: req.Get%s(),\n", form, field.GoName().String())
			} else {
				clientMethod.FormValueCode += fmt.Sprintf("%q: fmt.Sprint(req.Get%s()),\n", form, field.GoName().String())
			}
		}

		if anno := utils.GetAnnotation(field.Annotations, consts.AnnotationBody); len(anno) > 0 {
			hasAnnotation = true
			hasBodyAnnotation = true
		}

		if anno := utils.GetAnnotation(field.Annotations, consts.AnnotationFileName); len(anno) > 0 {
			hasAnnotation = true
			fileName := anno[0]
			hasFormAnnotation = true

			clientMethod.FormFileCode += fmt.Sprintf("%q: req.Get%s(),\n", fileName, str.
				ToCamel(field.GoName().String()))
		}
		if !hasAnnotation && strings.EqualFold(clientMethod.HTTPMethod, "get") {
			clientMethod.QueryParamsCode += fmt.Sprintf("%q: req.Get%s(),\n", field.GetName(), str.ToCamel(field.GoName().String()))
		}
	}
	clientMethod.BodyParamsCode = consts.SetBodyParam
	if hasBodyAnnotation && hasFormAnnotation {
		clientMethod.FormValueCode = ""
		clientMethod.FormFileCode = ""
	}
	if !hasBodyAnnotation && hasFormAnnotation {
		clientMethod.BodyParamsCode = ""
	}

	return nil
}

type httpAnnotation struct {
	method string
	path   []string
}

type httpAnnotations []httpAnnotation

func (s httpAnnotations) Len() int {
	return len(s)
}

func (s httpAnnotations) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s httpAnnotations) Less(i, j int) bool {
	return s[i].method < s[j].method
}
