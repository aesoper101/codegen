package thrift

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/generator/hertz/consts"
	"github.com/aesoper101/codegen/internal/generator/hertz/types"
	"github.com/aesoper101/codegen/pkg/utils"
	"github.com/aesoper101/x/str"
	"github.com/cloudwego/hertz/cmd/hz/meta"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
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

type Convertor struct {
	cache *Cache
	cu    *golang.CodeUtils
}

func NewConvertor(packagePrefix string) *Convertor {
	cu := golang.NewCodeUtils(backend.DummyLogFunc())
	cu.SetFeatures(defaultFeatures)
	cu.SetPackagePrefix(packagePrefix)

	return &Convertor{
		cache: newIDLCache(cu),
		cu:    cu,
	}
}

func (c *Convertor) Convert(files ...string) ([]*types.HttpPackage, error) {
	fs := utils.GetThriftFiles(files)
	for _, f := range fs {
		if _, err := c.cache.AddFile(f); err != nil {
			return nil, err
		}
	}

	return c.convert()
}

func (c *Convertor) convert() ([]*types.HttpPackage, error) {
	scopes := c.cache.GetWaitProcessScopes()
	var httpPackages []*types.HttpPackage

	for _, scope := range scopes {
		c.cu.SetRootScope(scope)

		resolver := golang.NewResolver(scope, c.cu)
		pkg, err := c.makePackage(scope, resolver)
		if err != nil {
			return nil, err
		}
		httpPackages = append(httpPackages, pkg)
	}
	return httpPackages, nil
}

func (c *Convertor) makePackage(scope *golang.Scope, resolver *golang.Resolver) (*types.HttpPackage, error) {
	pkg := &types.HttpPackage{
		IdlName:   scope.AST().GetFilename(),
		Package:   scope.FilePackage(),
		Namespace: scope.AST().GetNamespaceOrReferenceName("go"),
	}

	var services []*types.Service
	for _, s := range scope.Services() {
		svc, err := c.makeService(s, resolver)
		if err != nil {
			return nil, err
		}
		services = append(services, svc)
	}

	pkg.AddServices(services...)
	return pkg, nil
}

func (c *Convertor) makeService(s *golang.Service, resolver *golang.Resolver) (*types.Service, error) {
	svc := &types.Service{
		Name: s.GetName(),
	}

	baseDomainAnnos := utils.GetAnnotation(s.Annotations, consts.ApiBaseDomain)
	if len(baseDomainAnnos) > 0 {
		svc.BaseDomain = baseDomainAnnos[0]
	}
	serviceGroupAnnos := utils.GetAnnotation(s.Annotations, consts.ApiServiceGroup)
	if len(serviceGroupAnnos) > 0 {
		svc.ServiceGroup = serviceGroupAnnos[0]
	}

	ms := s.Functions()

	extendSvcs, err := c.getAllExtendFunction(s, resolver)
	if err != nil {
		return nil, err
	}

	ms = append(ms, extendSvcs...)

	httpMethods := make([]*types.HttpMethod, 0)
	clientMethods := make([]*types.ClientMethod, 0)
	for _, m := range ms {
		method, err := c.makeHttpMethod(m, resolver)
		if err != nil {
			return nil, err
		}

		if err := checkHttpMethodSamePath(httpMethods, method); err != nil {
			return nil, err
		}
		httpMethods = append(httpMethods, method)

		clientMethod := &types.ClientMethod{
			HttpMethod: method,
		}
		if err := c.parseAnnotationToClient(clientMethod, m); err != nil {
			return nil, err
		}
		clientMethods = append(clientMethods, clientMethod)
	}

	imps, _ := s.From().ResolveImports()

	svc.Imports = imps

	svc.AddHttpMethods(httpMethods...)
	svc.ClientMethods = clientMethods
	return svc, nil
}

func (c *Convertor) parseAnnotationToClient(clientMethod *types.ClientMethod, m *golang.Function) error {
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
	clientMethod.BodyParamsCode = meta.SetBodyParam
	if hasBodyAnnotation && hasFormAnnotation {
		clientMethod.FormValueCode = ""
		clientMethod.FormFileCode = ""
	}
	if !hasBodyAnnotation && hasFormAnnotation {
		clientMethod.BodyParamsCode = ""
	}

	return nil
}

func checkHttpMethodSamePath(ms []*types.HttpMethod, m *types.HttpMethod) error {
	for _, method := range ms {
		if method.Path == m.Path {
			return fmt.Errorf("http method %s has same path %s with method %s", m.Name, m.Path, method.Name)
		}
	}
	return nil
}

func (c *Convertor) makeHttpMethod(m *golang.Function, resolver *golang.Resolver) (*types.HttpMethod, error) {
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

	var reqName, reqRawName, reqPackage, reqArgName string
	if len(m.GetArguments()) >= 1 {
		reqArgName = m.GetArguments()[0].GetName()
		if len(m.GetArguments()) > 1 {
			slog.Warn(fmt.Sprintf("function '%s' has more than one argument, but only the first can be used in hertz now", m.GetName()))
		}

		reqName = m.Arguments()[0].GetType().GetName()

		tyName, err := resolver.GetTypeName(m.Service().From(), m.Arguments()[0].GetType())
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
		}
	}
	var respName, respRawName, respPackage string
	if !m.Oneway {
		tyName, err := resolver.GetTypeName(m.Service().From(), m.GetFunctionType())
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
		}
	}

	return &types.HttpMethod{
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
	}, nil
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
