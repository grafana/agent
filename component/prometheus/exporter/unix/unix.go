package unix

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	node_integration "github.com/grafana/agent/pkg/integrations/node_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.exporter.unix",
		Args:      Config{},
		Exports:   exporter.Exports{},
		Singleton: true,
		// set name to node_exporter instead of unix. This is for backward compatibility
		// with cloud integrations, dashboards, and other systems that look for `job:integrations/node_exporter`
		Build: exporter.New(createExporter, "node_exporter"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Config)
	return node_integration.New(opts.Logger, cfg.Convert())
}
