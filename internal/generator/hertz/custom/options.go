package custom

import "github.com/aesoper101/codegen/pkg/utils"

type Option func(*Generator) error

type IdlType string

const (
	Proto  IdlType = "proto"
	Thrift IdlType = "thrift"
)

// WithPackagePrefix 设置生成代码的包前缀
func WithPackagePrefix(prefix string) Option {
	return func(g *Generator) error {
		g.packagePrefix = prefix
		return nil
	}
}

func WithIdlType(idlType IdlType, files ...string) Option {
	return func(g *Generator) error {
		g.idlType = idlType
		g.idlPaths = files
		switch idlType {
		case Proto, Thrift:
			return nil
		default:
			return ErrNoSupportIdlType
		}
	}
}

// WithTplPaths 设置模板文件路径
func WithTplPaths(paths ...string) Option {
	return func(g *Generator) error {
		g.tplPaths = utils.GetYamlFiles(paths)
		return nil
	}
}
