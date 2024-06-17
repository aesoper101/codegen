package types

type Convertor interface {
	Convert(files ...string) ([]*PackageInfo, error)
}
