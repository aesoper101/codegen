package utils

import "github.com/jinzhu/copier"

func Copy(source interface{}, target interface{}) error {
	return copier.Copy(target, source)
}

func DeepCopy(source interface{}, target interface{}) error {
	return copier.CopyWithOption(target, source, copier.Option{
		DeepCopy: true,
	})
}

func CopyWithOption(source interface{}, target interface{}, option copier.Option) error {
	return copier.CopyWithOption(target, source, option)
}
