package dnsmasq_exporter //nolint:golint

import (
	"github.com/grafana/agent/pkg/integrations/config"
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

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}
