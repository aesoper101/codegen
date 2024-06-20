package thrift

import "github.com/cloudwego/thriftgo/generator/golang"

type Option func(opts *options) error

type options struct {
	packagePrefix string
	features      golang.Features
	importReplace map[string]string
}

func (o *options) apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return err
		}
	}
	return nil
}

func PackagePrefix(prefix string) Option {
	return func(o *options) error {
		o.packagePrefix = prefix
		return nil
	}
}

func Features(features golang.Features) Option {
	return func(o *options) error {
		o.features = features
		return nil
	}
}

func UsePackage(path, repl string) Option {
	return func(o *options) error {
		if o.importReplace == nil {
			o.importReplace = make(map[string]string)
		}
		o.importReplace[path] = repl
		return nil
	}
}
