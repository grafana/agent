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

// FlowProcessor is a Flow component implementation which manages OpenTelemetry
// Collector processors.
type FlowProcessor struct {
	ctx    context.Context
	cancel context.CancelFunc

	opts    component.Options
	factory otelcomponent.ProcessorFactory

	runner *componentRunner
}

type ProcessorArguments interface {
	component.Arguments

	// Convert should convert the ProcessorArguments into a Processor confguration
	// used by OpenTelemetry Collector.
	Convert() otelconfig.Processor

	// Extensions should return the set of extensions that this component is
	// allowed to use.
	Extensions() map[otelconfig.ComponentID]otelcomponent.Extension

	// Exporters should return the set of exporters that are exposed to this
	// component.
	Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter

	// NextArguments returns the set of components to send data to.
	NextArguments() *NextReceiverArguments
}

// NewFlowProcessor creates a new Flow component which encapsules an
// OpenTelemetry Collector processor. args must be the argument type registered
// with the component.
func NewFlowProcessor(opts component.Options, f otelcomponent.ProcessorFactory, args ProcessorArguments) (*FlowProcessor, error) {
	ctx, cancel := context.WithCancel(context.Background())

	fp := &FlowProcessor{
		ctx:    ctx,
		cancel: cancel,

		opts:    opts,
		factory: f,

		runner: newComponentRunner(opts.Logger),
	}
	if err := fp.Update(args); err != nil {
		return nil, err
	}
	return fp, nil
}

var (
	_ component.Component       = (*FlowProcessor)(nil)
	_ component.HealthComponent = (*FlowProcessor)(nil)
)

// Run runs the processor. Run will wait for OpenTelemetry Collector receiver
// configs to be created.
func (p *FlowProcessor) Run(ctx context.Context) error {
	defer p.cancel()
	return p.runner.Run(ctx)
}

// Update implements component.Component. It will generate OpenTelemetry
// Collector processor configs and schedule components to get created in the
// background once the FlowReceiver is running.
func (p *FlowProcessor) Update(args component.Arguments) error {
	rargs := args.(ProcessorArguments)

	h := flowHost{
		log: p.opts.Logger,

		extensions: rargs.Extensions(),
		exporters:  rargs.Exporters(),
	}

	settings := otelcomponent.ProcessorCreateSettings{
		TelemetrySettings: otelcomponent.TelemetrySettings{
			Logger: zapadapter.New(p.opts.Logger),

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
		processorConfig = rargs.Convert()

		nextMetrics = rargs.NextArguments().MetricsConsumer()
		nextLogs    = rargs.NextArguments().LogsConsumer()
		nextTraces  = rargs.NextArguments().TracesConsumer()
	)

	var schedule []otelcomponent.Component

	metricsProc, err := p.factory.CreateMetricsProcessor(p.ctx, settings, processorConfig, nextMetrics)
	if err != nil && !errors.Is(err, otelcomponenterror.ErrDataTypeIsNotSupported) {
		return err
	} else if metricsProc != nil {
		schedule = append(schedule, metricsProc)
	}

	logsProc, err := p.factory.CreateLogsProcessor(p.ctx, settings, processorConfig, nextLogs)
	if err != nil && !errors.Is(err, otelcomponenterror.ErrDataTypeIsNotSupported) {
		return err
	} else if metricsProc != nil {
		schedule = append(schedule, logsProc)
	}

	tracesProc, err := p.factory.CreateTracesProcessor(p.ctx, settings, processorConfig, nextTraces)
	if err != nil && !errors.Is(err, otelcomponenterror.ErrDataTypeIsNotSupported) {
		return err
	} else if metricsProc != nil {
		schedule = append(schedule, tracesProc)
	}

	// Schedule the components to run when our component is running.
	p.runner.Schedule(&h, schedule...)
	return nil
}

// CurrentHealth implements component.HealthComponent.
func (p *FlowProcessor) CurrentHealth() component.Health {
	return p.runner.CurrentHealth()
}
