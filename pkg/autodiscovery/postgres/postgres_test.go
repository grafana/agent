package postgres

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	p, err := New()
	require.NoError(t, err)
	res, err := p.Run()
	require.NoError(t, err)
	fmt.Println(res, err)
}
