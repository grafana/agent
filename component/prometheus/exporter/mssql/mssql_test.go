package mssql

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/integrations/mssql"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/river/rivertypes"
	config_util "github.com/prometheus/common/config"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshal(t *testing.T) {
	riverConfig := `
	connection_string = "sqlserver://user:pass@localhost:1433"
	max_idle_connections = 3
	max_open_connections = 3
	timeout = "10s"
	`

	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	expected := Arguments{
		ConnectionString:   rivertypes.Secret("sqlserver://user:pass@localhost:1433"),
		MaxIdleConnections: 3,
		MaxOpenConnections: 3,
		Timeout:            10 * time.Second,
	}

	require.Equal(t, expected, args)
}

func TestConvert(t *testing.T) {
	riverConfig := `
	connection_string = "sqlserver://user:pass@localhost:1433"
	`
	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	res := args.Convert()

	expected := mssql.Config{
		ConnectionString:   config_util.Secret("sqlserver://user:pass@localhost:1433"),
		MaxIdleConnections: DefaultArguments.MaxIdleConnections,
		MaxOpenConnections: DefaultArguments.MaxOpenConnections,
		Timeout:            DefaultArguments.Timeout,
	}
	require.Equal(t, expected, *res)
}
