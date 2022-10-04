// Package processor exposes utilities to create a Flow component from
// OpenTelemetry Collector processors.
package processor

import (
	"context"
	"errors"
	"os"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/fanoutconsumer"
	"github.com/grafana/agent/component/otelcol/internal/lazyconsumer"
	"github.com/grafana/agent/component/otelcol/internal/scheduler"
	"github.com/grafana/agent/component/otelcol/internal/zapadapter"
	"github.com/grafana/agent/pkg/build"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Arguments is an extension of component.Arguments which contains necessary
// settings for OpenTelemetry Collector processors.
type Arguments interface {
	component.Arguments

	// Convert converts the Arguments into an OpenTelemetry Collector processor
	// configuration.
	Convert() otelconfig.Processor

	// Extensions returns the set of extensions that the configured component is
	// allowed to use.
	Extensions() map[otelconfig.ComponentID]otelcomponent.Extension

	// Exporters returns the set of exporters that are exposed to the configured
	// component.
	Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter

	// NextConsumers returns the set of consumers to send data to.
	NextConsumers() *otelcol.ConsumerArguments
}

// Processor is a Flow component shim which manages an OpenTelemetry Collector
// processor component.
type Processor struct {
	ctx    context.Context
	cancel context.CancelFunc

	opts     component.Options
	factory  otelcomponent.ProcessorFactory
	consumer *lazyconsumer.Consumer

	sched *scheduler.Scheduler
}

var (
	_ component.Component       = (*Processor)(nil)
	_ component.HealthComponent = (*Processor)(nil)
)

// New creates a new Flow component which encapsulates an OpenTelemetry
// Collector processor. args must hold a value of the argument type registered
// with the Flow component.
//
// The registered component must be registered to export the
// otelcol.ConsumerExports type, otherwise New will panic.
func New(opts component.Options, f otelcomponent.ProcessorFactory, args Arguments) (*Processor, error) {
	ctx, cancel := context.WithCancel(context.Background())

	consumer := lazyconsumer.New(ctx)

	// Immediately set our state with our consumer. The exports will never change
	// throughout the lifetime of our component.
	//
	// This will panic if the wrapping component is not registered to export
	// otelcol.ConsumerExports.
	opts.OnStateChange(otelcol.ConsumerExports{Input: consumer})

	p := &Processor{
		ctx:    ctx,
		cancel: cancel,

		opts:     opts,
		factory:  f,
		consumer: consumer,

		sched: scheduler.New(opts.Logger),
	}
	if err := p.Update(args); err != nil {
		return nil, err
	}
	return p, nil
}

// Run starts the Processor component.
func (p *Processor) Run(ctx context.Context) error {
	defer p.cancel()
	return p.sched.Run(ctx)
}

// Update implements component.Component. It will convert the Arguments into
// configuration for OpenTelemetry Collector processor configuration and manage
// the underlying OpenTelemetry Collector processor.
func (p *Processor) Update(args component.Arguments) error {
	pargs := args.(Arguments)

	host := scheduler.NewHost(
		p.opts.Logger,
		scheduler.WithHostExtensions(pargs.Extensions()),
		scheduler.WithHostExporters(pargs.Exporters()),
	)

	settings := otelcomponent.ProcessorCreateSettings{
		TelemetrySettings: otelcomponent.TelemetrySettings{
			Logger: zapadapter.New(p.opts.Logger),

			// TODO(rfratto): expose tracing and logging statistics.
			//
			// We may want to put off tracing until we have native tracing
			// instrumentation from Flow, but metrics should come sooner since we're
			// already set up for supporting component-specific metrics.
			TracerProvider: trace.NewNoopTracerProvider(),
			MeterProvider:  metric.NewNoopMeterProvider(),
		},

		BuildInfo: otelcomponent.BuildInfo{
			Command:     os.Args[0],
			Description: "Grafana Agent",
			Version:     build.Version,
		},
	}

	processorConfig := pargs.Convert()

	var (
		next        = pargs.NextConsumers()
		nextTraces  = fanoutconsumer.Traces(next.Traces)
		nextMetrics = fanoutconsumer.Metrics(next.Metrics)
		nextLogs    = fanoutconsumer.Logs(next.Logs)
	)

	// Create instances of the processor from our factory for each of our
	// supported telemetry signals.
	var components []otelcomponent.Component

	tracesProcessor, err := p.factory.CreateTracesProcessor(p.ctx, settings, processorConfig, nextTraces)
	if err != nil && !errors.Is(err, otelcomponent.ErrDataTypeIsNotSupported) {
		return err
	} else if tracesProcessor != nil {
		components = append(components, tracesProcessor)
	}

	metricsProcessor, err := p.factory.CreateMetricsProcessor(p.ctx, settings, processorConfig, nextMetrics)
	if err != nil && !errors.Is(err, otelcomponent.ErrDataTypeIsNotSupported) {
		return err
	} else if metricsProcessor != nil {
		components = append(components, metricsProcessor)
	}

	logsProcessor, err := p.factory.CreateLogsProcessor(p.ctx, settings, processorConfig, nextLogs)
	if err != nil && !errors.Is(err, otelcomponent.ErrDataTypeIsNotSupported) {
		return err
	} else if logsProcessor != nil {
		components = append(components, logsProcessor)
	}

	// Schedule the components to run once our component is running.
	p.sched.Schedule(host, components...)
	p.consumer.SetConsumers(tracesProcessor, metricsProcessor, logsProcessor)
	return nil
}

// CurrentHealth implements component.HealthComponent.
func (p *Processor) CurrentHealth() component.Health {
	return p.sched.CurrentHealth()
}
