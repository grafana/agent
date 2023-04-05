// Package extension provides utilities to create a Flow component from
// OpenTelemetry Collector extensions.
//
// Other OpenTelemetry Collector extensions are better served as generic Flow
// components rather than being placed in the otelcol namespace.
package extension

import (
	"context"
	"os"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol/internal/lazycollector"
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
// settings for OpenTelemetry Collector extensions.
type Arguments interface {
	component.Arguments

	// Convert converts the Arguments into an OpenTelemetry Collector
	// extension configuration.
	Convert() (otelconfig.Extension, error)

	// Extensions returns the set of extensions that the configured component is
	// allowed to use.
	Extensions() map[otelconfig.ComponentID]otelcomponent.Extension

	// Exporters returns the set of exporters that are exposed to the configured
	// component.
	Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter
}

// Extension is a Flow component shim which manages an OpenTelemetry Collector
// extension.
type Extension struct {
	ctx    context.Context
	cancel context.CancelFunc

	opts    component.Options
	factory otelcomponent.ExtensionFactory

	sched     *scheduler.Scheduler
	collector *lazycollector.Collector
}

var (
	_ component.Component       = (*Extension)(nil)
	_ component.HealthComponent = (*Extension)(nil)
)

// New creates a new Flow component which encapsulates an OpenTelemetry
// Collector extension. args must hold a value of the argument
// type registered with the Flow component.
func New(opts component.Options, f otelcomponent.ExtensionFactory, args Arguments) (*Extension, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create a lazy collector where metrics from the upstream component will be
	// forwarded.
	collector := lazycollector.New()
	opts.Registerer.MustRegister(collector)

	r := &Extension{
		ctx:    ctx,
		cancel: cancel,

		opts:    opts,
		factory: f,

		sched:     scheduler.New(opts.Logger),
		collector: collector,
	}
	if err := r.Update(args); err != nil {
		return nil, err
	}
	return r, nil
}

// Run starts the Extension component.
func (e *Extension) Run(ctx context.Context) error {
	defer e.cancel()
	return e.sched.Run(ctx)
}

// Update implements component.Component. It will convert the Arguments into
// configuration for OpenTelemetry Collector extension
// configuration and manage the underlying OpenTelemetry Collector extension.
func (e *Extension) Update(args component.Arguments) error {
	rargs := args.(Arguments)

	host := scheduler.NewHost(
		e.opts.Logger,
		scheduler.WithHostExtensions(rargs.Extensions()),
		scheduler.WithHostExporters(rargs.Exporters()),
	)

	reg := prometheus.NewRegistry()
	e.collector.Set(reg)

	promExporter, err := sdkprometheus.New(sdkprometheus.WithRegisterer(reg), sdkprometheus.WithoutTargetInfo())
	if err != nil {
		return err
	}

	settings := otelcomponent.ExtensionCreateSettings{
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

	extensionConfig, err := rargs.Convert()
	if err != nil {
		return err
	}

	// Create instances of the extension from our factory.
	var components []otelcomponent.Component

	ext, err := e.factory.CreateExtension(e.ctx, settings, extensionConfig)
	if err != nil {
		return err
	} else if ext != nil {
		components = append(components, ext)
	}

	// Schedule the components to run once our component is running.
	e.sched.Schedule(host, components...)
	return nil
}

// CurrentHealth implements component.HealthComponent.
func (e *Extension) CurrentHealth() component.Health {
	return e.sched.CurrentHealth()
}
