package parser

type Service struct {
	File *File `json:"-"`

	Name          string
	Methods       []*HttpMethod
	ClientMethods []*ClientMethod
	BaseDomain    string // base domain for client code
	ServiceGroup  string // service level router group
}
