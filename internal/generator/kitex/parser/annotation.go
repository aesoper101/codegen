package parser

import (
	"fmt"
	"strings"
)

type Annotations []*Annotation

// Append append key value pair to Annotation slice.
func (a *Annotations) Append(key, value string) {
	old := *a
	for _, anno := range old {
		if strings.ToLower(anno.Key) == strings.ToLower(key) {
			anno.Values = append(anno.Values, value)
			return
		}
	}
	*a = append(old, &Annotation{Key: key, Values: []string{value}})
}

// Get return annotations values.
func (a *Annotations) Get(key string) []string {
	for _, anno := range *a {
		if strings.ToLower(anno.Key) == strings.ToLower(key) {
			return anno.Values
		}
	}
	return nil
}

func (a *Annotations) GetFirstValue(key string) string {
	return a.ILocValueByKey(key, 0)
}

// ILocValueByKey return annotation value by key and index.
func (a *Annotations) ILocValueByKey(key string, idx int) string {
	for _, anno := range *a {
		if strings.ToLower(anno.Key) == strings.ToLower(key) && idx < len(anno.Values) {
			return anno.Values[idx]
		}
	}
	return ""
}

type Annotation struct {
	Key    string
	Values []string
}

func (p *Annotation) GetKey() (v string) {
	return p.Key
}

func (p *Annotation) GetValues() (v []string) {
	return p.Values
}

func (p *Annotation) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Annotation(%+v)", *p)
}
