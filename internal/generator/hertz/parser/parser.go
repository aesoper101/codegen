package parser

type Parser interface {
	Parse(files ...string) ([]*Package, error)
}
