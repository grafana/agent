package mssql_exporter

import (
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/mssql/common"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestConfig_UnmarshalYaml(t *testing.T) {
	t.Run("only required values", func(t *testing.T) {
		strConfig := `connection_string: "sqlserver://user:pass@localhost:1433"`

		var c Config

		require.NoError(t, yaml.UnmarshalStrict([]byte(strConfig), &c))

		require.Equal(t, Config{
			Config: common.Config{
				ConnectionString:   "sqlserver://user:pass@localhost:1433",
				MaxIdleConnections: 3,
				MaxOpenConnections: 3,
				Timeout:            10 * time.Second,
			},
		}, c)
	})

	t.Run("all values", func(t *testing.T) {
		strConfig := `
connection_string: "sqlserver://user:pass@localhost:1433"
max_idle_connections: 5
max_open_connections: 6
timeout: 1m
`

		var c Config

		require.NoError(t, yaml.UnmarshalStrict([]byte(strConfig), &c))

		require.Equal(t, Config{
			Config: common.Config{
				ConnectionString:   "sqlserver://user:pass@localhost:1433",
				MaxIdleConnections: 5,
				MaxOpenConnections: 6,
				Timeout:            time.Minute,
			},
		}, c)
	})
}

func TestConfig_Identifier(t *testing.T) {
	t.Run("Identifier is in common config", func(t *testing.T) {
		c := Config{}

		ik := "mssql-instance-key"
		c.Common.InstanceKey = &ik

		id, err := c.Identifier(integrations_v2.Globals{})
		require.NoError(t, err)
		require.Equal(t, ik, id)
	})

	t.Run("Identifier is not in common config", func(t *testing.T) {
		c := Config{}
		c.Config.ConnectionString = "sqlserver://user:pass@localhost:1433"

		id, err := c.Identifier(integrations_v2.Globals{})
		require.NoError(t, err)
		require.Equal(t, "localhost:1433", id)
	})

	t.Run("Identifier is not in common config, invalid URL", func(t *testing.T) {
		c := Config{}
		c.Config.ConnectionString = "\u0001"

		_, err := c.Identifier(integrations_v2.Globals{})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to parse connection string URL:")
	})
}

func TestConfig_NewIntegration(t *testing.T) {
	t.Run("integration with valid config", func(t *testing.T) {
		c := &Config{
			Config: common.Config{
				ConnectionString:   "sqlserver://user:pass@localhost:1433",
				MaxIdleConnections: 3,
				MaxOpenConnections: 3,
				Timeout:            10 * time.Second,
			},
		}

		i, err := c.NewIntegration(log.NewJSONLogger(os.Stdout), integrations_v2.Globals{})
		require.NoError(t, err)
		require.NotNil(t, i)
	})

	t.Run("integration with invalid config", func(t *testing.T) {
		c := &Config{
			Config: common.Config{
				ConnectionString:   "\u0001",
				MaxIdleConnections: 3,
				MaxOpenConnections: 3,
				Timeout:            10 * time.Second,
			},
		}

		i, err := c.NewIntegration(log.NewJSONLogger(os.Stdout), integrations_v2.Globals{})
		require.Nil(t, i)
		require.ErrorContains(t, err, "failed to validate config:")
	})
}

func TestConfig_ApplyDefaults(t *testing.T) {
	c := Config{}

	err := c.ApplyDefaults(integrations_v2.Globals{})
	require.NoError(t, err)
}
