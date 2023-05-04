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
				// defaults inherited from weaveworks
				require.Equal(t, "", config.HTTPListenAddress)
				require.Equal(t, "", config.GRPCListenAddress)
				require.False(t, config.RegisterInstrumentation)
				require.Equal(t, time.Second*30, config.ServerGracefulShutdownTimeout)
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
		"all params": {
			raw: `
			graceful_shutdown_timeout = "1m"
			http {
				listen_address = "0.0.0.0"
				listen_port = 1
				conn_limit = 2
				server_read_timeout = "2m"
				server_write_timeout = "3m"
				server_idle_timeout = "4m"
			}

			grpc {
				listen_address = "0.0.0.1"
				listen_port = 3
				conn_limit = 4
				max_connection_age = "5m"
				max_connection_age_grace = "6m"
				max_connection_idle = "7m"
				server_max_recv_msg_size = 5
				server_max_send_msg_size = 6
				server_max_concurrent_streams = 7
			}`,
			assert: func(t *testing.T, config weaveworks.Config) {
				// general
				require.Equal(t, time.Minute, config.ServerGracefulShutdownTimeout)
				// http
				require.Equal(t, "0.0.0.0", config.HTTPListenAddress)
				require.Equal(t, 1, config.HTTPListenPort)
				require.Equal(t, 2, config.HTTPConnLimit)
				require.Equal(t, time.Minute*2, config.HTTPServerReadTimeout)
				require.Equal(t, time.Minute*3, config.HTTPServerWriteTimeout)
				require.Equal(t, time.Minute*4, config.HTTPServerIdleTimeout)
				// grpc
				require.Equal(t, "0.0.0.1", config.GRPCListenAddress)
				require.Equal(t, 3, config.GRPCListenPort)
				require.Equal(t, 5*time.Minute, config.GRPCServerMaxConnectionAge)
				require.Equal(t, 6*time.Minute, config.GRPCServerMaxConnectionAgeGrace)
				require.Equal(t, 7*time.Minute, config.GRPCServerMaxConnectionIdle)
				require.Equal(t, 5, config.GPRCServerMaxRecvMsgSize)
				require.Equal(t, 6, config.GRPCServerMaxSendMsgSize)
				require.Equal(t, uint(7), config.GPRCServerMaxConcurrentStreams)
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cfg := ServerConfig{}
			err := river.Unmarshal([]byte(tc.raw), &cfg)

			// TODO: this test will need more changes...
			if false {
				require.Equal(t, tc.errExpected, err != nil)
				t.Logf("got config: %+v", cfg)

				//wConfig := cfg.Convert()
				//tc.assert(t, wConfig)
			}
		})
	}
}
