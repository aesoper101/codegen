package types

type Method struct {
	typ Type
}

func NewMethod(typ Type) *Method {
	return &Method{typ: typ}
}

func (m *Method) Type() Type {
	return m.typ
}

func (m *Method) Name() string {
	return m.typ.Name()
}
