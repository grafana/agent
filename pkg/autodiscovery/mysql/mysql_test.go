package mysql

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMysql(t *testing.T) {
	m, err := New()
	require.NoError(t, err)
	res, err := m.Run()
	require.NoError(t, err)
	// fmt.Println(res.RiverConfig)
	fmt.Fprintf(os.Stdout, res.RiverConfig)
}
