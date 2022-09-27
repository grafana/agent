package otelcol

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.uber.org/multierr"
)

// componentScheduler implements manages a set of OpenTelemetry Collector
// components. componentScheduler is intended to be used from Flow components
// which need to schedule OpenTelemetry Collector components; it does not
// implement the full component.Component interface.
//
// Each OpenTelemetry Collector component has one instance per supported
// telemetry signal, hence supporting multiple OpenTelemetry Collector
// components inside the scheduler. componentScheduler should only be used to
// manage multiple instances of the same OpenTelemetry Collector component.
type componentScheduler struct {
	log log.Logger

	healthMut sync.RWMutex
	health    component.Health

	schedMut        sync.Mutex
	schedComponents []otelcomponent.Component // Most recently created components
	host            otelcomponent.Host

	// newComponentsCh is written to when schedComponents gets updated.
	newComponentsCh chan struct{}
}

func newComponentScheduler(l log.Logger) *componentScheduler {
	return &componentScheduler{
		log:             l,
		newComponentsCh: make(chan struct{}, 1),
	}
}

func (cs *componentScheduler) Schedule(h otelcomponent.Host, cc ...otelcomponent.Component) {
	cs.schedMut.Lock()
	defer cs.schedMut.Unlock()

	cs.schedComponents = cc
	cs.host = h

	select {
	case cs.newComponentsCh <- struct{}{}:
		// Queued new message.
	default:
		// A message is already queued for refreshing running components so we
		// don't have to do anything here.
	}
}

// Run starts the componentScheduler. Run will watch for schedule components to
// appear and run them, terminating previously running components if they
// exist.
func (cs *componentScheduler) Run(ctx context.Context) error {
	var components []otelcomponent.Component

	// Make sure we terminate all of our running components on shutdown.
	defer func() {
		cs.stopComponents(context.Background(), components...)
	}()

	// Wait for a write to cs.newComponentsCh. The initial list of components is
	// always empty so there's nothing to do until cs.newComponentsCh is written
	// to.
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-cs.newComponentsCh:
			// Stop the old components before running new scheduled ones.
			cs.stopComponents(ctx, components...)

			cs.schedMut.Lock()
			components = cs.schedComponents
			host := cs.host
			cs.schedMut.Unlock()

			level.Debug(cs.log).Log("msg", "scheduling components", "count", len(components))
			cs.startComponents(ctx, host, components...)
		}
	}
}

func (cs *componentScheduler) stopComponents(ctx context.Context, cc ...otelcomponent.Component) {
	for _, c := range cc {
		if err := c.Shutdown(ctx); err != nil {
			level.Error(cs.log).Log("msg", "failed to stop scheduled component; future updates may fail", "err", err)
		}
	}
}

func (cs *componentScheduler) startComponents(ctx context.Context, h otelcomponent.Host, cc ...otelcomponent.Component) {
	var errs error

	for _, c := range cc {
		if err := c.Start(ctx, h); err != nil {
			level.Error(cs.log).Log("msg", "failed to start scheduled component", "err", err)
			errs = multierr.Append(errs, err)
		}
	}

	if errs != nil {
		cs.setHealth(component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    fmt.Sprintf("failed to create components: %s", errs),
			UpdateTime: time.Now(),
		})
	} else {
		cs.setHealth(component.Health{
			Health:     component.HealthTypeHealthy,
			Message:    "started scheduled components",
			UpdateTime: time.Now(),
		})
	}
}

// CurrentHealth implements component.HealthComponent.
func (cs *componentScheduler) CurrentHealth() component.Health {
	cs.healthMut.RLock()
	defer cs.healthMut.RUnlock()
	return cs.health
}

func (cs *componentScheduler) setHealth(h component.Health) {
	cs.healthMut.Lock()
	defer cs.healthMut.Unlock()
	cs.health = h
}
