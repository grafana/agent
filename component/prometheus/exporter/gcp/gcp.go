package gcp

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/gcp_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.gcp",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "gcp"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

var DefaultArguments = Arguments(gcp_exporter.DefaultConfig)

type Arguments gcp_exporter.Config

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

// Validate implements river.Validator.
func (a *Arguments) Validate() error {
	if err := a.Convert().Validate(); err != nil {
		return err
	}
	return nil
}

func (a *Arguments) Convert() *gcp_exporter.Config {
	// NOTE(tburgessdev): this works because we set up this exporter's Arguments struct
	// to have the exact same field types as the gcp_exporter.Config struct.
	c := gcp_exporter.Config(*a)
	return &c
}
