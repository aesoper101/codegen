package custom

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	g, err := New(WithIdlType(Thrift, "../testdata"), WithTplPaths("../testdata"))
	require.Nil(t, err)
	require.NotNil(t, g)

	err = g.Generate()
	require.Nil(t, err)
}
