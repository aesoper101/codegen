package thrift

import "github.com/cloudwego/thriftgo/generator/golang"

type Option func(convertor *options) error

type options struct {
	packagePrefix string
	features      golang.Features
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
