// Package vcenter provides an otelcol.receiver.vcenter component.
package vcenter

import (
	"fmt"
	"net/url"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/receiver"
	otel_service "github.com/grafana/agent/service/otel"
	"github.com/grafana/river/rivertypes"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:          "otelcol.receiver.vcenter",
		Args:          Arguments{},
		NeedsServices: []string{otel_service.ServiceName},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := vcenterreceiver.NewFactory()
			return receiver.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.receiver.vcenter component.
type Arguments struct {
	Endpoint string            `river:"endpoint,attr"`
	Username string            `river:"username,attr"`
	Password rivertypes.Secret `river:"password,attr"`

	ScraperControllerArguments otelcol.ScraperControllerArguments `river:",squash"`
	TLS                        otelcol.TLSClientArguments         `river:"tls,block,optional"`

	// DebugMetrics configures component internal metrics. Optional.
	DebugMetrics otelcol.DebugMetricsArguments `river:"debug_metrics,block,optional"`

	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var _ receiver.Arguments = Arguments{}

var (
	// DefaultArguments holds default values for Arguments.
	DefaultArguments = Arguments{
		ScraperControllerArguments: otelcol.DefaultScraperControllerArguments,
	}
)

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Convert implements receiver.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	return &vcenterreceiver.Config{
		TLSClientSetting:          *args.TLS.Convert(),
		ScraperControllerSettings: *args.ScraperControllerArguments.Convert(),
		Endpoint:                  args.Endpoint,
		Username:                  args.Username,
		Password:                  configopaque.String(args.Password),
	}, nil
}

// Validate checks to see if the supplied config will work for the receiver
func (args Arguments) Validate() error {
	res, err := url.Parse(args.Endpoint)
	if err != nil {
		return fmt.Errorf("unable to parse url %s: %w", args.Endpoint, err)
	}

	if res.Scheme != "http" && res.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https")
	}
	return nil
}

// Extensions implements receiver.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return nil
}

// Exporters implements receiver.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// NextConsumers implements receiver.Arguments.
func (args Arguments) NextConsumers() *otelcol.ConsumerArguments {
	return args.Output
}

// DebugMetricsConfig implements receiver.Arguments.
func (args Arguments) DebugMetricsConfig() otelcol.DebugMetricsArguments {
	return args.DebugMetrics
}
