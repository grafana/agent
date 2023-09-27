package apache

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/apache_http"
)

func init() {
	component.Register(component.Registration{
		Name:          "prometheus.exporter.apache",
		Args:          Arguments{},
		Exports:       exporter.Exports{},
		NeedsServices: exporter.RequiredServices(),
		Build:         exporter.New(createExporter, "apache"),
	})
}

func createExporter(opts component.Options, args component.Arguments, defaultInstanceKey string) (integrations.Integration, string, error) {
	a := args.(Arguments)
	return integrations.NewIntegrationWithInstanceKey(opts.Logger, a.Convert(), defaultInstanceKey)
}

// DefaultArguments holds the default settings for the apache exporter
var DefaultArguments = Arguments{
	ApacheAddr:         "http://localhost/server-status?auto",
	ApacheHostOverride: "",
	ApacheInsecure:     false,
}

// Arguments controls the apache exporter.
type Arguments struct {
	ApacheAddr         string `river:"scrape_uri,attr,optional"`
	ApacheHostOverride string `river:"host_override,attr,optional"`
	ApacheInsecure     bool   `river:"insecure,attr,optional"`
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

func (a *Arguments) Convert() *apache_http.Config {
	return &apache_http.Config{
		ApacheAddr:         a.ApacheAddr,
		ApacheHostOverride: a.ApacheHostOverride,
		ApacheInsecure:     a.ApacheInsecure,
	}
}
