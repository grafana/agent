package frontendcollector

import (
	"time"

	"github.com/grafana/agent/pkg/util/server"
)

type Config struct {
	Configs []*InstanceConfig `yaml:"configs,omitempty"`
}

func (c *Config) ApplyDefaults() error {
	for _, ic := range c.Configs {
		err := ic.ApplyDefaults()
		if err != nil {
			return err
		}
	}
	return nil
}

type InstanceConfig struct {
	Name                     string            `yaml:"name,omitempty"`
	Server                   server.Config     `yaml:"server"`
	AllowedOrigins           []string          `yaml:"allowed_origins"`
	RateLimitRPS             int               `yaml:"rate_limit_rps"`
	RateLimitBurst           int               `yaml:"rate_limit_burst"`
	LokiName                 string            `yaml:"loki_name"`
	LokiTimeout              time.Duration     `yaml:"loki_timeout"`
	StaticLabels             map[string]string `yaml:"static_labels"`
	LogToStdout              bool              `yaml:"log_to_stdout"`
	DownloadSourcemaps       bool              `yaml:"download_sourcemaps"`
	DownloadSourcemapTimeout time.Duration     `yaml:"download_sourcemaps_timeout"`
}

func (c *InstanceConfig) ApplyDefaults() error {
	if c.RateLimitRPS == 0 {
		c.RateLimitRPS = 10
	}
	if c.RateLimitBurst == 0 {
		c.RateLimitBurst = 100
	}
	if c.LokiTimeout == 0 {
		c.LokiTimeout = 100 * time.Millisecond
	}
	if c.DownloadSourcemapTimeout == 0 {
		c.DownloadSourcemapTimeout = 1 * time.Second
	}
	return nil
}
