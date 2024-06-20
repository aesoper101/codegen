package protobuf

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	p, err := NewParser()
	require.Nil(t, err)
	require.NotNil(t, p)

	pkg, err := p.Parse("../testdata/proto/hello.proto")
	require.Nil(t, err)
	require.Equal(t, 0, len(pkg))
}
