package squid

import (
	"errors"
	"testing"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/integrations/squid_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/prometheus/common/config"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshal(t *testing.T) {
	riverConfig := `
	address = "some_address:port"
	username     = "some_user"
	password     = "some_password"
	`

	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	expected := Arguments{
		SquidAddr:     "some_address:port",
		SquidUser:     "some_user",
		SquidPassword: rivertypes.Secret("some_password"),
	}

	require.Equal(t, expected, args)
}

func TestConvert(t *testing.T) {
	riverConfig := `
	address = "some_address:port"
	username     = "some_user"
	password     = "some_password"
	`
	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	res := args.Convert()

	expected := squid_exporter.Config{
		Address:  "some_address:port",
		Username: "some_user",
		Password: config.Secret("some_password"),
	}
	require.Equal(t, expected, *res)
}

func TestCustomizeTarget(t *testing.T) {
	args := Arguments{
		SquidAddr: "some_address",
	}

	baseTarget := discovery.Target{}
	newTargets := customizeTarget(baseTarget, args)
	require.Equal(t, 1, len(newTargets))
	require.Equal(t, "some_address", newTargets[0]["instance"])
}

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name        string
		getConfig   func() Arguments
		expectedErr error
	}{
		{
			name: "valid",
			getConfig: func() Arguments {
				c := Arguments{
					SquidAddr: "localhost:3128",
				}
				return c
			},
		},
		{
			name: "no hostname",
			getConfig: func() Arguments {
				c := Arguments{}
				c.SquidAddr = ":3128"
				return c
			},
			expectedErr: errNoHostname,
		},
		{
			name: "no port",
			getConfig: func() Arguments {
				c := Arguments{}
				c.SquidAddr = "localhost:"
				return c
			},
			expectedErr: errNoPort,
		},
		{
			name: "no empty config",
			getConfig: func() Arguments {
				cfg := Arguments{
					SquidAddr:     "",
					SquidUser:     "",
					SquidPassword: "",
				}
				return cfg
			},
			expectedErr: errNoAddress,
		},
		{
			name: "invalid config",
			getConfig: func() Arguments {
				cfg := Arguments{}
				cfg.SquidAddr = "a@#$%:asdf::12312"
				return cfg
			},
			expectedErr: errors.New("address a@#$%:asdf::12312: too many colons in address"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.getConfig()
			err := cfg.Validate()
			if tc.expectedErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.expectedErr.Error())
		})
	}
}
