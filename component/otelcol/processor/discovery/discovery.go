// Package discovery provides an otelcol.processor.discovery component.
package discovery

import (
	"context"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/fanoutconsumer"
	"github.com/grafana/agent/component/otelcol/internal/lazyconsumer"
	"github.com/grafana/agent/pkg/river"
	promsdconsumer "github.com/grafana/agent/pkg/traces/promsdprocessor/consumer"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.processor.discovery",
		Args:    Arguments{},
		Exports: otelcol.ConsumerExports{},

		Build: func(o component.Options, a component.Arguments) (component.Component, error) {
			return New(o, a.(Arguments))
		},
	})
}

// Arguments configures the otelcol.processor.discovery component.
type Arguments struct {
	Targets         []discovery.Target `river:"targets,attr"`
	OperationType   string             `river:"operation_type,attr,optional"`
	PodAssociations []string           `river:"pod_associations,attr,optional"`

	// Output configures where to send processed data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var (
	_ river.Defaulter = (*Arguments)(nil)
	_ river.Validator = (*Arguments)(nil)
)

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	OperationType: promsdconsumer.OperationTypeUpsert,
	PodAssociations: []string{
		promsdconsumer.PodAssociationIPLabel,
		promsdconsumer.PodAssociationOTelIPLabel,
		promsdconsumer.PodAssociationk8sIPLabel,
		promsdconsumer.PodAssociationHostnameLabel,
		promsdconsumer.PodAssociationConnectionIP,
	},
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	err := promsdconsumer.ValidateOperationType(args.OperationType)
	if err != nil {
		return err
	}

	err = promsdconsumer.ValidatePodAssociations(args.PodAssociations)
	if err != nil {
		return err
	}

	return nil
}

// Component is the otelcol.exporter.discovery component.
type Component struct {
	cfg      Arguments
	consumer *promsdconsumer.Consumer
	logger   log.Logger
}

var _ component.Component = (*Component)(nil)

// New creates a new otelcol.exporter.discovery component.
func New(o component.Options, c Arguments) (*Component, error) {
	if c.Output.Logs != nil || c.Output.Metrics != nil {
		level.Warn(o.Logger).Log("msg", "non-trace input detected; this component only works for traces")
	}

	nextTraces := fanoutconsumer.Traces(c.Output.Traces)
	consumer, err := promsdconsumer.NewConsumer(nextTraces, c.OperationType, c.PodAssociations, o.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create a traces consumer due to error: %w", err)
	}

	res := &Component{
		consumer: consumer,
		logger:   o.Logger,
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
	for _ = range ctx.Done() {
		return nil
	}
	return nil
}

// Update implements Component.
func (c *Component) Update(newConfig component.Arguments) error {

	cfg := newConfig.(Arguments)
	c.cfg = cfg

	hostLabels := make(map[string]discovery.Target)

	for _, labels := range c.cfg.Targets {
		host, err := promsdconsumer.GetHostFromLabels(labels)
		if err != nil {
			level.Warn(c.logger).Log("msg", "ignoring target, unable to find address", "err", err)
			continue
		}
		promsdconsumer.CleanupLabels(labels)
		hostLabels[host] = labels
	}

	c.consumer.SetHostLabels(hostLabels)

	return nil
}
