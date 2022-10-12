// Package auth provides utilities to create a Flow component from
// OpenTelemetry Collector authentication extensions.
//
// Other OpenTelemetry Collector extensions are better served as generic Flow
// components rather than being placed in the otelcol namespace.
package auth

import (
	"context"
	"os"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol/internal/scheduler"
	"github.com/grafana/agent/component/otelcol/internal/zapadapter"
	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/river"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Arguments is an extension of component.Arguments which contains necessary
// settings for OpenTelemetry Collector authentication extensions.
type Arguments interface {
	component.Arguments

	// Convert converts the Arguments into an OpenTelemetry Collector
	// authentication extension configuration.
	Convert() otelconfig.Extension

	// Extensions returns the set of extensions that the configured component is
	// allowed to use.
	Extensions() map[otelconfig.ComponentID]otelcomponent.Extension

	// Exporters returns the set of exporters that are exposed to the configured
	// component.
	Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter
}

// Exports is a common Exports type for Flow components which expose
// OpenTelemetry Collector authentication extensions.
type Exports struct {
	// Handler is the managed compoent. Handler is updated any time the extension
	// is updated.
	Handler Handler `river:"handler,attr"`
}

// Handler combines an extension with its ID.
type Handler struct {
	ID        otelconfig.ComponentID
	Extension otelcomponent.Extension
}

var _ river.Capsule = Handler{}

// RiverCapsule marks Handler as a capsule type.
func (Handler) RiverCapsule() {}

// Auth is a Flow component shim which manages an OpenTelemetry Collector
// authentication extension.
type Auth struct {
	ctx    context.Context
	cancel context.CancelFunc

	opts    component.Options
	factory otelcomponent.ExtensionFactory

	sched *scheduler.Scheduler
}

var (
	_ component.Component       = (*Auth)(nil)
	_ component.HealthComponent = (*Auth)(nil)
)

// New creates a new Flow component which encapsulates an OpenTelemetry
// Collector authentication extension. args must hold a value of the argument
// type registered with the Flow component.
//
// The registered component must be registered to export the Exports type from
// this package, otherwise New will panic.
func New(opts component.Options, f otelcomponent.ExtensionFactory, args Arguments) (*Auth, error) {
	ctx, cancel := context.WithCancel(context.Background())

	r := &Auth{
		ctx:    ctx,
		cancel: cancel,

		opts:    opts,
		factory: f,

		sched: scheduler.New(opts.Logger),
	}
	if err := r.Update(args); err != nil {
		return nil, err
	}
	return r, nil
}

// Run starts the Auth component.
func (r *Auth) Run(ctx context.Context) error {
	defer r.cancel()
	return r.sched.Run(ctx)
}

// Update implements component.Component. It will convert the Arguments into
// configuration for OpenTelemetry Collector authentication extension
// configuration and manage the underlying OpenTelemetry Collector extension.
func (r *Auth) Update(args component.Arguments) error {
	rargs := args.(Arguments)

	host := scheduler.NewHost(
		r.opts.Logger,
		scheduler.WithHostExtensions(rargs.Extensions()),
		scheduler.WithHostExporters(rargs.Exporters()),
	)

	settings := otelcomponent.ExtensionCreateSettings{
		TelemetrySettings: otelcomponent.TelemetrySettings{
			Logger: zapadapter.New(r.opts.Logger),

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

	extensionConfig := rargs.Convert()

	// Create instances of the extension from our factory.
	var components []otelcomponent.Component

	ext, err := r.factory.CreateExtension(r.ctx, settings, extensionConfig)
	if err != nil {
		return err
	} else if ext != nil {
		components = append(components, ext)
	}

	// Inform listeners that our handler changed.
	r.opts.OnStateChange(Exports{
		Handler: Handler{
			ID:        otelconfig.NewComponentID(otelconfig.Type(r.opts.ID)),
			Extension: ext,
		},
	})

	// Schedule the components to run once our component is running.
	r.sched.Schedule(host, components...)
	return nil
}

// CurrentHealth implements component.HealthComponent.
func (r *Auth) CurrentHealth() component.Health {
	return r.sched.CurrentHealth()
}
