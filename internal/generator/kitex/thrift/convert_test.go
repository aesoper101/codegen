package thrift

import (
	"encoding/json"
	"fmt"
	"github.com/aesoper101/x/tablewriterx"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewConvertor(t *testing.T) {
	c, err := NewConvertor(PackagePrefix("github.com/aesoper101/x"))

	require.Nil(t, err)
	require.NotNil(t, c)

	ps, err := c.Convert("../testdata/echo.thrift")
	require.Nil(t, err)
	require.Equal(t, 1, len(ps))

	data, err := json.Marshal(&ps)
	t.Log(err)
	require.Nil(t, err)

	fmt.Println(string(data))
}

func printTable(in interface{}) {
	tab := tablewriterx.NewWriter()
	_ = tab.SetStructs(in)
	tab.Render()
}
