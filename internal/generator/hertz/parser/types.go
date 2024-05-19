package parser

type Service struct {
	Name    string
	Methods []Method
}

type Method struct {
	Name string
}
