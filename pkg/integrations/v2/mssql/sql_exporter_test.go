package snowflake_exporter

import (
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
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
			ConnectionString:   "sqlserver://user:pass@localhost:1433",
			MaxIdleConnections: 3,
			MaxConnections:     3,
			Timeout:            10 * time.Second,
		}, c)
	})

	t.Run("all values", func(t *testing.T) {
		strConfig := `
connection_string: "sqlserver://user:pass@localhost:1433"
max_idle_connections: 5
max_connections: 6
timeout: 1m
`

		var c Config

		require.NoError(t, yaml.UnmarshalStrict([]byte(strConfig), &c))

		require.Equal(t, Config{
			ConnectionString:   "sqlserver://user:pass@localhost:1433",
			MaxIdleConnections: 5,
			MaxConnections:     6,
			Timeout:            time.Minute,
		}, c)
	})
}

func TestConfig_validate(t *testing.T) {
	testCases := []struct {
		name  string
		input Config
		err   string
	}{
		{
			name: "valid config",
			input: Config{
				ConnectionString:   "sqlserver://user:pass@localhost:1433",
				MaxIdleConnections: 3,
				MaxConnections:     3,
				Timeout:            10 * time.Second,
			},
		},
		{
			name: "incorrect connection_string scheme",
			input: Config{
				ConnectionString:   "mysql://user:pass@localhost:1433",
				MaxIdleConnections: 3,
				MaxConnections:     3,
				Timeout:            10 * time.Second,
			},
			err: "scheme of provided connection_string URL must be sqlserver",
		},
		{
			name: "invalid URL",
			input: Config{
				ConnectionString:   "\u0001",
				MaxIdleConnections: 3,
				MaxConnections:     3,
				Timeout:            10 * time.Second,
			},
			err: "failed to parse connection_string",
		},
		{
			name: "missing connection_string",
			input: Config{
				MaxIdleConnections: 3,
				MaxConnections:     3,
				Timeout:            10 * time.Second,
			},
			err: "the connection_string parameter is required",
		},
		{
			name: "max connections is less than 1",
			input: Config{
				ConnectionString:   "sqlserver://user:pass@localhost:1433",
				MaxIdleConnections: 3,
				MaxConnections:     0,
				Timeout:            10 * time.Second,
			},
			err: "max_connections must be at least 1",
		},
		{
			name: "max idle connections is less than 1",
			input: Config{
				ConnectionString:   "sqlserver://user:pass@localhost:1433",
				MaxIdleConnections: 0,
				MaxConnections:     3,
				Timeout:            10 * time.Second,
			},
			err: "max_idle_connection must be at least 1",
		},
		{
			name: "timeout is not positive",
			input: Config{
				ConnectionString:   "sqlserver://user:pass@localhost:1433",
				MaxIdleConnections: 3,
				MaxConnections:     3,
				Timeout:            0,
			},
			err: "timeout must be positive",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.validate()
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.err)
			}
		})
	}
}
func TestConfig_Identifier(t *testing.T) {
	t.Run("Identifier is in common config", func(t *testing.T) {
		c := DefaultConfig

		ik := "mssql-instance-key"
		c.Common.InstanceKey = &ik

		id, err := c.Identifier(integrations_v2.Globals{})
		require.NoError(t, err)
		require.Equal(t, ik, id)
	})

	t.Run("Identifier is not in common config", func(t *testing.T) {
		c := DefaultConfig
		c.ConnectionString = "sqlserver://user:pass@localhost:1433"

		id, err := c.Identifier(integrations_v2.Globals{})
		require.NoError(t, err)
		require.Equal(t, "localhost:1433", id)
	})

	t.Run("Identifier is not in common config, invalid URL", func(t *testing.T) {
		c := DefaultConfig
		c.ConnectionString = "\u0001"

		_, err := c.Identifier(integrations_v2.Globals{})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to parse connection string URL:")
	})
}

func TestConfig_NewIntegration(t *testing.T) {
	t.Run("integration with valid config", func(t *testing.T) {
		c := &Config{
			ConnectionString:   "sqlserver://user:pass@localhost:1433",
			MaxIdleConnections: 3,
			MaxConnections:     3,
			Timeout:            10 * time.Second,
		}

		i, err := c.NewIntegration(log.NewJSONLogger(os.Stdout), integrations_v2.Globals{})
		require.NoError(t, err)
		require.NotNil(t, i)
	})

	t.Run("integration with invalid config", func(t *testing.T) {
		c := &Config{
			ConnectionString:   "\u0001",
			MaxIdleConnections: 3,
			MaxConnections:     3,
			Timeout:            10 * time.Second,
		}

		i, err := c.NewIntegration(log.NewJSONLogger(os.Stdout), integrations_v2.Globals{})
		require.Nil(t, i)
		require.ErrorContains(t, err, "failed to validate config:")
	})
}

func TestConfig_ApplyDefaults(t *testing.T) {
	c := DefaultConfig

	err := c.ApplyDefaults(integrations_v2.Globals{})
	require.NoError(t, err)
}
