package mssql

import (
	"testing"
	"time"

	"github.com/grafana/agent/component/discovery"
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

func TestUnmarshalInvalid(t *testing.T) {
	invalidRiverConfig := `
	connection_string = "sqlserver://user:pass@localhost:1433"
	max_idle_connections = 1
	max_open_connections = 1
	timeout = "-1s"
	`

	var invalidArgs Arguments
	err := river.Unmarshal([]byte(invalidRiverConfig), &invalidArgs)
	require.Error(t, err)
}

func TestArgumentsValidate(t *testing.T) {
	tests := []struct {
		name    string
		args    Arguments
		wantErr bool
	}{
		{
			name: "invalid max open connections",
			args: Arguments{
				ConnectionString:   rivertypes.Secret("test"),
				MaxIdleConnections: 1,
				MaxOpenConnections: 0,
				Timeout:            10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid max idle connections",
			args: Arguments{
				ConnectionString:   rivertypes.Secret("test"),
				MaxIdleConnections: 0,
				MaxOpenConnections: 1,
				Timeout:            10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			args: Arguments{
				ConnectionString:   rivertypes.Secret("test"),
				MaxIdleConnections: 1,
				MaxOpenConnections: 1,
				Timeout:            0,
			},
			wantErr: true,
		},
		{
			name: "valid",
			args: Arguments{
				ConnectionString:   rivertypes.Secret("test"),
				MaxIdleConnections: 1,
				MaxOpenConnections: 1,
				Timeout:            10 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
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

func TestCustomizeTargetValid(t *testing.T) {
	args := Arguments{
		ConnectionString: rivertypes.Secret("sqlserver://user:pass@localhost:1433"),
	}

	baseTarget := discovery.Target{}
	newTargets := customizeTarget(baseTarget, args)
	require.Equal(t, 1, len(newTargets))
	require.Equal(t, "localhost:1433", newTargets[0]["instance"])
}

func TestCustomizeTargetInvalid(t *testing.T) {
	args := Arguments{
		ConnectionString: rivertypes.Secret("bad_cs:pass@localhost:1433"),
	}

	baseTarget := discovery.Target{
		"instance": "default instance",
	}
	newTargets := customizeTarget(baseTarget, args)
	require.Equal(t, 1, len(newTargets))
	require.Equal(t, "default instance", newTargets[0]["instance"])
}
