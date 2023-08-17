package app_agent_receiver

import (
	internal "github.com/grafana/agent/pkg/integrations/v2/app_agent_receiver"
)

type serverConfig struct {
	host                  string             `river:"host,string,optional"`
	port                  int                `river:"port,number,optional"`
	cORSAllowedOrigins    []string           `river:"cors_allowed_origins,string,optional"`
	rateLimiting          rateLimitingConfig `river:"rate_limiting,block,optional"`
	aPIKey                string             `river:"api_key,string,optional"`
	maxAllowedPayloadSize int64              `river:"max_allowed_payload_size,number,optional"`
}

type rateLimitingConfig struct {
	enabled    bool    `river:"enabled,bool,optional"`
	rPS        float64 `river:"rps,number,optional"`
	burstiness int     `river:"burstiness,number,optional"`
}

func (config *serverConfig) toInternal() internal.ServerConfig {
	return internal.ServerConfig{
		Host:                  config.host,
		Port:                  config.port,
		CORSAllowedOrigins:    config.cORSAllowedOrigins,
		RateLimiting:          config.rateLimiting.toInternal(),
		APIKey:                config.aPIKey,
		MaxAllowedPayloadSize: config.maxAllowedPayloadSize,
	}
}

func (config *rateLimitingConfig) toInternal() internal.RateLimitingConfig {
	return internal.RateLimitingConfig{
		Enabled:    config.enabled,
		RPS:        config.rPS,
		Burstiness: config.burstiness,
	}
}
