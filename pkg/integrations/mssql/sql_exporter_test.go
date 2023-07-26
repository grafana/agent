package mssql

import (
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

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
				MaxOpenConnections: 3,
				Timeout:            10 * time.Second,
			},
		},
		{
			name: "incorrect connection_string scheme",
			input: Config{
				ConnectionString:   "mysql://user:pass@localhost:1433",
				MaxIdleConnections: 3,
				MaxOpenConnections: 3,
				Timeout:            10 * time.Second,
			},
			err: "scheme of provided connection_string URL must be sqlserver",
		},
		{
			name: "invalid URL",
			input: Config{
				ConnectionString:   "\u0001",
				MaxIdleConnections: 3,
				MaxOpenConnections: 3,
				Timeout:            10 * time.Second,
			},
			err: "failed to parse connection_string",
		},
		{
			name: "missing connection_string",
			input: Config{
				MaxIdleConnections: 3,
				MaxOpenConnections: 3,
				Timeout:            10 * time.Second,
			},
			err: "the connection_string parameter is required",
		},
		{
			name: "max connections is less than 1",
			input: Config{
				ConnectionString:   "sqlserver://user:pass@localhost:1433",
				MaxIdleConnections: 3,
				MaxOpenConnections: 0,
				Timeout:            10 * time.Second,
			},
			err: "max_connections must be at least 1",
		},
		{
			name: "max idle connections is less than 1",
			input: Config{
				ConnectionString:   "sqlserver://user:pass@localhost:1433",
				MaxIdleConnections: 0,
				MaxOpenConnections: 3,
				Timeout:            10 * time.Second,
			},
			err: "max_idle_connection must be at least 1",
		},
		{
			name: "timeout is not positive",
			input: Config{
				ConnectionString:   "sqlserver://user:pass@localhost:1433",
				MaxIdleConnections: 3,
				MaxOpenConnections: 3,
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
func TestConfig_UnmarshalYaml(t *testing.T) {
	t.Run("only required values", func(t *testing.T) {
		strConfig := `connection_string: "sqlserver://user:pass@localhost:1433"`

		var c Config

		require.NoError(t, yaml.UnmarshalStrict([]byte(strConfig), &c))

		require.Equal(t, Config{
			ConnectionString:   "sqlserver://user:pass@localhost:1433",
			MaxIdleConnections: 3,
			MaxOpenConnections: 3,
			Timeout:            10 * time.Second,
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
			ConnectionString:   "sqlserver://user:pass@localhost:1433",
			MaxIdleConnections: 5,
			MaxOpenConnections: 6,
			Timeout:            time.Minute,
		}, c)
	})
}

func TestConfig_NewIntegration(t *testing.T) {
	t.Run("integration with valid config", func(t *testing.T) {
		c := &Config{
			ConnectionString:   "sqlserver://user:pass@localhost:1433",
			MaxIdleConnections: 3,
			MaxOpenConnections: 3,
			Timeout:            10 * time.Second,
		}

		i, err := c.NewIntegration(log.NewJSONLogger(os.Stdout))
		require.NoError(t, err)
		require.NotNil(t, i)
	})

	t.Run("integration with invalid config", func(t *testing.T) {
		c := &Config{
			ConnectionString:   "mysql://user:pass@localhost:1433",
			MaxIdleConnections: 3,
			MaxOpenConnections: 3,
			Timeout:            10 * time.Second,
		}

		i, err := c.NewIntegration(log.NewJSONLogger(os.Stdout))
		require.Nil(t, i)
		require.ErrorContains(t, err, "failed to validate config:")
	})
}

func TestConfig_AgentKey(t *testing.T) {
	t.Run("valid url", func(t *testing.T) {
		c := Config{
			ConnectionString:   "mssql://user:pass@localhost:1433",
			MaxIdleConnections: 3,
			MaxOpenConnections: 3,
			Timeout:            10 * time.Second,
		}

		ik, err := c.InstanceKey("agent-key")

		require.NoError(t, err)
		require.Equal(t, "localhost:1433", ik)
	})

	t.Run("invalid url", func(t *testing.T) {
		c := Config{
			ConnectionString:   "\u0001",
			MaxIdleConnections: 3,
			MaxOpenConnections: 3,
			Timeout:            10 * time.Second,
		}

		_, err := c.InstanceKey("agent-key")

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to parse connection string URL")
	})
}
