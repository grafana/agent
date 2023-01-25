package oracledbexporter

import (
	"testing"
	"time"

	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestOracleDBConfigUnmarshal(t *testing.T) {
	strConfig := `
connection_string: oracle://user:password@localhost:1521/orcl.localnet
instance: my-oracledb
metrics_scrape_interval: 1m
max_idle_connections: 0
max_open_connections: 10
query_timeout: 5`

	var c Config
	require.NoError(t, yaml.UnmarshalStrict([]byte(strConfig), &c))

	ik := "my-oracledb"
	require.Equal(t, Config{
		ConnectionString:      "oracle://user:password@localhost:1521/orcl.localnet",
		MaxIdleConns:          0,
		MaxOpenConns:          10,
		MetricsScrapeInterval: 1 * time.Minute,
		QueryTimeout:          5,
		Common: common.MetricsConfig{
			InstanceKey: &ik,
		},
	}, c)
}

func TestConfig_Identifier(t *testing.T) {
	t.Run("Identifier is in common config", func(t *testing.T) {
		c := DefaultConfig

		ik := "my-oracledb-instance-key"
		c.Common.InstanceKey = &ik

		id, err := c.Identifier(integrations_v2.Globals{})
		require.NoError(t, err)
		require.Equal(t, ik, id)
	})

	t.Run("Identifier is not in common config", func(t *testing.T) {
		c := DefaultConfig
		c.ConnectionString = "oracle://user:password@localhost:1521/orcl.localnet"

		id, err := c.Identifier(integrations_v2.Globals{})
		require.NoError(t, err)
		require.Equal(t, "localhost:1521", id)
	})
}
