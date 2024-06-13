package types

type Parser interface {
	Parse(files ...string) ([]*HttpPackage, error)
}
