package node_exporter

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/integration"
	"github.com/grafana/agent/pkg/integrations"
	node_integration "github.com/grafana/agent/pkg/integrations/node_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.integration.node_exporter",
		Args:      Config{},
		Exports:   integration.Exports{},
		Singleton: true,
		Build:     integration.New(createIntegration, "node_exporter"),
	})
}

func createIntegration(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Config)
	return node_integration.New(opts.Logger, cfg.Convert())
}
