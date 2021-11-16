package integrations

import (
	"testing"

	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestIntegration can be used by integrations to run common tests against
// their integration implementation.
func TestIntegration(t *testing.T, c Config) {
	t.Helper()

	t.Run("Common Defaults", func(t *testing.T) {
		c := cloneIntegration(c)
		err := yaml.Unmarshal([]byte(`{}`), c)
		require.NoError(t, err)
		require.Equal(t, config.DefaultCommon, c.CommonConfig(), "integrations should unmarshal with common config defaults")
	})
}
