package hertz

import (
	"github.com/aesoper101/codegen/pkg/utils"
)

type options struct {
	configFiles []string
}

func (o *options) apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return err
		}
	}
	return nil
}

type Option func(*options) error

// WithConfigFiles 设置配置文件路径
func WithConfigFiles(files ...string) Option {
	return func(o *options) error {
		files = utils.GetYamlFiles(files)
		o.configFiles = files
		return nil
	}
}
