package hzmodel

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/generator"
	"github.com/aesoper101/codegen/pkg/utils"
	"github.com/aesoper101/x/filex"
	"github.com/aesoper101/x/golangx"
	"github.com/charmbracelet/huh"
	"os/exec"
)

var _ generator.Generator = (*Generator)(nil)

type Generator struct {
	idlPath string
	outPath string
	modName string
}

func New(opts ...ModelGeneratorOption) *Generator {
	moduleName, _, _ := golangx.SearchGoMod(filex.Getwd())
	m := &Generator{
		idlPath: "idl",
		outPath: "model",
		modName: moduleName,
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

func (m *Generator) beforeGenerate() error {
	if !filex.IsDir(m.idlPath) {
		var idlPath string
		err := huh.NewInput().
			Title("Please provide an input directory").
			Validate(func(s string) error {
				return filex.MkdirIfNotExist(s)
			}).
			Value(&idlPath).
			Run()
		if err != nil {
			return err
		}
		m.idlPath = idlPath
	}

	if err := filex.MkdirIfNotExist(m.outPath); err != nil {
		return err
	}

	return nil
}

func (m *Generator) Generate() error {
	if err := m.beforeGenerate(); err != nil {
		return err
	}

	files := utils.GetThriftFiles([]string{m.idlPath})
	for _, f := range files {
		if err := m.runCmd(f); err != nil {
			return err
		}
	}

	return nil
}

func (m *Generator) runCmd(idl string) error {
	fmt.Println("Generating model from", idl)

	args := []string{"model", "--idl", idl, "--model_dir", m.outPath}
	if m.modName != "" {
		args = append(args, "--module", m.modName)
	}

	cmd := exec.Command("hz", args...)
	return cmd.Run()
}

func (m *Generator) Name() string {
	return "HertzModel"
}
