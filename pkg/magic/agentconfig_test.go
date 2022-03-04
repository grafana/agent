package magic

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var seed = `
server: {}
`

func TestNewIntegration(t *testing.T) {
	ac := newConfigHandle(seed, "")
	output, err := ac.addWindowsIntegration()
	require.NoError(t, err)
	assert.True(t, strings.Contains(output, "integrations"))
	assert.True(t, strings.Contains(output, "windows_exporter"))
	assert.True(t, strings.Contains(output, "enabled: true"))

}
