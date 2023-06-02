package squid_exporter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

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
				return c
			},
		},
		{
			name: "no hostname",
			getConfig: func() Config {
				c := DefaultConfig
				c.Address = ":3128"
				return c
			},
			expectedErr: errNoHostname,
		},
		{
			name: "no empty config",
			getConfig: func() Config {
				cfg := Config{
					Address:  "",
					Login:    "",
					Password: "",
				}
				return cfg
			},
			expectedErr: errNoConnectionString,
		},
		{
			name: "invalid scheme - cockroachdb",
			getConfig: func() Config {
				c := DefaultConfig
				c.ConnectionString = "postgres://maxroach@localhost:26257/movr?password=pwd"
				return c
			},
			expectedErr: errors.New("unexpected scheme of type 'postgres'. Was expecting 'oracle'"),
		},
		{
			name: "invalid connection string",
			getConfig: func() Config {
				c := DefaultConfig
				c.ConnectionString = "localhost:1521"
				return c
			},
			expectedErr: errors.New("unexpected scheme of type 'localhost'. Was expecting 'oracle'"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.getConfig()

			if tc.expectedErr == nil {
				require.NoError(t, New(cfg))
				return
			}
			require.ErrorContains(t, New(cfg), tc.expectedErr.Error())
		})
	}
}

func TestConfig_InstanceKey(t *testing.T) {
	c := DefaultConfig
	c.Address = "localhost:3128"

	ik := "agent-key"
	id, err := c.InstanceKey(ik)
	require.NoError(t, err)
	require.Equal(t, "localhost:3128", id)
}
