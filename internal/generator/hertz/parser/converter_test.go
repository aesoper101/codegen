package parser

import (
	"encoding/json"
	"fmt"
	"github.com/aesoper101/x/tablewriterx"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewConverter(t *testing.T) {
	c := NewConverter()
	err := c.Convert("./testdata")

	require.Nil(t, err)
	require.Len(t, c.Services(), 2)

	printTable(c.Services())
	printJSON(c.Services())
}

func printTable(data interface{}) {
	tab := tablewriterx.NewWriter()
	_ = tab.SetStructs(data)
	tab.Render()
}

func printJSON(data interface{}) {
	b, _ := json.Marshal(data)
	fmt.Println(string(b))
}
