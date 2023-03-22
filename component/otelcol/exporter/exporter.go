// Package exporter exposes utilities to create a Flow component from
// OpenTelemetry Collector exporters.
package exporter

import (
	"context"
	"errors"
	"os"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/lazycollector"
	"github.com/grafana/agent/component/otelcol/internal/lazyconsumer"
	"github.com/grafana/agent/component/otelcol/internal/scheduler"
	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/util/zapadapter"
	"github.com/prometheus/client_golang/prometheus"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
	sdkprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"

	_ "github.com/grafana/agent/component/otelcol/internal/featuregate" // Enable needed feature gates
)

// Arguments is an extension of component.Arguments which contains necessary
// settings for OpenTelemetry Collector exporters.
type Arguments interface {
	component.Arguments

	// Convert converts the Arguments into an OpenTelemetry Collector exporter
	// configuration.
	Convert() (otelconfig.Exporter, error)

	// Extensions returns the set of extensions that the configured component is
	// allowed to use.
	Extensions() map[otelconfig.ComponentID]otelcomponent.Extension

	// Exporters returns the set of exporters that are exposed to the configured
	// component.
	Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter
}

// Exporter is a Flow component shim which manages an OpenTelemetry Collector
// exporter component.
type Exporter struct {
	ctx    context.Context
	cancel context.CancelFunc

	opts     component.Options
	factory  otelcomponent.ExporterFactory
	consumer *lazyconsumer.Consumer

	sched     *scheduler.Scheduler
	collector *lazycollector.Collector
}

var (
	_ component.Component       = (*Exporter)(nil)
	_ component.HealthComponent = (*Exporter)(nil)
)

// New creates a new Flow component which encapsulates an OpenTelemetry
// Collector exporter. args must hold a value of the argument type registered
// with the Flow component.
//
// The registered component must be registered to export the
// otelcol.ConsumerExports type, otherwise New will panic.
func New(opts component.Options, f otelcomponent.ExporterFactory, args Arguments) (*Exporter, error) {
	ctx, cancel := context.WithCancel(context.Background())

	consumer := lazyconsumer.New(ctx)

	// Create a lazy collector where metrics from the upstream component will be
	// forwarded.
	collector := lazycollector.New()
	opts.Registerer.MustRegister(collector)

	// Immediately set our state with our consumer. The exports will never change
	// throughout the lifetime of our component.
	//
	// This will panic if the wrapping component is not registered to export
	// otelcol.ConsumerExports.
	opts.OnStateChange(otelcol.ConsumerExports{Input: consumer})

	e := &Exporter{
		ctx:    ctx,
		cancel: cancel,

		opts:     opts,
		factory:  f,
		consumer: consumer,

		sched:     scheduler.New(opts.Logger),
		collector: collector,
	}
	if err := e.Update(args); err != nil {
		return nil, err
	}
	return e, nil
}

// Run starts the Exporter component.
func (e *Exporter) Run(ctx context.Context) error {
	defer e.cancel()
	return e.sched.Run(ctx)
}

// Update implements component.Component. It will convert the Arguments into
// configuration for OpenTelemetry Collector exporter configuration and manage
// the underlying OpenTelemetry Collector exporter.
func (e *Exporter) Update(args component.Arguments) error {
	eargs := args.(Arguments)

	host := scheduler.NewHost(
		e.opts.Logger,
		scheduler.WithHostExtensions(eargs.Extensions()),
		scheduler.WithHostExporters(eargs.Exporters()),
	)

	reg := prometheus.NewRegistry()
	e.collector.Set(reg)

	promExporter, err := sdkprometheus.New(sdkprometheus.WithRegisterer(reg), sdkprometheus.WithoutTargetInfo())
	if err != nil {
		return err
	}

	settings := otelcomponent.ExporterCreateSettings{
		TelemetrySettings: otelcomponent.TelemetrySettings{
			Logger: zapadapter.New(e.opts.Logger),

			TracerProvider: e.opts.Tracer,
			MeterProvider:  metric.NewMeterProvider(metric.WithReader(promExporter)),
		},

		BuildInfo: otelcomponent.BuildInfo{
			Command:     os.Args[0],
			Description: "Grafana Agent",
			Version:     build.Version,
		},
	}

	exporterConfig, err := eargs.Convert()
	if err != nil {
		return err
	}

	// Create instances of the exporter from our factory for each of our
	// supported telemetry signals.
	var components []otelcomponent.Component

	tracesExporter, err := e.factory.CreateTracesExporter(e.ctx, settings, exporterConfig)
	if err != nil && !errors.Is(err, otelcomponent.ErrDataTypeIsNotSupported) {
		return err
	} else if tracesExporter != nil {
		components = append(components, tracesExporter)
	}

	metricsExporter, err := e.factory.CreateMetricsExporter(e.ctx, settings, exporterConfig)
	if err != nil && !errors.Is(err, otelcomponent.ErrDataTypeIsNotSupported) {
		return err
	} else if metricsExporter != nil {
		components = append(components, metricsExporter)
	}

	logsExporter, err := e.factory.CreateLogsExporter(e.ctx, settings, exporterConfig)
	if err != nil && !errors.Is(err, otelcomponent.ErrDataTypeIsNotSupported) {
		return err
	} else if logsExporter != nil {
		components = append(components, logsExporter)
	}

	// Schedule the components to run once our component is running.
	e.sched.Schedule(host, components...)
	e.consumer.SetConsumers(tracesExporter, metricsExporter, logsExporter)
	return nil
}

// CurrentHealth implements component.HealthComponent.
func (e *Exporter) CurrentHealth() component.Health {
	return e.sched.CurrentHealth()
}
