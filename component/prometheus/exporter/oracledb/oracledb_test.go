package oracledb

import (
	"testing"

	"github.com/grafana/agent/pkg/integrations/oracledb_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/river/rivertypes"
	config_util "github.com/prometheus/common/config"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshal(t *testing.T) {
	riverConfig := `
	connection_string  = "oracle://user:password@localhost:1521/orcl.localnet"
	max_idle_conns     = 0
	max_open_conns     = 10
	query_timeout      = 5
	`

	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	expected := Arguments{
		ConnectionString: rivertypes.Secret("oracle://user:password@localhost:1521/orcl.localnet"),
		MaxIdleConns:     0,
		MaxOpenConns:     10,
		QueryTimeout:     5,
	}

	require.Equal(t, expected, args)
}

func TestConvert(t *testing.T) {
	riverConfig := `
	connection_string  = "oracle://user:password@localhost:1521/orcl.localnet"
	`
	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	res := args.Convert()

	expected := oracledb_exporter.Config{
		ConnectionString: config_util.Secret("oracle://user:password@localhost:1521/orcl.localnet"),
		MaxIdleConns:     DefaultArguments.MaxIdleConns,
		MaxOpenConns:     DefaultArguments.MaxOpenConns,
		QueryTimeout:     DefaultArguments.QueryTimeout,
	}
	require.Equal(t, expected, *res)
}
