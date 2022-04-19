package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestConfig_DefaultConfig(t *testing.T) {
	var cfg AppO11yReceiverConfig
	cb := `
test-conf: test-val`
	err := yaml.Unmarshal([]byte(cb), &cfg)
	require.NoError(t, err)
	require.Equal(t, []string(nil), cfg.Server.CORSAllowedOrigins)
	require.Equal(t, "127.0.0.1", cfg.Server.Host)
	require.Equal(t, 12347, cfg.Server.Port)
	require.Equal(t, true, cfg.Server.RateLimiting.Enabled)
}

func TestConfig_EnableRateLimitNoRPS(t *testing.T) {
	var cfg AppO11yReceiverConfig
	cb := `
server:
  rate_limiting:
    enabled: true`
	err := yaml.Unmarshal([]byte(cb), &cfg)
	require.NoError(t, err)
	require.Equal(t, true, cfg.Server.RateLimiting.Enabled)
	require.Equal(t, 100.0, cfg.Server.RateLimiting.RPS)
	require.Equal(t, 50, cfg.Server.RateLimiting.Burstiness)
}

func TestConfig_EnableRateLimitRPS(t *testing.T) {
	var cfg AppO11yReceiverConfig
	cb := `
server:
  rate_limiting:
    enabled: true
    rps: 142`
	err := yaml.Unmarshal([]byte(cb), &cfg)
	require.NoError(t, err)
	require.Equal(t, true, cfg.Server.RateLimiting.Enabled)
	require.Equal(t, 142.0, cfg.Server.RateLimiting.RPS)
	require.Equal(t, 50, cfg.Server.RateLimiting.Burstiness)
}
