package v2

type Service struct {
	Name          string
	Methods       []*HttpMethod
	ClientMethods []*ClientMethod
	BaseDomain    string // base domain for client code
	ServiceGroup  string // service level router group
}
