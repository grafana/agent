package statsd

import (
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.statsd",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "statsd"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Arguments)
	statsdConfig, err := cfg.Convert()
	if err != nil {
		return nil, fmt.Errorf("failed to create statsd exporter: %w", err)
	}
	return statsdConfig.NewIntegration(opts.Logger)
}
