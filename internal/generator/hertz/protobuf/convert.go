package protobuf

import "github.com/aesoper101/codegen/internal/generator/hertz/types"

var _ types.Convertor = (*Convertor)(nil)

type Convertor struct {
	packagePrefix string
}

func NewConvertor() (*Convertor, error) {
	return &Convertor{}, nil
}

func (c *Convertor) Convert(files ...string) ([]*types.HttpPackage, error) {
	//TODO implement me
	panic("implement me")
}