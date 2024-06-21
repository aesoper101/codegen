package hzmodel

import (
	"encoding/json"
	"fmt"
	"github.com/cloudwego/thriftgo/args"
	"github.com/cloudwego/thriftgo/config"
	"github.com/cloudwego/thriftgo/generator"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
	"github.com/cloudwego/thriftgo/semantic"
	"github.com/cloudwego/thriftgo/version"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	var g generator.Generator
	var a args.Arguments

	a = args.Arguments{
		AskVersion:   false,
		Recursive:    true,
		Verbose:      false,
		Quiet:        false,
		CheckKeyword: true,
		OutputPath:   "./testdata/output",
		Includes:     nil,
		Plugins:      nil,
		Langs: []string{
			"go:reserve_comments,gen_json_tag=false",
		},
		IDL:             "",
		PluginTimeLimit: plugin.MaxExecutionTime,
	}

	//err := a.Parse([]string{
	//	"thriftgo",
	//	"-g",
	//	"go:reserve_comments,gen_json_tag=false",
	//	"-r",
	//	"-o",
	//	"./testdata/output",
	//	"./testdata/thrift/echo.thrift",
	//})
	//require.Nil(t, err)

	b, _ := json.Marshal(a)
	fmt.Println(string(b))

	err := config.LoadConfig()
	require.Nil(t, err)

	log := a.MakeLogFunc()

	_ = g.RegisterBackend(new(golang.GoBackend))

	ast, err := parser.ParseFile("./testdata/thrift/echo.thrift", nil, true)
	require.Nil(t, err)

	err = semantic.ResolveSymbols(ast)
	require.Nil(t, err)

	out := "./testdata/output"
	req := &plugin.Request{
		Version:    version.ThriftgoVersion,
		OutputPath: out,
		Recursive:  true,
		AST:        ast,
	}

	plugin.MaxExecutionTime = a.PluginTimeLimit
	plugins, err := a.UsedPlugins()
	require.Nil(t, err)

	langs, err := a.Targets()
	require.Nil(t, err)

	for _, out := range langs {
		out.UsedPlugins = plugins
		req.Language = out.Language
		req.OutputPath = a.Output(out.Language)

		arg := &generator.Arguments{Out: out, Req: req, Log: log}
		res := g.Generate(arg)
		log.MultiWarn(res.Warnings)

		err = g.Persist(res)
		require.Nil(t, err)
	}
}

func TestGenerator_Generate2(t *testing.T) {
	//in := []byte(`protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative helloworld.proto`)
	//comp := protocompile.Compiler{
	//	Resolver:       &protocompile.SourceResolver{},
	//	SourceInfoMode: protocompile.SourceInfoStandard,
	//}
	//ctx := context.Background()
	//files, err := comp.Compile(ctx, "./testdata/buf/hello.proto")
	//require.Nil(t, err)
	//require.Equal(t, 1, len(files))
	//
	//for _, file := range files {
	//
	//}

	//req := &pluginpb.CodeGeneratorRequest{
	//	//Parameter: proto.String("go:plugins=grpc"),
	//	//FileToGenerate: []string{
	//	//	"./testdata/buf/hello.proto",
	//	//},
	//}

	//req := &pluginpb.CodeGeneratorRequest{}
	//in := []byte(`protoc --go_out=. --go_opt=paths=source_relative --proto-path=./testdata/buf ./testdata/buf/hello.proto`)

	//args := []string{
	//	"protoc",
	//	"--go_out=.",
	//	"--go_opt=paths=source_relative",
	//	"--proto_path=./testdata/buf",
	//	"./testdata/buf/hello.proto",
	//}
	//in := []byte(strings.Join(args, " "))
	//err := proto.Unmarshal(in, req)
	//require.Nil(t, err)

	// 往 stdin 写入数据
	//os.Stdin.Write(in)

	// TODO 还是得编写插件才行

	cmd := exec.Command("protoc", "--go_out=.", "--go_opt=paths=source_relative", "--proto_path=./testdata/buf", "-I", "./testdata/buf/hello.proto")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	require.Nil(t, err)

	//opts := protogen.Options{
	//	ParamFunc: flag.CommandLine.Set,
	//}
	//opts.Run(func(p *protogen.Plugin) error {
	//	fmt.Println(p.Request)
	//	return nil
	//})

	//require.Nil(t, err)
	//require.NotNil(t, gen)
	//
	//response := gen.Response()
	//require.NotNil(t, response)
}
