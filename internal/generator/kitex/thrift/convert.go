package thrift

import (
	"github.com/aesoper101/codegen/internal/generator/kitex/types"
	"github.com/aesoper101/codegen/pkg/utils"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/generator/golang/streaming"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
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

var _ types.Convertor = (*Convertor)(nil)

type Convertor struct {
	opts  options
	cache *utils.ThriftIDLCache
	cu    *golang.CodeUtils
}

func NewConvertor(opts ...Option) (types.Convertor, error) {
	o := options{
		features: defaultFeatures,
	}
	if err := o.apply(opts...); err != nil {
		return nil, err
	}

	cu := golang.NewCodeUtils(backend.DummyLogFunc())
	cu.SetFeatures(o.features)

	if o.packagePrefix != "" {
		cu.SetPackagePrefix(o.packagePrefix)
	}

	return &Convertor{
		opts:  o,
		cache: utils.NewThriftIDLCache(cu),
		cu:    cu,
	}, nil
}

func (c *Convertor) Convert(files ...string) ([]*types.PackageInfo, error) {
	fs := utils.GetThriftFiles(files)
	for _, f := range fs {
		if _, err := c.cache.AddFile(f); err != nil {
			return nil, err
		}
	}
	return c.convert()
}

func (c *Convertor) convert() (out []*types.PackageInfo, err error) {
	for _, scope := range c.cache.GetWaitProcessScopes() {
		pkg, err := c.convertPackage(scope)
		if err != nil {
			return nil, err
		}

		out = append(out, pkg)
	}
	return out, nil
}

func (c *Convertor) convertPackage(scope *golang.Scope) (*types.PackageInfo, error) {
	c.cu.SetRootScope(scope)
	resolver := golang.NewResolver(scope, c.cu)

	packageInfo := &types.PackageInfo{
		Namespace: scope.AST().GetNamespaceOrReferenceName("go"),
		Package:   scope.FilePackage(),
		IdlName:   scope.AST().GetFilename(),
	}

	var services []*types.ServiceInfo
	for _, service := range scope.Services() {
		serviceInfo, err := c.convertService(service, resolver)
		if err != nil {
			return nil, err
		}

		services = append(services, serviceInfo)
	}

	if err := packageInfo.AddServices(services...); err != nil {
		return nil, err
	}

	return packageInfo, nil
}

func (c *Convertor) convertService(service *golang.Service, resolver *golang.Resolver) (*types.ServiceInfo, error) {
	ast := service.From().AST()
	pkg, pth := c.cu.Import(ast)
	pi := types.PkgInfo{
		PkgName:    pkg,
		PkgRefName: pkg,
		ImportPath: pth,
	}

	serviceInfo := &types.ServiceInfo{
		ServiceFilePath: ast.GetFilename(),
		ServiceName:     service.GetName(),
		RawServiceName:  service.GetName(),
		PkgInfo:         pi,
	}

	methods := service.Functions()

	extendMethods, err := c.getAllExtendFunction(service, resolver)
	if err != nil {
		return nil, err
	}

	if len(extendMethods) > 0 {
		methods = append(methods, extendMethods...)
	}

	var methodInfos []*types.MethodInfo
	for _, method := range methods {
		methodInfo, err := c.convertMethod(serviceInfo, method, resolver)
		if err != nil {
			return nil, err
		}

		methodInfos = append(methodInfos, methodInfo)
	}

	if err := serviceInfo.AddMethods(methodInfos...); err != nil {
		return nil, err
	}

	serviceInfo.UseThriftReflection = c.cu.Features().WithReflection
	serviceInfo.ServiceTypeName = func() string { return serviceInfo.ServiceName }
	return serviceInfo, nil
}

func (c *Convertor) convertMethod(si *types.ServiceInfo, f *golang.Function,
	resolver *golang.Resolver) (*types.MethodInfo, error) {
	st, err := streaming.ParseStreaming(f.Function)
	if err != nil {
		return nil, err
	}

	mi := &types.MethodInfo{
		PkgInfo:            si.PkgInfo,
		ServiceName:        si.ServiceName,
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

	if !f.Oneway {
		mi.ResStructName = f.ResType().GoName().String()
	}
	if !f.Void {
		typeName := f.ResponseGoTypeName().String()
		//typeName, err := resolver.GetTypeName(f.Service().From(), f.GetFunctionType())
		//if err != nil {
		//	return nil, err
		//}
		mi.Resp = &types.Parameter{
			Deps: c.getImports(f.FunctionType),
			Type: typeName,
		}
		mi.IsResponseNeedRedirect = "*"+typeName == f.ResType().Fields()[0].GoTypeName().String()
	}

	for _, a := range f.Arguments() {
		//typeName, err := resolver.GetTypeName(f.Service().From(), a.GetType())
		//if err != nil {
		//	return nil, err
		//}

		arg := &types.Parameter{
			Deps:    c.getImports(a.GetType()),
			Name:    f.ArgType().Field(a.Name).GoName().String(),
			RawName: a.GoName().String(),
			Type:    a.GoTypeName().String(),
			//RawName: ,
			//Type: typeName.String(),
		}
		mi.Args = append(mi.Args, arg)
	}
	for _, t := range f.Throws() {
		ex := &types.Parameter{
			Deps:    c.getImports(t.Type),
			Name:    f.ResType().Field(t.Name).GoName().String(),
			RawName: t.GoName().String(),
			Type:    t.GoTypeName().String(),
		}
		mi.Exceptions = append(mi.Exceptions, ex)
	}
	return mi, nil
}

func (c *Convertor) getImports(t *parser.Type) (res []types.PkgInfo) {
	switch t.Name {
	case "void":
		return nil
	case "bool", "byte", "i8", "i16", "i32", "i64", "double", "string", "binary":
		return nil
	case "map":
		res = append(res, c.getImports(t.KeyType)...)
		fallthrough
	case "set", "list":
		res = append(res, c.getImports(t.ValueType)...)
		return res
	default:
		if ref := t.GetReference(); ref != nil {
			inc := c.cu.RootScope().Includes().ByIndex(int(ref.GetIndex()))
			res = append(res, types.PkgInfo{
				PkgRefName: inc.PackageName,
				ImportPath: inc.ImportPath,
			})
		}
		return
	}
}

func (c *Convertor) getAllExtendFunction(svc *golang.Service, resolver *golang.Resolver) (res []*golang.Function, err error) {
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
			extendFuncs, err := c.getAllExtendFunction(extendSvc, resolver)
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
				for _, f := range res {
					processExtendsType(f, inc.PackageName)
				}

				extendFuncs, err := c.getAllExtendFunction(extendSvc, resolver)
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

func processExtendsType(f *golang.Function, base string) {
	// the method of other file is extended, and the package of req/resp needs to be changed
	// ex. axx.common.thrift -> Resp Method(Req){}
	//					  base.Resp Method(base.Req){}
	if len(f.GetArguments()) > 0 {
		if f.GetArguments()[0].Type.Category.IsContainerType() {
			switch f.GetArguments()[0].Type.Category {
			case parser.Category_Set, parser.Category_List:
				if !strings.Contains(f.GetArguments()[0].Type.ValueType.Name, ".") && f.GetArguments()[0].Type.ValueType.Category.IsStruct() {
					f.GetArguments()[0].Type.ValueType.Name = base + "." + f.GetArguments()[0].Type.ValueType.Name
				}
			case parser.Category_Map:
				if !strings.Contains(f.GetArguments()[0].Type.ValueType.Name, ".") && f.GetArguments()[0].Type.ValueType.Category.IsStruct() {
					f.GetArguments()[0].Type.ValueType.Name = base + "." + f.GetArguments()[0].Type.ValueType.Name
				}
				if !strings.Contains(f.GetArguments()[0].Type.KeyType.Name, ".") && f.GetArguments()[0].Type.KeyType.Category.IsStruct() {
					f.GetArguments()[0].Type.KeyType.Name = base + "." + f.GetArguments()[0].Type.KeyType.Name
				}
			}
		} else {
			if !strings.Contains(f.GetArguments()[0].Type.Name, ".") && f.GetArguments()[0].Type.Category.IsStruct() {
				f.GetArguments()[0].Type.Name = base + "." + f.GetArguments()[0].Type.Name
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
