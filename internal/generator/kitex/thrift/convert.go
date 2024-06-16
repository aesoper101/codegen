package thrift

import "github.com/aesoper101/codegen/internal/generator/kitex/types"

var _ types.Convertor = (*Convertor)(nil)

type Convertor struct {
	opts options
}

func NewConvertor(opts ...Option) (types.Convertor, error) {
	o := options{}
	if err := o.apply(opts...); err != nil {
		return nil, err
	}

	return &Convertor{opts: o}, nil
}

func (c *Convertor) Convert(files ...string) (out []*types.PackageInfo, err error) {
	//TODO implement me
	panic("implement me")
}
