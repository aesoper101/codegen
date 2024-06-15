package thrift

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConvertor_Convert(t *testing.T) {
	c := NewConvertor("")
	pkgs, err := c.Convert("../testdata/hello.thrift")

	require.Nil(t, err)
	require.Equal(t, 1, len(pkgs))

	data, _ := json.Marshal(pkgs)
	fmt.Println(string(data))
}
