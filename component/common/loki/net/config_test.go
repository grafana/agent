package net

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/pkg/river"
	weaveworks "github.com/weaveworks/common/server"
)

func TestConfig(t *testing.T) {
	type testcase struct {
		raw         string
		errExpected bool
		assert      func(t *testing.T, config weaveworks.Config)
	}
	var cases = map[string]testcase{
		"empty config applies defaults": {
			raw: ``,
			assert: func(t *testing.T, config weaveworks.Config) {
				require.Equal(t, "", config.HTTPListenAddress)
				require.Equal(t, 0, config.HTTPListenPort)
				require.Equal(t, "", config.GRPCListenAddress)
				require.Equal(t, 0, config.GRPCListenPort)
			},
		},
		"overriding defaults": {
			raw: `
			http {
				listen_port = 8080
				listen_address = "0.0.0.0"
			}`,
			assert: func(t *testing.T, config weaveworks.Config) {
				require.Equal(t, 8080, config.HTTPListenPort)
				require.Equal(t, "0.0.0.0", config.HTTPListenAddress)
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cfg := ServerConfig{}
			err := river.Unmarshal([]byte(tc.raw), &cfg)
			require.Equal(t, tc.errExpected, err != nil)
			wConfig := cfg.Convert()
			tc.assert(t, wConfig)
		})
	}
}
