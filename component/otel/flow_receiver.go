package otel

import (
	"context"
	"errors"
	"os"

	otelcomponent "go.opentelemetry.io/collector/component"
	otelcomponenterror "go.opentelemetry.io/collector/component/componenterror"
	otelconfig "go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otel/internal/zapadapter"
	"github.com/grafana/agent/pkg/build"
)

// FlowReceiver is a Flow component implementation which manages OpenTelemetry
// Collector receivers.
type FlowReceiver struct {
	ctx    context.Context
	cancel context.CancelFunc

	opts    component.Options
	factory otelcomponent.ReceiverFactory

	runner *componentRunner
}

type ReceiverArguments interface {
	component.Arguments

	// Convert should convert the ReceiverArguments into a Receiver confguration
	// used by OpenTelemetry Collector.
	Convert() otelconfig.Receiver

	// Extensions should return the set of extensions that this component is
	// allowed to use.
	Extensions() map[otelconfig.ComponentID]otelcomponent.Extension

	// Exporters should return the set of exporters that are exposed to this
	// component.
	Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter

	// NextArguments returns the set of components to send data to.
	NextArguments() *NextReceiverArguments
}

// NewFlowReceiver creates a new Flow component which encapsules an
// OpenTelemetry Collector receiver. args must be the argument type registered
// with the component.
func NewFlowReceiver(opts component.Options, f otelcomponent.ReceiverFactory, args ReceiverArguments) (*FlowReceiver, error) {
	ctx, cancel := context.WithCancel(context.Background())

	fr := &FlowReceiver{
		ctx:    ctx,
		cancel: cancel,

		opts:    opts,
		factory: f,

		runner: newComponentRunner(opts.Logger),
	}
	if err := fr.Update(args); err != nil {
		return nil, err
	}
	return fr, nil
}

var (
	_ component.Component       = (*FlowReceiver)(nil)
	_ component.HealthComponent = (*FlowReceiver)(nil)
)

// Run runs the receiver. Run will wait for OpenTelemetry Collector receiver
// configs to be created.
func (r *FlowReceiver) Run(ctx context.Context) error {
	defer r.cancel()
	return r.runner.Run(ctx)
}

// Update implements component.Component. It will generate OpenTelemetry
// Collector receiver configs and schedule components to get created in the
// background once the FlowReceiver is running.
func (r *FlowReceiver) Update(args component.Arguments) error {
	rargs := args.(ReceiverArguments)

	h := flowHost{
		log: r.opts.Logger,

		extensions: rargs.Extensions(),
		exporters:  rargs.Exporters(),
	}

	settings := otelcomponent.ReceiverCreateSettings{
		TelemetrySettings: otelcomponent.TelemetrySettings{
			Logger: zapadapter.New(r.opts.Logger),

			// TODO(rfratto): should these be set to something real at some point?
			TracerProvider: trace.NewNoopTracerProvider(),
			MeterProvider:  metric.NewNoopMeterProvider(),
		},

		BuildInfo: otelcomponent.BuildInfo{
			Command:     os.Args[0],
			Description: "Grafana Agent",
			Version:     build.Version,
		},
	}

	var (
		receiverConfig = rargs.Convert()

		nextMetrics = rargs.NextArguments().MetricsConsumer()
		nextLogs    = rargs.NextArguments().LogsConsumer()
		nextTraces  = rargs.NextArguments().TracesConsumer()
	)

	var schedule []otelcomponent.Component

	metricsRecv, err := r.factory.CreateMetricsReceiver(r.ctx, settings, receiverConfig, nextMetrics)
	if err != nil && !errors.Is(err, otelcomponenterror.ErrDataTypeIsNotSupported) {
		return err
	} else if metricsRecv != nil {
		schedule = append(schedule, metricsRecv)
	}

	logsRecv, err := r.factory.CreateLogsReceiver(r.ctx, settings, receiverConfig, nextLogs)
	if err != nil && !errors.Is(err, otelcomponenterror.ErrDataTypeIsNotSupported) {
		return err
	} else if logsRecv != nil {
		schedule = append(schedule, logsRecv)
	}

	tracesRecv, err := r.factory.CreateTracesReceiver(r.ctx, settings, receiverConfig, nextTraces)
	if err != nil && !errors.Is(err, otelcomponenterror.ErrDataTypeIsNotSupported) {
		return err
	} else if tracesRecv != nil {
		schedule = append(schedule, tracesRecv)
	}

	// Schedule the components to run when our component is running.
	r.runner.Schedule(&h, schedule...)
	return nil
}

// CurrentHealth implements component.HealthComponent.
func (r *FlowReceiver) CurrentHealth() component.Health {
	return r.runner.CurrentHealth()
}
