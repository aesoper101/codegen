package thrift

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/generator/hertz/consts"
	"github.com/aesoper101/codegen/internal/generator/hertz/types"
	"github.com/aesoper101/codegen/pkg/utils"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
	"log/slog"
	"sort"
	"strings"
)

var _ types.Parser = (*Parser)(nil)

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

type Parser struct {
	packages    []*types.HttpPackage
	thriftFiles []*parser.Thrift
	all         map[string][]*parser.Thrift
	cu          *golang.CodeUtils
}

func NewParser() *Parser {
	cu := golang.NewCodeUtils(backend.DummyLogFunc())
	cu.SetFeatures(defaultFeatures)
	cu.SetPackagePrefix("github.com/aesoper101/xxx")
	return &Parser{
		cu: cu,
	}
}

func (p *Parser) Parse(files ...string) ([]*types.HttpPackage, error) {
	fs := utils.GetThriftFiles(files)
	if err := p.parseFiles(fs); err != nil {
		return nil, err
	}

	if err := p.makePackages(); err != nil {
		return nil, err
	}

	return p.packages, nil
}

func (p *Parser) parseFiles(files []string) error {
	for _, f := range files {
		ast, err := parser.ParseFile(f, []string{}, true)
		if err != nil {
			return err
		}

		checker := semantic.NewChecker(semantic.Options{FixWarnings: true})
		if warns, err := checker.CheckAll(ast); err != nil {
			slog.Warn(fmt.Sprintf("check all failed: %v", err), "warns", warns)
			return err
		}

		p.thriftFiles = append(p.thriftFiles, ast)
	}

	for _, ast := range p.thriftFiles {
		for ast1 := range ast.DepthFirstSearch() {
			if err := semantic.ResolveSymbols(ast1); err != nil {
				return err
			}

			_, _, err := golang.BuildRefScope(p.cu, ast1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Parser) makePackages() error {
	for _, ast := range p.thriftFiles {
		scope, err := golang.BuildScope(p.cu, ast)
		if err != nil {
			return err
		}
		p.cu.SetRootScope(scope)
		pkg, err := p.makePackage(ast)
		if err != nil {
			return err
		}

		if err := p.checkAndMerge(pkg); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) checkAndMerge(pkg *types.HttpPackage) error {
	same := false
	for _, p := range p.packages {
		if p.IsSame(pkg) {
			same = true
			if err := p.MergeSame(pkg); err != nil {
				return err
			}
		}
	}
	if !same {
		p.packages = append(p.packages, pkg)
	}

	return nil
}

func (p *Parser) findOrCreatePackage(ast *parser.Thrift) *types.HttpPackage {
	ns := ast.GetNamespaceOrReferenceName("go")
	packageName := p.cu.GetPackageName(ast)
	pth := p.cu.NamespaceToFullImportPath(ns)

	for _, pkg := range p.packages {
		if pkg.Namespace == ns && pkg.Package == pth {
			return pkg
		}
	}

	return &types.HttpPackage{
		PackageName: packageName,
		Package:     pth,
		Namespace:   ns,
	}
}

func (p *Parser) makePackage(ast *parser.Thrift) (*types.HttpPackage, error) {
	httpPackage := p.findOrCreatePackage(ast)
	file := &types.File{
		IdlName:     p.cu.RootScope().IDLName(),
		FilePath:    ast.GetFilename(),
		GoFilePath:  p.cu.GetFilePath(ast),
		Package:     httpPackage.Package,
		PackageName: httpPackage.PackageName,
		Services:    nil,
	}

	for _, s := range ast.Services {
		service, err := p.makeService(file, s, ast)
		if err != nil {
			return nil, err
		}
		file.AddService(service)
	}

	httpPackage.AddFile(file)
	return httpPackage, nil
}

func (p *Parser) makeService(file *types.File, s *parser.Service, ast *parser.Thrift) (*types.Service, error) {
	baseDomainAnnos := utils.GetAnnotation(s.Annotations, consts.ApiBaseDomain)
	baseDomain := ""
	if len(baseDomainAnnos) > 0 {
		baseDomain = baseDomainAnnos[0]
	}
	serviceGroupAnnos := utils.GetAnnotation(s.Annotations, consts.ApiServiceGroup)
	serviceGroup := ""
	if len(serviceGroupAnnos) > 0 {
		serviceGroup = serviceGroupAnnos[0]
	}

	service := &types.Service{
		Name:         s.Name,
		Package:      file.Package,
		PackageName:  file.PackageName,
		BaseDomain:   baseDomain,
		ServiceGroup: serviceGroup,
	}

	httpMethods := make([]*types.HttpMethod, 0)
	clientMethods := make([]*types.ClientMethod, 0)
	ms := s.GetFunctions()
	//var ms []*parser.Function
	resolver := golang.NewResolver(p.cu.RootScope(), p.cu)

	extendFunctions, err := p.getAllExtendFunction(s, ast, resolver)
	if err != nil {
		return nil, err
	}

	ms = append(ms, extendFunctions...)
	for _, m := range ms {
		method, err := p.makeHttpMethod(service, m)
		if err != nil {
			return nil, err
		}

		httpMethods = append(httpMethods, method)
	}

	service.Methods = httpMethods
	service.ClientMethods = clientMethods
	return service, nil
}

func (p *Parser) makeHttpMethod(s *types.Service, f *parser.Function) (*types.HttpMethod, error) {
	sr, _ := utils.GetFirstKV(utils.GetAnnotations(f.Annotations, consts.SerializerTags))
	rs := utils.GetAnnotations(f.Annotations, consts.HttpMethodAnnotations)
	if len(rs) == 0 {
		return nil, fmt.Errorf("no http method annotations found for method %s.%s", s.Name, f.Name)
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
		return nil, fmt.Errorf("invalid api.%s  for %s.%s: %s", hmethod, s.Name, f.Name, path)
	}

	resolver := golang.NewResolver(p.cu.RootScope(), p.cu)

	var reqName, reqRawName, reqPackage, reqArgName string
	if len(f.Arguments) >= 1 {
		reqArgName = f.Arguments[0].GetName()
		if len(f.Arguments) > 1 {
			slog.Warn(fmt.Sprintf("function '%s' has more than one argument, but only the first can be used in hertz now", f.GetName()))
		}
		var err error
		reqResolveTypeName, err := resolver.ResolveTypeName(f.Arguments[0].GetType())
		if err != nil {
			return nil, err
		}
		if imps := p.getImports(f.Arguments[0].GetType()); len(imps) > 0 {
			s.AddImports(imps...)
		}
		reqName = reqResolveTypeName.Deref().String()
		if strings.Contains(reqName, ".") && !f.Arguments[0].GetType().Category.IsContainerType() {
			// If reqName contains "." , then it must be of the form "pkg.name".
			// so reqRawName='name', reqPackage='pkg'
			names := strings.Split(reqName, ".")
			if len(names) != 2 {
				return nil, fmt.Errorf("request name: %s is wrong", reqName)
			}
			reqRawName = names[1]
			reqPackage = names[0]
		}
	}
	var respName, respRawName, respPackage string
	if !f.Oneway {
		var err error
		respResolveTypeName, err := resolver.ResolveTypeName(f.GetFunctionType())
		if err != nil {
			return nil, err
		}

		if imps := p.getImports(f.GetFunctionType()); len(imps) > 0 {
			s.AddImports(imps...)
		}

		respName = respResolveTypeName.Deref().String()
		if strings.Contains(respName, ".") && !f.GetFunctionType().Category.IsContainerType() {
			names := strings.Split(respName, ".")
			if len(names) != 2 {
				return nil, fmt.Errorf("response name: %s is wrong", respName)
			}
			// If respName contains "." , then it must be of the form "pkg.name".
			// so respRawName='name', respPackage='pkg'
			respRawName = names[1]
			respPackage = names[0]
		}
	}

	return &types.HttpMethod{
		Name:               f.Name,
		Comment:            f.GetReservedComments(),
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
	}, nil
}

func (p *Parser) getImports(t *parser.Type) (res []types.PkgInfo) {
	switch t.Name {
	case "void":
		return nil
	case "bool", "byte", "i8", "i16", "i32", "i64", "double", "string", "binary":
		return nil
	case "map":
		res = append(res, p.getImports(t.KeyType)...)
		fallthrough
	case "set", "list":
		res = append(res, p.getImports(t.ValueType)...)
		return res
	default:
		if ref := t.GetReference(); ref != nil {
			inc := p.cu.RootScope().Includes().ByIndex(int(ref.GetIndex()))
			res = append(res, types.PkgInfo{
				PkgName:    inc.PackageName,
				PkgRefName: inc.PackageName,
				ImportPath: inc.ImportPath,
			})
		}
		return
	}
}

func (p *Parser) getAllExtendFunction(svc *parser.Service, ast *parser.Thrift, resolver *golang.Resolver) (res []*parser.Function, err error) {
	if len(svc.Extends) == 0 {
		return
	}
	parts := semantic.SplitType(svc.Extends)
	fmt.Println(ast)
	switch len(parts) {
	case 1:
		// find the service in the same file
	case 2:
		refAst, found := ast.GetReference(parts[0])
		if found {
			extendSvc, found := refAst.GetService(parts[1])
			if found {
				funcs := extendSvc.GetFunctions()

				//resolver := golang.NewResolver(p.cu.RootScope(), p.cu)
				//base := p.cu.GetPackageName(refAst)
				//base := utils.BaseName(refAst.GetFilename(), "")
				//fmt.Println("base = ", base)
				for _, f := range funcs {
					//golang.NewResolver().

					processExtendsType(f, p.cu.GetPackageName(refAst))
					//scope, _ := golang.BuildScope(p.cu, refAst)
					//fmt.Println(ast)
					ty, err := resolver.ResolveTypeName(f.GetFunctionType())
					fmt.Println("ty = ", ty, err)
					res = append(res, f)
					fmt.Println(f)
				}

			}
		}
		return res, nil
	}
	return res, nil
}

func processExtendsType(f *parser.Function, base string) {
	// the method of other file is extended, and the package of req/resp needs to be changed
	// ex. base.thrift -> Resp Method(Req){}
	//					  base.Resp Method(base.Req){}
	if len(f.Arguments) > 0 {
		if f.Arguments[0].Type.Category.IsContainerType() {
			switch f.Arguments[0].Type.Category {
			case parser.Category_Set, parser.Category_List:
				if !strings.Contains(f.Arguments[0].Type.ValueType.Name, ".") && f.Arguments[0].Type.ValueType.Category.IsStruct() {
					f.Arguments[0].Type.ValueType.Name = base + "." + f.Arguments[0].Type.ValueType.Name
				}
			case parser.Category_Map:
				if !strings.Contains(f.Arguments[0].Type.ValueType.Name, ".") && f.Arguments[0].Type.ValueType.Category.IsStruct() {
					f.Arguments[0].Type.ValueType.Name = base + "." + f.Arguments[0].Type.ValueType.Name
				}
				if !strings.Contains(f.Arguments[0].Type.KeyType.Name, ".") && f.Arguments[0].Type.KeyType.Category.IsStruct() {
					f.Arguments[0].Type.KeyType.Name = base + "." + f.Arguments[0].Type.KeyType.Name
				}
			}
		} else {
			if !strings.Contains(f.Arguments[0].Type.Name, ".") && f.Arguments[0].Type.Category.IsStruct() {
				f.Arguments[0].Type.Name = base + "." + f.Arguments[0].Type.Name
			}
		}
	}

	if f.FunctionType.Category.IsContainerType() {
		switch f.FunctionType.Category {
		case parser.Category_Set, parser.Category_List:
			if !strings.Contains(f.FunctionType.ValueType.Name, ".") && f.FunctionType.ValueType.Category.IsStruct() {
				f.FunctionType.ValueType.Name = base + "." + f.FunctionType.ValueType.Name
			}
		case parser.Category_Map:
			if !strings.Contains(f.FunctionType.ValueType.Name, ".") && f.FunctionType.ValueType.Category.IsStruct() {
				f.FunctionType.ValueType.Name = base + "." + f.FunctionType.ValueType.Name
			}
			if !strings.Contains(f.FunctionType.KeyType.Name, ".") && f.FunctionType.KeyType.Category.IsStruct() {
				f.FunctionType.KeyType.Name = base + "." + f.FunctionType.KeyType.Name
			}
		}
	} else {
		if !strings.Contains(f.FunctionType.Name, ".") && f.FunctionType.Category.IsStruct() {
			f.FunctionType.Name = base + "." + f.FunctionType.Name
		}
	}
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
