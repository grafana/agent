package tree

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddComponent(t *testing.T) {
	tr := &Tree{}
	err := tr.Parse("t", []byte(""))
	require.NoError(t, err)
	err = tr.AddComponent([]byte(`comp {t="1"}`))
	require.NoError(t, err)
	output, err := tr.Print()
	require.NoError(t, err)
	require.True(t, strings.Contains(output, "comp"))
}
