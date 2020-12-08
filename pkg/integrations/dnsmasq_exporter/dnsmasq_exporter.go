// Package dnsmasq_exporter embeds https://github.com/google/dnsmasq_exporter
package dnsmasq_exporter //nolint:golint

import (
	"github.com/go-kit/kit/log"
	"github.com/google/dnsmasq_exporter/collector"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/common"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/miekg/dns"
)

// DefaultConfig is the default config for dnsmasq_exporter.
var DefaultConfig Config = Config{
	DnsmasqAddress: "localhost:53",
	LeasesPath:     "/var/lib/misc/dnsmasq.leases",
}

// Config controls the dnsmasq_exporter integration.
type Config struct {
	// Enabled enables the integration.
	Enabled bool `yaml:"enabled"`

	CommonConfig config.Common `yaml:",inline"`

	// DnsmasqAddress is the address of the dnsmasq server (host:port).
	DnsmasqAddress string `yaml:"dnsmasq_address"`

	// Path to the dnsmasq leases file.
	LeasesPath string `yaml:"leases_path"`
}

func (c *Config) Name() string { return "dnsmasq_exporter" }

func (c *Config) IsEnabled() bool { return c.Enabled }

func (c *Config) NewIntegration(l log.Logger) (common.Integration, error) {
	return New(l, c)
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// New creates a new dnsmasq_exporter integration. The integration scrapes metrics
// from a dnsmasq server.
func New(log log.Logger, c *Config) (common.Integration, error) {
	exporter := collector.New(&dns.Client{
		SingleInflight: true,
	}, c.DnsmasqAddress, c.LeasesPath)

	return common.NewCollectorIntegration(
		"dnsmasq_exporter",
		c.CommonConfig,
		exporter,
		false,
	), nil
}
