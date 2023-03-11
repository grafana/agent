package prometheus

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	m, err := New()
	require.NoError(t, err)
	res, err := m.Run()
	require.NoError(t, err)
	fmt.Println(res, err)
}
