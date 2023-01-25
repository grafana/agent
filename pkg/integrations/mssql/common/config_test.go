package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
			err := tc.input.Validate()
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.err)
			}
		})
	}
}
