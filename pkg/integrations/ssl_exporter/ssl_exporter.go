// Package ssl_exporter embeds https://github.com/ribbybibby/ssl_exporter/v2
package ssl_exporter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	ssl_config "github.com/ribbybibby/ssl_exporter/v2/config"
)

// DefaultConfig holds the default settings for the ssl_exporter integration.
var DefaultConfig = Config{
	ConfigFile: "",
	SSLTargets: []SSLTarget{},
}

// SSLTarget represents a target to scrape.
type SSLTarget struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
	Module string `yaml:"module"`
}

// Config controls the ssl_exporter integration.
type Config struct {
	ConfigFile string      `yaml:"key_file,omitempty"`
	SSLTargets []SSLTarget `yaml:"ssl_targets"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration.
func (c *Config) Name() string {
	return "ssl"
}

// InstanceKey returns the hostname:port of the agent.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

// NewIntegration converts the config into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// New creates a new ssl_exporter integration. The integration scrapes
// metrics from ssl certificates
func New(log log.Logger, c *Config) (integrations.Integration, error) {
	var modules *ssl_config.Config
	var err error

	modules = ssl_config.DefaultConfig
	if c.ConfigFile != "" {
		modules, err = ssl_config.LoadConfig(c.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load ssl config from file %v: %w", c.ConfigFile, err)
		}
	}

	// The `name` and `target` fields are mandatory for the ssl targets are mandatory.
	// Enforce this check and fail the creation of the integration if they're missing.
	for _, target := range c.SSLTargets {
		if target.Name == "" || target.Target == "" {
			return nil, fmt.Errorf("failed to load ssl_targets; the `name` and `target` fields are mandatory")
		}
	}

	sh := &sslHandler{
		cfg:     c,
		modules: modules,
		log:     log,
	}
	integration := &Integration{
		sh: sh,
	}

	return integration, nil
}

// Integration is an integration for ssl_exporter.
type Integration struct {
	sh *sslHandler
}

// MetricsHandler implements Integration.
func (i *Integration) MetricsHandler() (http.Handler, error) {
	return i.sh, nil
}

// Run satisfies Integration.Run.
func (i *Integration) Run(ctx context.Context) error {
	// We don't need to do anything here, so we can just wait for the context to
	// finish.
	<-ctx.Done()
	return ctx.Err()
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (i *Integration) ScrapeConfigs() []config.ScrapeConfig {
	var res []config.ScrapeConfig
	for _, target := range i.sh.cfg.SSLTargets {
		queryParams := url.Values{}
		queryParams.Add("target", target.Target)
		queryParams.Add("module", target.Module)
		res = append(res, config.ScrapeConfig{
			JobName:     i.sh.cfg.Name() + "/" + target.Name,
			MetricsPath: "/metrics",
			QueryParams: queryParams,
		})
	}
	return res
}
