package thrift

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewParser(t *testing.T) {
	p := NewParser()

	packages, err := p.Parse("./testdata")

	require.Nil(t, err)
	require.Equal(t, 2, len(packages))

	data, _ := json.Marshal(packages)
	fmt.Println(string(data))
}
