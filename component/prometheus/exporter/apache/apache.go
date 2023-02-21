package apache

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/apache_http"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.apache",
		Args:    Config{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "apache"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Config)
	return cfg.Convert().NewIntegration(opts.Logger)
}

// DefaultConfig holds the default settings for the apache exporter
var DefaultConfig = Config{
	ApacheAddr:         "http://localhost/server-status?auto",
	ApacheHostOverride: "",
	ApacheInsecure:     false,
}

// Config controls the apache exporter.
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
