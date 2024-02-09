// Package spanhostmetrics provides an otelcol.connector.spanhostmetrics component.
package spanhostmetrics

import (
	"fmt"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/connector"
	"github.com/grafana/river"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.connector.spanhostmetrics",
		Args:    Arguments{},
		Exports: otelcol.ConsumerExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := NewFactory()
			return connector.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.connector.spanmetrics component.
type Arguments struct {
	AttributeNames       []string      `river:"attribute_names,attr,optional"`
	MetricsFlushInterval time.Duration `river:"metrics_flush_interval,attr,optional"`

	// Output configures where to send processed data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var (
	_ river.Validator     = (*Arguments)(nil)
	_ river.Defaulter     = (*Arguments)(nil)
	_ connector.Arguments = (*Arguments)(nil)
)

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	AttributeNames:       []string{"k8s.node.name", "host.name"},
	MetricsFlushInterval: 1 * time.Second,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	if len(args.AttributeNames) == 0 {
		return fmt.Errorf("attribute_names must not be empty")
	}
	return nil
}

// Convert implements connector.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {

	return &Config{
		AttributeNames: args.AttributeNames,
	}, nil
}

// Extensions implements connector.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return nil
}

// Exporters implements connector.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// NextConsumers implements connector.Arguments.
func (args Arguments) NextConsumers() *otelcol.ConsumerArguments {
	return args.Output
}

// ConnectorType() int implements connector.Arguments.
func (Arguments) ConnectorType() int {
	return connector.ConnectorTracesToMetrics
}
