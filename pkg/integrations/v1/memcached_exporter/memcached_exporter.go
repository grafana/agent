// Package memcached_exporter embeds https://github.com/google/memcached_exporter
package memcached_exporter //nolint:golint

import (
	"time"

	"github.com/grafana/agent/pkg/integrations/shared"

	"github.com/go-kit/log"
	"github.com/prometheus/memcached_exporter/pkg/exporter"
)

// DefaultConfig is the default shared for memcached_exporter.
var DefaultConfig Config = Config{
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

// Name returns the name of the integration that this shared represents.
func (c *Config) Name() string {
	return "memcached_exporter"
}

// InstanceKey returns the address:port of the memcached server.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return c.MemcachedAddress, nil
}

// NewIntegration converts this shared into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger) (shared.Integration, error) {
	return New(l, c)
}

// New creates a new memcached_exporter integration. The integration scrapes metrics
// from a memcached server.
func New(log log.Logger, c *Config) (shared.Integration, error) {
	return shared.NewCollectorIntegration(
		c.Name(),
		shared.WithCollectors(
			exporter.New(c.MemcachedAddress, c.Timeout, log),
		),
	), nil
}
