package parser

type Parser interface {
	Parse(files ...string) ([]*Package, error)
}

type Options interface {
	NewParser() (Parser, error)
}
