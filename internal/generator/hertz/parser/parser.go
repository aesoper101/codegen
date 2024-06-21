package parser

type Parser interface {
	Parse(files ...string) (Packages, error)
}

type Options interface {
	NewParser() (Parser, error)
}
