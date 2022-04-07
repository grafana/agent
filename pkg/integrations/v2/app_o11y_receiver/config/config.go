package config

import "time"

const (
	// DefaultRateLimitingRPS is the default value of Requests Per Second
	// for ratelimiting
	DefaultRateLimitingRPS = 100
	// DefaultRateLimitingBurstiness is the default burstiness factor of the
	// token bucket algorigthm
	DefaultRateLimitingBurstiness = 50
	// DefaultMaxPayloadSize is the max paylad size in bytes
	DefaultMaxPayloadSize = 5e6
)

// DefaultConfig holds the default configuration of the receiver
var DefaultConfig = AppO11yReceiverConfig{
	// Default JS agent port
	CORSAllowedOrigins: []string{},
	RateLimiting: RateLimitingConfig{
		Enabled:    false,
		RPS:        DefaultRateLimitingRPS,
		Burstiness: DefaultRateLimitingBurstiness,
	},
	MaxAllowedPayloadSize: DefaultRateLimitingRPS,
	Server: ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
	},
	TracesInstance:  "",
	LogsInstance:    "",
	LogsLabels:      map[string]string{},
	LogsSendTimeout: 2000,
	SourceMaps: SourceMapConfig{
		Download:            false,
		DownloadFromOrigins: []string{"*"},
		DownloadTimeout:     time.Duration(1000000),
		FileSystem:          nil,
	},
}

// ServerConfig holds the receiver http server configuration
type ServerConfig struct {
	Host string `yaml:"host,omitempty"`
	Port int    `yaml:"port,omitempty"`
}

// RateLimitingConfig holds the configuration of the rate limiter
type RateLimitingConfig struct {
	Enabled    bool    `yaml:"enabled,omitempty"`
	RPS        float64 `yaml:"rps,omitempty"`
	Burstiness int     `yaml:"burstiness,omitempty"`
}

// SourceMapFileLocation holds sourcemap location on file system
type SourceMapFileLocation struct {
	Path               string `yaml:"path"`
	MinifiedPathPrefix string `yaml:"minified_path_prefix,omitempty"`
}

// SourceMapConfig configure source map locations
type SourceMapConfig struct {
	Download            bool                    `yaml:"download"`
	DownloadFromOrigins []string                `yaml:"download_origins,omitempty"`
	DownloadTimeout     time.Duration           `yaml:"download_timeout,omitempty"`
	FileSystem          []SourceMapFileLocation `yaml:"filesystem,omitempty"`
}

// AppO11yReceiverConfig is the configuration struct of the
// integration
type AppO11yReceiverConfig struct {
	CORSAllowedOrigins    []string           `yaml:"cors_allowed_origins,omitempty"`
	RateLimiting          RateLimitingConfig `yaml:"rate_limiting,omitempty"`
	APIKey                string             `yaml:"api_key,omitempty"`
	MaxAllowedPayloadSize int64              `yaml:"max_allowed_payload_size,omitempty"`
	Server                ServerConfig       `yaml:"server,omitempty"`
	TracesInstance        string             `yaml:"traces_instance,omitempty"`
	LogsInstance          string             `yaml:"logs_instance,omitempty"`
	LogsLabels            map[string]string  `yaml:"logs_labels,omitempty"`
	LogsSendTimeout       int                `yaml:"logs_send_timeout,omitempty"`
	SourceMaps            SourceMapConfig    `yaml:"sourcemaps,omitempty"`
}

// UnmarshalYAML implements the Unmarshaller interface
func (c *AppO11yReceiverConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig
	type cA AppO11yReceiverConfig

	if err := unmarshal((*cA)(c)); err != nil {
		return err
	}

	if c.RateLimiting.Enabled && c.RateLimiting.RPS == 0 {
		c.RateLimiting.RPS = DefaultRateLimitingRPS
	}

	return nil
}
