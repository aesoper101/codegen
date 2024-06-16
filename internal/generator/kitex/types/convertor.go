package types

type Convertor interface {
	Convert(files ...string) (out []*PackageInfo, err error)
}
