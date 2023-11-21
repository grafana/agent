package mssql

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/integrations/mssql"
	"github.com/grafana/river"
	"github.com/grafana/river/rivertypes"
	config_util "github.com/prometheus/common/config"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshal(t *testing.T) {
	goodQueryPath, _ := filepath.Abs("../../../../pkg/integrations/mssql/collector_config.yaml")

	riverConfig := `
	connection_string = "sqlserver://user:pass@localhost:1433"
	max_idle_connections = 3
	max_open_connections = 3
	timeout = "10s"
    query_config_file = "` + goodQueryPath + `"`

	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	expected := Arguments{
		ConnectionString:   rivertypes.Secret("sqlserver://user:pass@localhost:1433"),
		MaxIdleConnections: 3,
		MaxOpenConnections: 3,
		Timeout:            10 * time.Second,
		QueryConfigFile:    goodQueryPath,
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
	goodQueryPath, _ := filepath.Abs("../../../../pkg/integrations/mssql/collector_config.yaml")

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
				QueryConfigFile:    goodQueryPath,
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
				QueryConfigFile:    goodQueryPath,
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
				QueryConfigFile:    goodQueryPath,
			},
			wantErr: true,
		},
		{
			name: "invalid query_config_file",
			args: Arguments{
				ConnectionString:   rivertypes.Secret("test"),
				MaxIdleConnections: 1,
				MaxOpenConnections: 1,
				Timeout:            0,
				QueryConfigFile:    "doesnotexist.YAML",
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
				QueryConfigFile:    goodQueryPath,
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
	goodQueryPath, _ := filepath.Abs("../../../../pkg/integrations/mssql/collector_config.yaml")
	riverConfig := `
	connection_string = "sqlserver://user:pass@localhost:1433"
	query_config_file = "` + goodQueryPath + `"`
	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	res := args.Convert()

	expected := mssql.Config{
		ConnectionString:   config_util.Secret("sqlserver://user:pass@localhost:1433"),
		MaxIdleConnections: DefaultArguments.MaxIdleConnections,
		MaxOpenConnections: DefaultArguments.MaxOpenConnections,
		Timeout:            DefaultArguments.Timeout,
		QueryConfigFile:    goodQueryPath,
	}
	require.Equal(t, expected, *res)
}
