package unix

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	node_integration "github.com/grafana/agent/pkg/integrations/node_exporter"
	"github.com/grafana/agent/service/http"
)

func init() {
	component.Register(component.Registration{
		Name:          "prometheus.exporter.unix",
		Args:          Arguments{},
		Exports:       exporter.Exports{},
		Singleton:     true,
		NeedsServices: []string{http.ServiceName},
		Build:         exporter.New(createExporter, "unix"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return node_integration.New(opts.Logger, a.Convert())
}
