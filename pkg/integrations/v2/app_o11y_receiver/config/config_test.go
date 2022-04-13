package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestConfig_DefaultConfig(t *testing.T) {
	var cfg AppO11yReceiverConfig
	cb := `
test-conf: test-val`
	err := yaml.Unmarshal([]byte(cb), &cfg)
	assert.Nil(t, err)
	assert.Equal(t, []string{}, cfg.Server.CORSAllowedOrigins)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 12347, cfg.Server.Port)
	assert.Equal(t, true, cfg.Server.RateLimiting.Enabled)
}

func TestConfig_EnableRateLimitNoRPS(t *testing.T) {
	var cfg AppO11yReceiverConfig
	cb := `
server:
  rate_limiting:
    enabled: true`
	err := yaml.Unmarshal([]byte(cb), &cfg)
	assert.Nil(t, err)
	assert.Equal(t, true, cfg.Server.RateLimiting.Enabled)
	assert.Equal(t, 100.0, cfg.Server.RateLimiting.RPS)
	assert.Equal(t, 50, cfg.Server.RateLimiting.Burstiness)
}

func TestConfig_EnableRateLimitRPS(t *testing.T) {
	var cfg AppO11yReceiverConfig
	cb := `
server:
  rate_limiting:
    enabled: true
    rps: 142`
	err := yaml.Unmarshal([]byte(cb), &cfg)
	assert.Nil(t, err)
	assert.Equal(t, true, cfg.Server.RateLimiting.Enabled)
	assert.Equal(t, 142.0, cfg.Server.RateLimiting.RPS)
	assert.Equal(t, 50, cfg.Server.RateLimiting.Burstiness)
}
