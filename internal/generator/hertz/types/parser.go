package types

type Convertor interface {
	Convert(files ...string) ([]*HttpPackage, error)
}
