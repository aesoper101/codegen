package parser

import (
	"encoding/json"
	"fmt"
	"github.com/aesoper101/codegen/internal/generator/hertz/config"
	"github.com/aesoper101/x/flagx"
	"github.com/stretchr/testify/require"
	"testing"
)

var testIdlContent = `
namespace go api

struct QueryStudentRequest {
    1: i32 Num (api.query="num", api.vd="$<100; msg:'num must less than 100'"); // 学号，通过query参数进行绑定
}

struct QueryStudentResponse {
    1: string Num;
    2: string Name;
    3: string Gender;
    4: string Msg; // 返回信息，如果没有查询到则返回原因

}

struct InsertStudentRequest {
    1: string Num (api.form="num");
    2: string Name (api.form="name");
    3: string Gender (api.form="gender");
}

struct InsertStudentResponse {
    1: bool Ok; // 是否插入成功
    2: string Msg; // 返回信息，如果没有查询到则返回原因
}

service StudentApi {
   // 查询接口：queryStudent
   // 功能： 根据query参数中提供的学号来查询学生信息
   QueryStudentResponse queryStudent(1: QueryStudentRequest req) (api.get="student/query");
   // 插入接口：insertStudent
   // 功能： 以学号为key，插入学生的信息
   InsertStudentResponse insertStudent(1: InsertStudentRequest req) (api.post="student/insert");
}
`

func Test_astToService(t *testing.T) {
	args := config.NewArgument()

	fg := flagx.NewFlagSet("test")
	fg.StringSlice("idl", []string{}, "thrift idl file")
	fg.String("model_dir", "./model", "model dir")

	err := fg.Parse([]string{
		"--idl", "student_api.thrift",
		"--model_dir", "model",
	})
	require.Nil(t, err)

	nArgs, err := args.Parse(fg, "model")
	require.Nil(t, err)

	cmd, err := config.BuildPluginCmd(nArgs)
	require.Nil(t, err)
	require.NotNil(t, cmd)

	fmt.Printf("%+v\n", cmd.Args)

	p, err := New(WithArgument(nArgs))

	require.Nil(t, err)
	require.NotNil(t, p)

	packages, err := p.Parse()
	require.Nil(t, err)

	printTable(packages[0].Models)
	jsonData, _ := json.Marshal(packages)
	fmt.Println(string(jsonData))

}
