package generator

type IdlType string

const (
	Proto  IdlType = "proto"
	Thrift IdlType = "thrift"
)

func (i IdlType) String() string {
	return string(i)
}

func (i IdlType) IsValid() bool {
	switch i {
	case Proto, Thrift:
		return true
	default:
		return false
	}
}

func IDLTypeFromString(s string) IdlType {
	switch s {
	case Proto.String():
		return Proto
	case Thrift.String():
		return Thrift
	default:
		return ""
	}
}
