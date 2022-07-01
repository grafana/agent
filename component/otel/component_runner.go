package otel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
	"go.uber.org/multierr"
)

// componentRunner is shared between Flow component implementations, manaaging
// the running of OpenTelemetry Collector components.
type componentRunner struct {
	log log.Logger

	healthMut sync.RWMutex
	health    component.Health

	schedMut        sync.Mutex
	schedComponents []otelcomponent.Component // Most recently created components
	host            otelcomponent.Host

	onComponents chan struct{}
}

func newComponentRunner(l log.Logger) *componentRunner {
	return &componentRunner{
		log:          l,
		onComponents: make(chan struct{}, 1),
	}
}

// Schedule schedules a new set of components to run. Schedule may be called
// before Run, but scheduled components won't start until Run has been invoked.
func (cr *componentRunner) Schedule(h otelcomponent.Host, cc ...otelcomponent.Component) {
	cr.schedMut.Lock()
	defer cr.schedMut.Unlock()

	cr.schedComponents = cc
	cr.host = h

	select {
	case cr.onComponents <- struct{}{}:
		// Queued new message.
	default:
		// Nothing to do: this would be the case if onComponents is full (i.e., the
		// message is already sent but not handled yet) and we don't need to do
		// anything else.
	}
}

// Run starts the componentRunner. Run will watch for scheduled components to
// appear and run them, terminating previously running components if they
// exist.
func (cr *componentRunner) Run(ctx context.Context) error {
	var components []otelcomponent.Component

	// Make sure we terminate all of our running components on shutdown.
	defer func() {
		cr.stopComponents(context.Background(), components...)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-cr.onComponents:
			// We must stop the old components before running the new scheduled ones.
			cr.stopComponents(ctx, components...)

			cr.schedMut.Lock()
			components = cr.schedComponents
			host := cr.host
			cr.schedMut.Unlock()

			level.Debug(cr.log).Log("msg", "scheduling components", "count", len(components))
			cr.startComponents(ctx, host, components...)
		}
	}
}

func (cr *componentRunner) stopComponents(ctx context.Context, cc ...otelcomponent.Component) {
	for _, c := range cc {
		if err := c.Shutdown(ctx); err != nil {
			level.Error(cr.log).Log("msg", "failed to stop down inner otel component, future updates may fail", "err", err)
		}
	}
}

func (cr *componentRunner) startComponents(ctx context.Context, h otelcomponent.Host, cc ...otelcomponent.Component) {
	var errs error

	for _, c := range cc {
		if err := c.Start(ctx, h); err != nil {
			level.Error(cr.log).Log("msg", "failed when starting component", "err", err)
			errs = multierr.Append(errs, err)
		}
	}

	if errs != nil {
		cr.setHealth(component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    fmt.Sprintf("failed to create components: %s", errs),
			UpdateTime: time.Now(),
		})
	} else {
		cr.setHealth(component.Health{
			Health:     component.HealthTypeHealthy,
			Message:    "created components",
			UpdateTime: time.Now(),
		})
	}
}

// CurrentHealth implements component.HealthComponent.
func (cr *componentRunner) CurrentHealth() component.Health {
	cr.healthMut.RLock()
	defer cr.healthMut.RUnlock()
	return cr.health
}

func (cr *componentRunner) setHealth(h component.Health) {
	cr.healthMut.Lock()
	defer cr.healthMut.Unlock()
	cr.health = h
}

type flowHost struct {
	log log.Logger

	extensions map[otelconfig.ComponentID]otelcomponent.Extension
	exporters  map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter
}

var _ otelcomponent.Host = (*flowHost)(nil)

func (h *flowHost) ReportFatalError(err error) {
	level.Error(h.log).Log("msg", "fatal error running component", "err", err)
}

func (h *flowHost) GetFactory(kind otelcomponent.Kind, componentType otelconfig.Type) otelcomponent.Factory {
	// GetFactory is used for components to create other components. It's not
	// clear if we want to allow this right now, so it's disabled.
	return nil
}

func (h *flowHost) GetExtensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return h.extensions
}

func (h *flowHost) GetExporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return h.exporters
}
