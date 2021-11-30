package integrations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"sync"
	"sync/atomic"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	prom_config "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// Controller runs a set of integrations. Controller implements Integration and
// all known extension interfaces. Controller may be used inside other
// integrations to implement integration multiplexing.
type Controller struct {
	mut   sync.Mutex
	copts ControllerOptions
	iopts IntegrationOptions

	integrations       []*controlledIntegration // Integrations to run
	reloadIntegrations chan struct{}            // Inform Controller.Run to re-read integrations

	// Next generation value to use for an integration.
	gen uint64

	// onUpdateDone is used for testing and will be invoked when integrations
	// finish reloading.
	onUpdateDone func()
}

// NewController creates a new Controller. Controller implements Integration
// and integrations will not run until Controller.Run is invoked.
func NewController(copts ControllerOptions, iopts IntegrationOptions) (*Controller, error) {
	c := &Controller{
		copts: copts,
		iopts: iopts,

		reloadIntegrations: make(chan struct{}, 1),
	}
	if err := c.ApplyConfig(&copts); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements Integration. Run will run all integrations.
func (c *Controller) Run(ctx context.Context) error {
	defer func() {
		level.Debug(c.iopts.Logger).Log("msg", "stopping all integrations")

		c.mut.Lock()
		defer c.mut.Unlock()

		for _, exist := range c.integrations {
			exist.Stop()
		}
	}()

	var currentIntegrations []*controlledIntegration

	updateIntegrations := func() {
		// Lock the mutex to prevent another set of integrations from being
		// loaded in.
		c.mut.Lock()
		defer c.mut.Unlock()

		newIntegrations := c.integrations

		// Shut down all old integrations. If the integration exists in
		// newIntegrations but has a different gen number, then there's a new
		// instance to launch.
		for _, exist := range currentIntegrations {
			var found bool
			for _, current := range newIntegrations {
				if exist.id == current.id && current.gen == exist.gen {
					found = true
					break
				}
			}
			if !found {
				exist.Stop()
			}
		}

		// Now all non-running integrations can be launched.
		for _, current := range newIntegrations {
			if current.Running() {
				continue
			}
			go current.Run(ctx)
		}

		// Finally, store the current list of contolled integrations.
		currentIntegrations = newIntegrations
	}

	for {
		select {
		case <-ctx.Done():
			level.Debug(c.iopts.Logger).Log("msg", "controller exiting")
			return nil
		case <-c.reloadIntegrations:
			level.Debug(c.iopts.Logger).Log("msg", "updating running integrations")
			updateIntegrations()

			if c.onUpdateDone != nil {
				c.onUpdateDone()
			}
		}
	}
}

// controlledIntegration is a running Integration.
// A running integration is identified uniquely by its id and gen.
type controlledIntegration struct {
	id  integrationID
	gen uint64

	i Integration
	c Config // Config that generated i. Used for changing to see if a config changed.

	running uint64 // running must only be used atomically

	mut  sync.Mutex
	stop context.CancelFunc
}

func (ci *controlledIntegration) Running() bool {
	return atomic.LoadUint64(&ci.running) == 1
}

func (ci *controlledIntegration) Run(ctx context.Context) error {
	if !atomic.CompareAndSwapUint64(&ci.running, 0, 1) {
		return fmt.Errorf("already running")
	}
	defer atomic.StoreUint64(&ci.running, 0)

	ci.mut.Lock()
	ctx, ci.stop = context.WithCancel(ctx)
	ci.mut.Unlock()

	// Early optimization: don't do anything if ctx has already been canceled
	if ctx.Err() != nil {
		return nil
	}
	return ci.i.Run(ctx)
}

func (ci *controlledIntegration) Stop() {
	ci.mut.Lock()
	if ci.stop != nil {
		ci.stop()
	}
	ci.mut.Unlock()
}

// integrationID uses a tuple of Name and Identifier to uniquely identify an
// integration.
type integrationID struct {
	Name, Identifier string
}

// UpdateController updates the Controller with new Controller and
// IntegrationOptions.
func (c *Controller) UpdateController(copts ControllerOptions, iopts IntegrationOptions) error {
	c.mut.Lock()
	defer c.mut.Unlock()
	return c.updateController(copts, iopts)
}

func (c *Controller) updateController(copts ControllerOptions, iopts IntegrationOptions) error {
	// If iops has changed between calls, then we need to consider all
	// integrations as updated.
	//
	// NOTE(rfratto): while we _could_ pass the new iopts to UpdateIntegration
	// and only restart everything else, I don't think it's worth it. The things
	// that could eventually update in iopts will eventually be made static for the
	// process lifetime: https://github.com/grafana/agent/issues/581
	forceUpdate := !iopts.Equals(c.iopts)

	integrationIDMap := map[integrationID]struct{}{}

	integrations := make([]*controlledIntegration, 0, len(copts.Configs))

NextConfig:
	for _, ic := range copts.Configs {
		name := ic.Name()

		// Create a new set of integration options for each integration. This includes
		// a temporary logger for the next few calls. A final logger will be configured
		// before calling NewIntegration.
		icOpts := iopts
		icOpts.AgentHTTPClientConfig = copts.ClientConfig
		icOpts.Logger = log.With(iopts.Logger, "integration", name)

		identifier, err := ic.Identifier(icOpts)
		if err != nil {
			return fmt.Errorf("could not build identifier for integration %q: %w", name, err)
		}

		id := integrationID{Name: name, Identifier: identifier}
		if _, exist := integrationIDMap[id]; exist {
			return fmt.Errorf("multiple instance names %q in integration %q", identifier, name)
		}
		integrationIDMap[id] = struct{}{}

		// Now that we know the ID for an integration, we can check to see if it's
		// running and can be dynamically updated.
		if forceUpdate {
			// forceUpdate is true when something changed in the environment that cannot
			// be dynamically updated in configs. When this happens, we want to just
			// recreate everything.
			goto CreateIntegration
		}
		for _, ci := range c.integrations {
			if ci.id != id {
				continue
			}

			// If the configs haven't changed, then we don't need to do anything.
			if CompareConfigs(ci.c, ic) {
				integrations = append(integrations, ci)
				continue NextConfig
			}

			if ui, ok := ci.i.(UpdateIntegration); ok {
				if err := ui.ApplyConfig(ic); errors.Is(err, ErrDisabled) {
					// Ignore integration; treat it as removed.
					continue NextConfig
				} else if err != nil {
					return fmt.Errorf("failed to update %s integration %q: %w", name, identifier, err)
				}
				// Re-use the existing controlled integration.
				integrations = append(integrations, ci)
				continue NextConfig
			}

			break
		}

	CreateIntegration:
		// Figure out what logger to give to the integration. Integrations that are
		// controllers shouldn't have the integration/identifier logs set because
		// the fields would be duplicated in the logs.
		//
		// https://github.com/go-kit/log/issues/16 may make this easier.
		if _, ok := ic.(ControllerConfig); ok {
			icOpts.Logger = iopts.Logger
		} else {
			icOpts.Logger = log.With(iopts.Logger, "integration", name, "identifier", identifier)
		}
		integration, err := ic.NewIntegration(icOpts)
		if errors.Is(err, ErrDisabled) {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to construct %s integration %q: %w", name, identifier, err)
		}

		// Create a new conrolled integration.
		integrations = append(integrations, &controlledIntegration{
			id:  id,
			gen: atomic.AddUint64(&c.gen, 1),
			i:   integration,
			c:   ic,
		})

		//
		// Handle extensions
		//

		if integration, ok := integration.(MetricsIntegration); ok {
			// TODO(rfratto): handle?
			_ = integration
		}
	}

	// Recalculate HTTP paths to use for integrations.
	for _, integration := range integrations {
		integration, ok := integration.i.(HTTPIntegration)
		if !ok {
			continue
		}

		_ = integration
	}

	// TODO(rfratto): handle removing extensions for configs that have gone away

	// Update integrations and inform
	c.integrations = integrations
	c.reloadIntegrations <- struct{}{}

	c.copts = copts
	c.iopts = iopts
	return nil
}

// ApplyConfig implements UpdateIntegration. ApplyConfig will update the
// Controller and all running integrations. c MUST implement controllerConfig.
func (c *Controller) ApplyConfig(cfg Config) error {
	controllerConfig, ok := cfg.(ControllerConfig)
	if !ok {
		return fmt.Errorf("invalid type %T passed to Controller.ApplyConfig", cfg)
	}

	c.mut.Lock()
	defer c.mut.Unlock()
	return c.updateController(controllerConfig.ControllerOptions(), c.iopts)
}

// Handler implements HTTPIntegration. Handler will pass through requests to
// other running integrations.
//
// Handler is expensive to compute and should only be done after reloading the
// config.
func (c *Controller) Handler(prefix string) (http.Handler, error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	r := mux.NewRouter()

	// Pre-populate a mapping of integration name -> identifier. If there are
	// two instances of the same integration, we want to ensure unique routing.
	//
	// This special logic is done for backwards compatibility with the original
	// design of integrations.
	identifiersMap := map[string][]string{}
	for _, i := range c.integrations {
		identifiersMap[i.id.Name] = append(identifiersMap[i.id.Name], i.id.Identifier)
	}

	usedPrefixes := map[string]struct{}{}

	for _, i := range c.integrations {
		id := i.id
		multipleInstances := len(identifiersMap[id.Name]) > 1

		i, ok := i.i.(HTTPIntegration)
		if !ok {
			continue
		}

		var integrationPrefix string
		if multipleInstances {
			// i.e., /integrations/mysqld_exporter/server-a
			integrationPrefix = path.Join(prefix, id.Name, id.Identifier)
		} else {
			// i.e., /integrations/node_exporter
			integrationPrefix = path.Join(prefix, id.Name)
		}

		handler, err := i.Handler(integrationPrefix + "/")
		if errors.Is(err, ErrDisabled) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("could not generate HTTP handler for %s integration %q: %w", id.Name, id.Identifier, err)
		} else if handler == nil {
			continue
		}

		if _, exist := usedPrefixes[integrationPrefix]; exist {
			return nil, fmt.Errorf("BUG: duplicate integration prefix %q", integrationPrefix)
		}
		usedPrefixes[integrationPrefix] = struct{}{}

		// Anything that matches the integrationPrefix should be passed to the handler.
		r.PathPrefix(integrationPrefix).Handler(handler)
	}

	// TODO(rfratto): navigation page for exact prefix match

	return r, nil
}

// Targets implements MetricsIntegration. Targets will return a channel that
// emits the set of target groups across all running integrations that also
// implement MetricsIntegration.
func (c *Controller) Targets(prefix string) <-chan []*targetgroup.Group {
	panic("NYI")
}

// ScrapeConfigs implements MetricsIntegration. ScrapeConfigs will return a set
// of scrape configs to use for self-scraping.
func (c *Controller) ScrapeConfigs(d discovery.Configs) []*prom_config.ScrapeConfig {
	panic("NYI")
}
