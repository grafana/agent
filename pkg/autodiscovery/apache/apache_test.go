package apache_test

import (
	"fmt"
	"testing"

	"github.com/grafana/agent/pkg/autodiscovery/apache"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	m, err := apache.New()
	require.NoError(t, err)
	res, err := m.Run()
	require.NoError(t, err)
	fmt.Println(res, err)
}
