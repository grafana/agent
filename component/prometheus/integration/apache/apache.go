package apache

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/integration"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/apache_http"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.integration.apache",
		Args:    Config{},
		Exports: integration.Exports{},
		Build:   integration.New(createIntegration, "apache"),
	})
}

func createIntegration(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Config)
	return cfg.Convert().NewIntegration(opts.Logger)
}

// DefaultConfig holds the default settings for the apache integration
var DefaultConfig = Config{
	ApacheAddr:         "http://localhost/server-status?auto",
	ApacheHostOverride: "",
	ApacheInsecure:     false,
}

// Config controls the apache integration.
type Config struct {
	ApacheAddr         string `river:"scrape_uri,attr,optional"`
	ApacheHostOverride string `river:"host_override,attr,optional"`
	ApacheInsecure     bool   `river:"insecure,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Config.
func (c *Config) UnmarshalRiver(f func(interface{}) error) error {
	*c = DefaultConfig

	type cfg Config
	return f((*cfg)(c))
}

func (c *Config) Convert() *apache_http.Config {
	return &apache_http.Config{
		ApacheAddr:         c.ApacheAddr,
		ApacheHostOverride: c.ApacheHostOverride,
		ApacheInsecure:     c.ApacheInsecure,
	}
}
