package net

import (
	"testing"
	"time"

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
				// custom defaults
				require.Equal(t, DefaultHTTPPort, config.HTTPListenPort)
				require.Equal(t, DefaultGRPCPort, config.GRPCListenPort)
				// defaults inherited from weaveworks
				require.Equal(t, "", config.HTTPListenAddress)
				require.Equal(t, "", config.GRPCListenAddress)
				require.False(t, config.RegisterInstrumentation)
			},
		},
		"overriding defaults": {
			raw: `
			graceful_shutdown_timeout = "1m"
			http {
				listen_port = 8080
				listen_address = "0.0.0.0"
				conn_limit = 10
				server_write_timeout = "10s"
			}`,
			assert: func(t *testing.T, config weaveworks.Config) {
				require.Equal(t, 8080, config.HTTPListenPort)
				require.Equal(t, "0.0.0.0", config.HTTPListenAddress)
				require.Equal(t, 10, config.HTTPConnLimit)
				require.Equal(t, time.Second*10, config.HTTPServerWriteTimeout)

				require.Equal(t, time.Minute, config.ServerGracefulShutdownTimeout)
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
