package kubernetes

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestK8S(t *testing.T) {
	os.Setenv("KUBERNETES_SERVICE_HOST", "10.188.0.1")
	os.Setenv("KUBERNETES_SERVICE_PORT", "443")

	m, err := New()
	require.NoError(t, err)
	res, err := m.Run()
	require.NoError(t, err)
	fmt.Println(res, err)
}
