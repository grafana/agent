package windows

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.windows",
		Args:    Arguments{},
		Exports: exporter.Exports{},

		Build: exporter.New(createExporter, "windows"),
	})
}

func createExporter(opts component.Options, args component.Arguments, defaultInstanceKey string) (integrations.Integration, string, error) {
	a := args.(Arguments)
	return integrations.NewIntegrationWithInstanceKey(opts.Logger, a.Convert(), defaultInstanceKey)
}
