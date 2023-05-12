package windows

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/windows_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.exporter.windows",
		Args:      Arguments{},
		Exports:   exporter.Exports{},
		Singleton: false,
		Build:     exporter.New(createExporter, "windows", ""),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return windows_exporter.New(opts.Logger, a.Convert())
}
