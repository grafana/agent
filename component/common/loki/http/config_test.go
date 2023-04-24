package http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_Default(t *testing.T) {
	cfg := ServerConfig{}
	serverConfig := cfg.Convert()

	// assert over defaults specified by agent code
	require.Equal(t, 0, serverConfig.HTTPListenPort)
	require.Equal(t, 0, serverConfig.GRPCListenPort)
}

func TestConfig_OverridingDefaults(t *testing.T) {
	cfg := ServerConfig{
		HTTP: &HTTPConfig{
			ListenAddress: "0.0.0.0",
			ListenPort:    4000,
		},
		GRPC: &GRPCConfig{
			ListenAddress: "0.0.0.0",
			ListenPort:    4001,
		},
	}
	serverConfig := cfg.Convert()

	// assert over defaults specified by agent code
	require.Equal(t, 4000, serverConfig.HTTPListenPort)
	require.Equal(t, "0.0.0.0", serverConfig.HTTPListenAddress)
	require.Equal(t, 4001, serverConfig.GRPCListenPort)
	require.Equal(t, "0.0.0.0", serverConfig.GRPCListenAddress)
}
