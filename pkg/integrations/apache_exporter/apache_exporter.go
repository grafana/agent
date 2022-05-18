// Package apache_exporter embeds https://github.com/Lusitaniae/apache_exporter
package apache_exporter //nolint:golint

import (
	"fmt"
	"net/url"

	ae "github.com/Lusitaniae/apache_exporter/collector"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
)

// DefaultConfig holds the default settings for the apache_exporter integration
var DefaultConfig = Config{
	ApacheAddr:         "http://localhost/server-status?auto",
	ApacheHostOverride: "",
	ApacheInsecure:     false,
}

// Config controls the apache_exporter integration.
type Config struct {
	ApacheAddr         string `yaml:"scrape_uri,omitempty"`
	ApacheHostOverride string `yaml:"host_override,omitempty"`
	ApacheInsecure     bool   `yaml:"insecure,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration this config is for.
func (c *Config) Name() string {
	return "apache_exporter"
}

// InstanceKey returns the addr of the apache server.
func (c *Config) InstanceKey(agentKey string) (string, error) {

	u, err := url.Parse(c.ApacheAddr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", u.Hostname(), u.Port()), nil
}

// NewIntegration converts the config into an integration instance.
func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	return new(logger, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeMultiplex, metricsutils.NewNamedShim("apache"))
}

func new(logger log.Logger, c *Config) (integrations.Integration, error) {
	conf := &ae.Config{
		ScrapeURI:    c.ApacheAddr,
		HostOverride: c.ApacheHostOverride,
		Insecure:     c.ApacheInsecure,
	}

	//check scrape URI
	_, err := url.ParseRequestURI(conf.ScrapeURI)
	if err != nil {
		level.Error(logger).Log("msg", "scrape_uri is invalid", "err", err)
		return nil, err
	}
	aeExporter := ae.NewExporter(logger, conf)

	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(aeExporter),
	), nil
}
