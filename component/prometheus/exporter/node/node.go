package node_exporter

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	node_integration "github.com/grafana/agent/pkg/integrations/node_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.exporter.node",
		Args:      Config{},
		Exports:   exporter.Exports{},
		Singleton: true,
		Build:     exporter.New(createExporter, "node"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Config)
	return node_integration.New(opts.Logger, cfg.Convert())
}
