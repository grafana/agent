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

// FlowExporter is a Flow component implementation which manages OpenTelemetry
// Collector exporter.
type FlowExporter struct {
	ctx    context.Context
	cancel context.CancelFunc

	opts    component.Options
	factory otelcomponent.ExporterFactory

	runner *componentRunner
}

type ExporterArguments interface {
	component.Arguments

	// Convert should convert the ExporterArguments into a Exporter confguration
	// used by OpenTelemetry Collector.
	Convert() otelconfig.Exporter

	// Extensions should return the set of extensions that this component is
	// allowed to use.
	Extensions() map[otelconfig.ComponentID]otelcomponent.Extension

	// Exporters should return the set of exporters that are exposed to this
	// component.
	Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter
}

// NewFlowExporter creates a new Flow component which encapsules an
// OpenTelemetry Collector exporter. args must be the argument type registered
// with the component.
func NewFlowExporter(opts component.Options, f otelcomponent.ExporterFactory, args ExporterArguments) (*FlowExporter, error) {
	ctx, cancel := context.WithCancel(context.Background())

	fr := &FlowExporter{
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
	_ component.Component       = (*FlowExporter)(nil)
	_ component.HealthComponent = (*FlowExporter)(nil)
)

// Run runs the exporter. Run will wait for OpenTelemetry Collector exporter
// configs to be created.
func (r *FlowExporter) Run(ctx context.Context) error {
	defer r.cancel()
	return r.runner.Run(ctx)
}

// Update implements component.Component. It will generate OpenTelemetry
// Collector exporter configs and schedule components to get created in the
// background once the FlowExporter is running.
func (r *FlowExporter) Update(args component.Arguments) error {
	rargs := args.(ExporterArguments)

	h := flowHost{
		log: r.opts.Logger,

		extensions: rargs.Extensions(),
		exporters:  rargs.Exporters(),
	}

	settings := otelcomponent.ExporterCreateSettings{
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

	var exporterConfig = rargs.Convert()

	var schedule []otelcomponent.Component

	metricsRecv, err := r.factory.CreateMetricsExporter(r.ctx, settings, exporterConfig)
	if err != nil && !errors.Is(err, otelcomponenterror.ErrDataTypeIsNotSupported) {
		return err
	} else if metricsRecv != nil {
		schedule = append(schedule, metricsRecv)
	}

	logsRecv, err := r.factory.CreateLogsExporter(r.ctx, settings, exporterConfig)
	if err != nil && !errors.Is(err, otelcomponenterror.ErrDataTypeIsNotSupported) {
		return err
	} else if metricsRecv != nil {
		schedule = append(schedule, logsRecv)
	}

	tracesRecv, err := r.factory.CreateTracesExporter(r.ctx, settings, exporterConfig)
	if err != nil && !errors.Is(err, otelcomponenterror.ErrDataTypeIsNotSupported) {
		return err
	} else if metricsRecv != nil {
		schedule = append(schedule, tracesRecv)
	}

	// Schedule the components to run when our component is running.
	r.runner.Schedule(&h, schedule...)
	return nil
}

// CurrentHealth implements component.HealthComponent.
func (r *FlowExporter) CurrentHealth() component.Health {
	return r.runner.CurrentHealth()
}
