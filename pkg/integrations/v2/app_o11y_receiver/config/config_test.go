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
	assert.Equal(t, []string{}, cfg.CORSAllowedOrigins)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, false, cfg.RateLimiting.Enabled)
}

func TestConfig_EnableRateLimitNoRPS(t *testing.T) {
	var cfg AppO11yReceiverConfig
	cb := `
rate_limiting:
  enabled: true`
	err := yaml.Unmarshal([]byte(cb), &cfg)
	assert.Nil(t, err)
	assert.Equal(t, true, cfg.RateLimiting.Enabled)
	assert.Equal(t, 100.0, cfg.RateLimiting.RPS)
	assert.Equal(t, 50, cfg.RateLimiting.Burstiness)
}

func TestConfig_EnableRateLimitRPS(t *testing.T) {
	var cfg AppO11yReceiverConfig
	cb := `
rate_limiting:
  enabled: true
  rps: 142`
	err := yaml.Unmarshal([]byte(cb), &cfg)
	assert.Nil(t, err)
	assert.Equal(t, true, cfg.RateLimiting.Enabled)
	assert.Equal(t, 142.0, cfg.RateLimiting.RPS)
	assert.Equal(t, 50, cfg.RateLimiting.Burstiness)
}
