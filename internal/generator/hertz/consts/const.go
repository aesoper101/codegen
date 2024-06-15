package consts

const (
	AnnotationQuery    = "api.query"
	AnnotationForm     = "api.form"
	AnnotationPath     = "api.path"
	AnnotationHeader   = "api.header"
	AnnotationCookie   = "api.cookie"
	AnnotationBody     = "api.body"
	AnnotationRawBody  = "api.raw_body"
	AnnotationJsConv   = "api.js_conv"
	AnnotationNone     = "api.none"
	AnnotationFileName = "api.file_name"

	AnnotationValidator = "api.vd"

	AnnotationGoTag = "go.tag"
)

const (
	ApiGet        = "api.get"
	ApiPost       = "api.post"
	ApiPut        = "api.put"
	ApiPatch      = "api.patch"
	ApiDelete     = "api.delete"
	ApiOptions    = "api.options"
	ApiHEAD       = "api.head"
	ApiAny        = "api.any"
	ApiPath       = "api.path"
	ApiSerializer = "api.serializer"
	ApiGenPath    = "api.handler_path"
)

const (
	ApiBaseDomain   = "api.base_domain"
	ApiServiceGroup = "api.service_group"
)

var (
	HttpMethodAnnotations = map[string]string{
		ApiGet:     "GET",
		ApiPost:    "POST",
		ApiPut:     "PUT",
		ApiPatch:   "PATCH",
		ApiDelete:  "DELETE",
		ApiOptions: "OPTIONS",
		ApiHEAD:    "HEAD",
		ApiAny:     "ANY",
	}

	HttpMethodOptionAnnotations = map[string]string{
		ApiGenPath: "handler_path",
	}

	BindingTags = map[string]string{
		AnnotationPath:    "path",
		AnnotationQuery:   "query",
		AnnotationHeader:  "header",
		AnnotationCookie:  "cookie",
		AnnotationBody:    "json",
		AnnotationForm:    "form",
		AnnotationRawBody: "raw_body",
	}

	SerializerTags = map[string]string{
		ApiSerializer: "serializer",
	}

	ValidatorTags = map[string]string{AnnotationValidator: "vd"}
)

// template const value
const (
	SetBodyParam = "setBodyParam(req).\n"
)
