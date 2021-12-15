package integrations

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/integrations/config"
)

// StubConfig represents a base no-op configuration, which returns a StubIntegration
type StubConfig struct {
	name   string
	reason string
	Common config.Common          `yaml:",inline"`
	_      map[string]interface{} `yaml:"-"`
}

// UnmarshalYAML implements yaml.Unmarshaler for StubConfig
func (c *StubConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = StubConfig{}
	type plain StubConfig
	err := unmarshal((*plain)(c))
	return err
}

// Name returns the name of the integration that this config represents.
func (c *StubConfig) Name() string {
	return c.name
}

// CommonConfig returns the common settings shared across all configs for
// integrations.
func (c *StubConfig) CommonConfig() config.Common {
	return config.Common{}
}

// InstanceKey returns the agentKey passed to it
func (c *StubConfig) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

// NewIntegration creates a new StubIntegration, logging the StubConfig.reason
func (c *StubConfig) NewIntegration(l log.Logger) (Integration, error) {
	level.Warn(l).Log("msg", c.reason)
	return &StubIntegration{}, nil
}

// NewStubConfig creates a new stub config of the given name. NewIntegration will simply log the provided reason as a warning.
func NewStubConfig(name string, reason string) *StubConfig {
	return &StubConfig{name: name}
}

// StubIntegration implements a no-op integration for use on platforms not supported by an integration
type StubIntegration struct {
}

// MetricsHandler returns an http.NotFoundHandler to satisfy the Integration interface
func (i *StubIntegration) MetricsHandler() (http.Handler, error) {
	return http.NotFoundHandler(), nil
}

// ScrapeConfigs returns an empty list of scrape configs, since there is nothing to scrape
func (i *StubIntegration) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{}
}

// Run just waits for the context to finish
func (i *StubIntegration) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
