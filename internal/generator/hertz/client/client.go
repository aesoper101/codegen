package hzclient

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/generator"
	"github.com/aesoper101/x/filex"
	"github.com/aesoper101/x/golangx"
	"github.com/charmbracelet/huh"
	"os"
	"os/exec"
	"path/filepath"
)

var _ generator.Generator = (*Generator)(nil)

type Generator struct {
	idlPath    string
	outPath    string
	modelMod   string
	baseDomain string
}

func New(opts ...CliGeneratorOption) *Generator {
	moduleName, _, _ := golangx.SearchGoMod(filex.Getwd())

	m := &Generator{
		idlPath:  "idl",
		outPath:  "model",
		modelMod: moduleName + "/model",
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

func (g *Generator) Name() string {
	return "HertzClient"
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

	if m.modelMod == "" {
		err := huh.NewInput().
			Title("Please provide a model module path").
			Validate(func(s string) error {
				return filex.MkdirIfNotExist(s)
			}).
			Value(&m.modelMod).
			Run()
		if err != nil {
			return err
		}
	}

	if err := filex.MkdirIfNotExist(m.outPath); err != nil {
		return err
	}

	return nil
}

func (g *Generator) Generate() error {
	if err := g.beforeGenerate(); err != nil {
		return err
	}

	return filepath.WalkDir(g.idlPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".thrift" {
			return g.runCmd(path)
		}
		return nil
	})
}

func (g *Generator) runCmd(idl string) error {
	fmt.Println("Generating client from", idl)

	args := []string{"client", "--idl", idl, "--client_dir", g.outPath, "--use", g.modelMod}
	if g.baseDomain != "" {
		args = append(args, "--base_domain", g.baseDomain)
	}

	cmd := exec.Command("hz", args...)
	return cmd.Run()
}
