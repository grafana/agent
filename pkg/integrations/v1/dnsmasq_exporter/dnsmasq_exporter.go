// Package dnsmasq_exporter embeds https://github.com/google/dnsmasq_exporter
package dnsmasq_exporter //nolint:golint

import (
	"github.com/go-kit/log"
	"github.com/google/dnsmasq_exporter/collector"
	"github.com/grafana/agent/pkg/integrations/shared"
	"github.com/miekg/dns"
)

// DefaultConfig is the default shared for dnsmasq_exporter.
var DefaultConfig Config = Config{
	DnsmasqAddress: "localhost:53",
	LeasesPath:     "/var/lib/misc/dnsmasq.leases",
}

// Config controls the dnsmasq_exporter integration.
type Config struct {
	// DnsmasqAddress is the address of the dnsmasq server (host:port).
	DnsmasqAddress string `yaml:"dnsmasq_address,omitempty"`

	// Path to the dnsmasq leases file.
	LeasesPath string `yaml:"leases_path,omitempty"`
}

// Name returns the name of the integration that this shared is for.
func (c *Config) Name() string {
	return "dnsmasq_exporter"
}

// InstanceKey returns the address of the dnsmasq server.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return c.DnsmasqAddress, nil
}

// NewIntegration converts this shared into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger) (shared.Integration, error) {
	return New(l, c)
}

// New creates a new dnsmasq_exporter integration. The integration scrapes metrics
// from a dnsmasq server.
func New(log log.Logger, c *Config) (shared.Integration, error) {
	exporter := collector.New(log, &dns.Client{
		SingleInflight: true,
	}, c.DnsmasqAddress, c.LeasesPath)

	return shared.NewCollectorIntegration(c.Name(), shared.WithCollectors(exporter)), nil
}
