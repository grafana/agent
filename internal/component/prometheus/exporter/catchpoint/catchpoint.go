package catchpoint

import (
	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/component/prometheus/exporter"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/grafana/agent/static/integrations"
	"github.com/grafana/agent/static/integrations/catchpoint_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.exporter.catchpoint",
		Stability: featuregate.StabilityStable,
		Args:      Arguments{},
		Exports:   exporter.Exports{},
		Build:     exporter.New(createExporter, "catchpoint"),
	})
}

func createExporter(opts component.Options, args component.Arguments, defaultInstanceKey string) (integrations.Integration, string, error) {
	a := args.(Arguments)
	return integrations.NewIntegrationWithInstanceKey(opts.Logger, a.Convert(), defaultInstanceKey)
}

// DefaultArguments holds the default settings for the catchpoint exporter
var DefaultArguments = Arguments{
	VerboseLogging: false,
	WebhookPath:    "/catchpoint-webhook",
	Port:           "9090",
}

// Arguments controls the catchpoint exporter.
type Arguments struct {
	VerboseLogging bool   `river:"verbose_logging,attr"`
	WebhookPath    string `river:"webhook_path,attr"`
	Port           string `river:"port,attr"`
}

// UnmarshalRiver implements River unmarshalling for Arguments.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type args Arguments
	return f((*args)(a))
}

func (a *Arguments) Convert() *catchpoint_exporter.Config {
	return &catchpoint_exporter.Config{
		VerboseLogging: a.VerboseLogging,
		WebhookPath:    a.WebhookPath,
		Port:           a.Port,
	}
}
