package thrift

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewParser(t *testing.T) {
	c, err := newParser(&Options{
		PackagePrefix: "github.com/aesoper101/xx",
	})

	require.Nil(t, err)
	require.NotNil(t, c)

	ps, err := c.Parse("../testdata/hello.thrift")
	require.Nil(t, err)
	require.Equal(t, 1, len(ps))

	data, err := json.Marshal(&ps)
	t.Log(err)
	require.Nil(t, err)

	fmt.Println(string(data))
}
