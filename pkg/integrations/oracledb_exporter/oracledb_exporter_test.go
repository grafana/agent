package oracledbexporter

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestOracleDBConfig(t *testing.T) {
	strConfig := `
enabled: true
connection_string: oracle://user:password@localhost:1521/orcl.localnet
scrape_interval: "1m"
scrape_timeout: "1m"
scrape_integration: true
max_idle_connections: 0
max_open_connections: 10
query_timeout: 5`

	var c Config
	require.NoError(t, yaml.Unmarshal([]byte(strConfig), &c))

	require.Equal(t, Config{
		ConnectionString: "oracle://user:password@localhost:1521/orcl.localnet",
		MaxIdleConns:     0,
		MaxOpenConns:     10,
		ScrapeInterval:   1 * time.Minute,
		QueryTimeout:     5,
	}, c)
}

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name        string
		getConfig   func() Config
		expectedErr error
	}{
		{
			name: "valid",
			getConfig: func() Config {
				c := DefaultConfig
				c.ConnectionString = "oracle://user:password@localhost:1521/orcl.localnet"
				return c
			},
		},
		{
			name: "no connection string",
			getConfig: func() Config {
				return DefaultConfig
			},
			expectedErr: errNoConnectionString,
		},
		{
			name: "invalid conneciton string",
			getConfig: func() Config {
				c := DefaultConfig
				c.ConnectionString = "localhost:1521"
				return c
			},
			expectedErr: errors.New("unable to parse connection string"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.getConfig()
			if tc.expectedErr == nil {
				require.NoError(t, cfg.Validate())
				return
			}
			require.ErrorContains(t, cfg.Validate(), tc.expectedErr.Error())
		})
	}
}

func TestConfig_InstanceKey(t *testing.T) {
	c := DefaultConfig
	c.ConnectionString = "oracle://user:password@localhost:1521/orcl.localnet"

	ik := "agent-key"
	id, err := c.InstanceKey(ik)
	require.NoError(t, err)
	require.Equal(t, "localhost:1521", id)
}
