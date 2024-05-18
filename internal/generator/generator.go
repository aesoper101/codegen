package generator

type Generator interface {
	// Name 返回生成器的名称
	Name() string

	// Generate 生成代码
	Generate() error
}
