package apache

import (
	"net/url"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/apache_http"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.apache",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.NewWithTargetBuilder(createExporter, "apache", customizeTarget),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

func customizeTarget(baseTarget discovery.Target, args component.Arguments) []discovery.Target {
	a := args.(Arguments)
	target := baseTarget

	url, err := url.Parse(a.ApacheAddr)
	if err != nil {
		return []discovery.Target{target}
	}

	target["instance"] = url.Host
	return []discovery.Target{target}
}

// DefaultArguments holds the default settings for the apache exporter
var DefaultArguments = Arguments{
	ApacheAddr:         "http://localhost/server-status?auto",
	ApacheHostOverride: "",
	ApacheInsecure:     false,
	ClusteringEnabled:  false,
}

// Arguments controls the apache exporter.
type Arguments struct {
	ApacheAddr         string `river:"scrape_uri,attr,optional"`
	ApacheHostOverride string `river:"host_override,attr,optional"`
	ApacheInsecure     bool   `river:"insecure,attr,optional"`
	ClusteringEnabled  bool   `river:"clustering_enabled,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Arguments.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type args Arguments
	return f((*args)(a))
}

func (a *Arguments) Convert() *apache_http.Config {
	return &apache_http.Config{
		ApacheAddr:         a.ApacheAddr,
		ApacheHostOverride: a.ApacheHostOverride,
		ApacheInsecure:     a.ApacheInsecure,
	}
}
