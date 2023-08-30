// Package spanlogs provides an otelcol.connector.spanlogs component.
package spanlogs

import (
	"context"
	"fmt"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/fanoutconsumer"
	"github.com/grafana/agent/component/otelcol/internal/lazyconsumer"
	"github.com/grafana/river"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.connector.spanlogs",
		Args:    Arguments{},
		Exports: otelcol.ConsumerExports{},

		Build: func(o component.Options, a component.Arguments) (component.Component, error) {
			return New(o, a.(Arguments))
		},
	})
}

// Arguments configures the otelcol.connector.spanlogs component.
type Arguments struct {
	Spans             bool           `river:"spans,attr,optional"`
	Roots             bool           `river:"roots,attr,optional"`
	Processes         bool           `river:"processes,attr,optional"`
	SpanAttributes    []string       `river:"span_attributes,attr,optional"`
	ProcessAttributes []string       `river:"process_attributes,attr,optional"`
	Overrides         OverrideConfig `river:"overrides,block,optional"`
	Labels            []string       `river:"labels,attr,optional"`

	// Output configures where to send processed data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

type OverrideConfig struct {
	LogsTag     string `river:"logs_instance_tag,attr,optional"`
	ServiceKey  string `river:"service_key,attr,optional"`
	SpanNameKey string `river:"span_name_key,attr,optional"`
	StatusKey   string `river:"status_key,attr,optional"`
	DurationKey string `river:"duration_key,attr,optional"`
	TraceIDKey  string `river:"trace_id_key,attr,optional"`
}

var (
	_ river.Defaulter = (*Arguments)(nil)
)

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	Overrides: OverrideConfig{
		LogsTag:     "traces",
		ServiceKey:  "svc",
		SpanNameKey: "span",
		StatusKey:   "status",
		DurationKey: "dur",
		TraceIDKey:  "tid",
	},
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Component is the otelcol.exporter.spanlogs component.
type Component struct {
	consumer *consumer
}

var _ component.Component = (*Component)(nil)

// New creates a new otelcol.exporter.spanlogs component.
func New(o component.Options, c Arguments) (*Component, error) {
	if c.Output.Traces != nil || c.Output.Metrics != nil {
		level.Warn(o.Logger).Log("msg", "non-log output detected; this component only works for log outputs and trace inputs")
	}

	nextLogs := fanoutconsumer.Logs(c.Output.Logs)
	consumer, err := NewConsumer(c, nextLogs)
	if err != nil {
		return nil, fmt.Errorf("failed to create a traces consumer due to error: %w", err)
	}

	res := &Component{
		consumer: consumer,
	}

	if err := res.Update(c); err != nil {
		return nil, err
	}

	// Export the consumer.
	// This will remain the same throughout the component's lifetime,
	// so we do this during component construction.
	export := lazyconsumer.New(context.Background())
	export.SetConsumers(res.consumer, nil, nil)
	o.OnStateChange(otelcol.ConsumerExports{Input: export})

	return res, nil
}

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	for range ctx.Done() {
		return nil
	}
	return nil
}

// Update implements Component.
func (c *Component) Update(newConfig component.Arguments) error {
	cfg := newConfig.(Arguments)

	nextLogs := fanoutconsumer.Logs(cfg.Output.Logs)

	err := c.consumer.UpdateOptions(cfg, nextLogs)
	if err != nil {
		return fmt.Errorf("failed to update traces consumer due to error: %w", err)
	}

	return nil
}
