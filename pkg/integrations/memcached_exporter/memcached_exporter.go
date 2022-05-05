// Package memcached_exporter embeds https://github.com/google/memcached_exporter
package memcached_exporter //nolint:golint

import (
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	"github.com/prometheus/memcached_exporter/pkg/exporter"
)

// DefaultConfig is the default config for memcached_exporter.
var DefaultConfig = Config{
	MemcachedAddress: "localhost:11211",
	Timeout:          time.Second,
}

// Config controls the memcached_exporter integration.
type Config struct {
	// MemcachedAddress is the address of the memcached server (host:port).
	MemcachedAddress string `yaml:"memcached_address,omitempty"`

	// Timeout is the connection timeout for memcached.
	Timeout time.Duration `yaml:"timeout,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "memcached_exporter"
}

// InstanceKey returns the address:port of the memcached server.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return c.MemcachedAddress, nil
}

// NewIntegration converts this config into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeMultiplex, metricsutils.NewNamedShim("memcached"))
}

// New creates a new memcached_exporter integration. The integration scrapes metrics
// from a memcached server.
func New(log log.Logger, c *Config) (integrations.Integration, error) {
	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(
			exporter.New(c.MemcachedAddress, c.Timeout, log),
		),
	), nil
}
