package catchpoint_exporter

import (
	"os"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestConfig_UnmarshalYaml(t *testing.T) {
	strConfig := `
	port        = "3030"
	verbose     = true
	webhookpath = "/nondefault-webhook-path"
	`

	var c Config

	require.NoError(t, yaml.UnmarshalStrict([]byte(strConfig), &c))

	require.Equal(t, Config{
		Verbose:     true,
		Port:        "3030",
		WebhookPath: "/nondefault-webhook-path",
	}, c)
}

func TestConfig_NewIntegration(t *testing.T) {
	t.Run("integration with valid config", func(t *testing.T) {
		c := &Config{
			Verbose:     true,
			Port:        "3030",
			WebhookPath: "/nondefault-webhook-path",
		}

		i, err := c.NewIntegration(log.NewJSONLogger(os.Stdout))
		require.NoError(t, err)
		require.NotNil(t, i)
	})

	t.Run("integration with invalid config", func(t *testing.T) {
		c := &Config{
			Verbose:     "incorrect_value",
			Port:        "3030",
			WebhookPath: "/nondefault-webhook-path",
		}

		i, err := c.NewIntegration(log.NewJSONLogger(os.Stdout))
		require.Nil(t, i)
		require.ErrorContains(t, err, "")
	})
}

func TestConfig_AgentKey(t *testing.T) {
	c := DefaultConfig
	c.Port = "3030"

	ik := "agent-key"
	id, err := c.InstanceKey(ik)
	require.NoError(t, err)
	require.Equal(t, "localhost:3030", id)
}
