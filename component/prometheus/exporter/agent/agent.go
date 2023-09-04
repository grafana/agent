package agent

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/agent"
)

func init() {
	component.Register(component.Registration{
		Name:          "prometheus.exporter.agent",
		Args:          Arguments{},
		Exports:       exporter.Exports{},
		Singleton:     true,
		NeedsServices: exporter.RequiredServices(),
		Build:         exporter.New(createExporter, "agent"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

// Arguments holds values which are used to configured the prometheus.exporter.agent component.
type Arguments struct{}

// Exports holds the values exported by the prometheus.exporter.agent component.
type Exports struct{}

// DefaultArguments defines the default settings
var DefaultArguments = Arguments{}

// SetToDefault implements river.Defaulter
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

func (a *Arguments) Convert() *agent.Config {
	return &agent.Config{}
}
