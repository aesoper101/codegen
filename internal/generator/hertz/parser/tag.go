package parser

import (
	"github.com/cloudwego/thriftgo/parser"
	"strings"
)

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
	ApiBaseDomain    = "api.base_domain"
	ApiServiceGroup  = "api.service_group"
	ApiServiceGenDir = "api.service_gen_dir" // handler_dir for handler_by_service
	ApiServicePath   = "api.service_path"    // declare the path to the service's handler according to this annotation for handler_by_method
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

func getAnnotation(input parser.Annotations, target string) []string {
	if len(input) == 0 {
		return nil
	}
	for _, anno := range input {
		if strings.ToLower(anno.Key) == target {
			return anno.Values
		}
	}

	return []string{}
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

func getAnnotations(input parser.Annotations, targets map[string]string) map[string][]string {
	if len(input) == 0 || len(targets) == 0 {
		return nil
	}
	out := map[string][]string{}
	for k, t := range targets {
		var ret *parser.Annotation
		for _, anno := range input {
			if strings.ToLower(anno.Key) == k {
				ret = anno
				break
			}
		}
		if ret == nil {
			continue
		}
		out[t] = ret.Values
	}
	return out
}
