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
	assert.Equal(t, cfg.CORSAllowedOrigins, []string{"http://localhost:1234"})
	assert.Equal(t, cfg.Server.Host, "0.0.0.0")
	assert.Equal(t, cfg.Server.Port, 8080)
	assert.Equal(t, cfg.RateLimiting.Enabled, false)
}

func TestConfig_EnableRateLimitNoRPS(t *testing.T) {
	var cfg AppO11yReceiverConfig
	cb := `
rate_limiting:
  enabled: true`
	err := yaml.Unmarshal([]byte(cb), &cfg)
	assert.Nil(t, err)
	assert.Equal(t, cfg.RateLimiting.Enabled, true)
	assert.Equal(t, cfg.RateLimiting.RPS, 100.0)
	assert.Equal(t, cfg.RateLimiting.Burstiness, 50)
}

func TestConfig_EnableRateLimitRPS(t *testing.T) {
	var cfg AppO11yReceiverConfig
	cb := `
rate_limiting:
  enabled: true
  rps: 142`
	err := yaml.Unmarshal([]byte(cb), &cfg)
	assert.Nil(t, err)
	assert.Equal(t, cfg.RateLimiting.Enabled, true)
	assert.Equal(t, cfg.RateLimiting.RPS, 142.0)
	assert.Equal(t, cfg.RateLimiting.Burstiness, 50)
}
