package catchpoint

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/catchpoint_exporter"
	"github.com/grafana/agent/pkg/river/rivertypes"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.catchpoint",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "catchpoint"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

// DefaultArguments holds the default settings for the catchpoint exporter
var DefaultArguments = Arguments{
  Verbose: false,
  WebhookPath: "/catchpoint-webhook",
  Port: "9090",
}

// Arguments controls the catchpoint exporter.
type Arguments struct {
  Verbose bool `river:"verbose,attr"`
  WebhookPath string `river:"webhookpath,attr"`
  Port string `river:"port,attr"`

}

// UnmarshalRiver implements River unmarshalling for Arguments.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type args Arguments
	return f((*args)(a))
}

func (a *Arguments) Convert() *catchpoint_exporter.Config {
	return &catchpoint_exporter.Config{
    Verbose: a.Verbose,
    WebhookPath: a.WebhookPath,
    Port: a.Port,
	}
}

