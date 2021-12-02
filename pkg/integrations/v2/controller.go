package integrations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	prom_config "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	http_sd "github.com/prometheus/prometheus/discovery/http"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// controllerConfig holds a set of integration configs.
type controllerConfig []Config

// controller manages a set of integrations. controller is intended to be
// embedded inside of integrations that multiplex running other integrations.
type controller struct {
	mut  sync.Mutex
	cfg  controllerConfig
	opts Options

	integrations       []*controlledIntegration // Integrations to run
	reloadIntegrations chan struct{}            // Inform Controller.Run to re-read integrations

	// Next generation value to use for an integration.
	gen uint64

	// onUpdateDone is used for testing and will be invoked when integrations
	// finish reloading.
	onUpdateDone func()
}

// newController creates a new Controller. Controller is intended to be
// embedded inside of integrations that may want to multiplex other
// integrations.
func newController(cfg controllerConfig, opts Options) (*controller, error) {
	c := &controller{
		reloadIntegrations: make(chan struct{}, 1),
	}
	if err := c.UpdateController(cfg, opts); err != nil {
		return nil, err
	}
	return c, nil
}

// run starts the controller and blocks until ctx is canceled.
func (c *controller) run(ctx context.Context) {
	defer func() {
		level.Debug(c.opts.Logger).Log("msg", "stopping all integrations")

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

		level.Debug(c.opts.Logger).Log("msg", "updating running integrations")
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
			level.Debug(c.opts.Logger).Log("msg", "controller exiting")
			return
		case <-c.reloadIntegrations:
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
	return ci.i.RunIntegration(ctx)
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
//
// UpdateController updates running integrations. Extensions can be
// recalculated by calling relevant methods like Handler or Targets.
func (c *controller) UpdateController(cfg controllerConfig, opts Options) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	// If iops has changed between calls, then we need to consider all
	// integrations as updated.
	//
	// NOTE(rfratto): while we _could_ pass the new opts to UpdateIntegration
	// and only restart everything else, I don't think it's worth it. The things
	// that could eventually update in opts will eventually be made static for the
	// process lifetime: https://github.com/grafana/agent/issues/581
	forceUpdate := !opts.Equals(c.opts)

	integrationIDMap := map[integrationID]struct{}{}

	integrations := make([]*controlledIntegration, 0, len(cfg))

NextConfig:
	for _, ic := range cfg {
		name := ic.Name()

		// Create a new set of integration options for each integration. This includes
		// a temporary logger for the next few calls. A final logger will be configured
		// before calling NewIntegration.
		icOpts := opts
		icOpts.Logger = log.With(opts.Logger, "integration", name)

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
		// multiplexConfigs shouldn't have the integration/identifier logs set
		// because the fields would be duplicated in the logs.
		//
		// https://github.com/go-kit/log/issues/16 may make this easier.
		if _, ok := ic.(MultiplexConfig); ok {
			icOpts.Logger = opts.Logger
		} else {
			icOpts.Logger = log.With(opts.Logger, "integration", name, "identifier", identifier)
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
	}

	// Recalculate HTTP paths to use for integrations.
	for _, integration := range integrations {
		integration, ok := integration.i.(HTTPIntegration)
		if !ok {
			continue
		}

		_ = integration
	}

	// Update integrations and inform
	c.integrations = integrations
	c.reloadIntegrations <- struct{}{}

	c.cfg = cfg
	c.opts = opts
	return nil
}

// Handler returns an HTTP handler for the controller and its integrations.
// Handler will pass through requests to other running integrations. Handler
// may return a partial handler with an error.
//
// Handler is expensive to compute and should only be done after reloading the
// config.
func (c *controller) Handler(prefix string) (http.Handler, error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	var firstErr error
	saveFirstErr := func(err error) {
		if firstErr == nil {
			firstErr = err
		}
	}

	r := mux.NewRouter()

	forEachIntegration(c.integrations, prefix, func(ci *controlledIntegration, iprefix string) {
		id := ci.id

		i, ok := ci.i.(HTTPIntegration)
		if !ok {
			return
		}

		handler, err := i.Handler(iprefix + "/")
		if errors.Is(err, ErrDisabled) {
			return
		} else if err != nil {
			saveFirstErr(fmt.Errorf("could not generate HTTP handler for %s integration %q: %w", id.Name, id.Identifier, err))
			return
		} else if handler == nil {
			return
		}

		// Anything that matches the integrationPrefix should be passed to the handler.
		r.PathPrefix(iprefix).HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if !ci.Running() {
				http.Error(rw, fmt.Sprintf("%s integration intance %q not running", id.Name, id.Identifier), http.StatusServiceUnavailable)
				return
			}
			handler.ServeHTTP(rw, r)
		})
	})

	// TODO(rfratto): navigation page for exact prefix match

	return r, firstErr
}

// forEachIntegration calculates the prefix for each integration and calls f.
// prefix will not end in /.
func forEachIntegration(set []*controlledIntegration, basePrefix string, f func(ci *controlledIntegration, iprefix string)) error {
	// Pre-populate a mapping of integration name -> identifier. If there are
	// two instances of the same integration, we want to ensure unique routing.
	//
	// This special logic is done for backwards compatibility with the original
	// design of integrations.
	identifiersMap := map[string][]string{}
	for _, i := range set {
		identifiersMap[i.id.Name] = append(identifiersMap[i.id.Name], i.id.Identifier)
	}

	usedPrefixes := map[string]struct{}{}

	for _, ci := range set {
		id := ci.id
		multipleInstances := len(identifiersMap[id.Name]) > 1

		var integrationPrefix string
		if multipleInstances {
			// i.e., /integrations/mysqld_exporter/server-a
			integrationPrefix = path.Join(basePrefix, id.Name, id.Identifier)
		} else {
			// i.e., /integrations/node_exporter
			integrationPrefix = path.Join(basePrefix, id.Name)
		}

		f(ci, integrationPrefix)

		if _, exist := usedPrefixes[integrationPrefix]; exist {
			return fmt.Errorf("BUG: duplicate integration prefix %q", integrationPrefix)
		}
		usedPrefixes[integrationPrefix] = struct{}{}
	}
	return nil
}

// Targets returns a channel that emits the set of target groups across all
// running integrations that implement MetricsIntegration.
func (c *controller) Targets(prefix string) []*targetgroup.Group {
	// Grab the integrations as fast as possible. We don't want to spend too much
	// time holding the mutex.
	type prefixedMetricsIntegration struct {
		i      MetricsIntegration
		prefix string
	}
	var mm []prefixedMetricsIntegration

	c.mut.Lock()
	forEachIntegration(c.integrations, prefix, func(ci *controlledIntegration, iprefix string) {
		// Best effort liveness check. They might stop running when we request
		// their targets, which is fine, but we should save as much work as we
		// can.
		if !ci.Running() {
			return
		}
		if mi, ok := ci.i.(MetricsIntegration); ok {
			mm = append(mm, prefixedMetricsIntegration{i: mi, prefix: iprefix})
		}
	})
	c.mut.Unlock()

	var tgs []*targetgroup.Group
	for _, mi := range mm {
		tgs = append(tgs, mi.i.Targets(mi.prefix)...)
	}
	sort.Slice(tgs, func(i, j int) bool {
		return tgs[i].Source < tgs[j].Source
	})
	return tgs
}

// ScrapeConfigs returns a set of scrape configs to use for self-scraping.
// sdConfig should contain the full URL where the integrations SD API is
// exposed. ScrapeConfigs will inject unique query parameters per integration
// to limit what will be discovered.
func (c *controller) ScrapeConfigs(prefix string, sdConfig *http_sd.SDConfig) []*prom_config.ScrapeConfig {
	// Grab the integrations as fast as possible. We don't want to spend too much
	// time holding the mutex.
	type prefixedMetricsIntegration struct {
		id     integrationID
		i      MetricsIntegration
		prefix string
	}
	var mm []prefixedMetricsIntegration

	c.mut.Lock()
	forEachIntegration(c.integrations, prefix, func(ci *controlledIntegration, iprefix string) {
		// Best effort liveness check. They might stop running when we request
		// their targets, which is fine, but we should save as much work as we
		// can.
		if !ci.Running() {
			return
		}
		if mi, ok := ci.i.(MetricsIntegration); ok {
			mm = append(mm, prefixedMetricsIntegration{id: ci.id, i: mi, prefix: iprefix})
		}
	})
	c.mut.Unlock()

	var cfgs []*prom_config.ScrapeConfig
	for _, mi := range mm {
		// sdConfig will be pointing to the targets API. We want to use the query
		// parmaeters to inform the API to only return specific targets.
		var queryParams url.Values
		queryParams.Set("integration", mi.id.Name)
		queryParams.Set("identifier", mi.id.Identifier)

		integrationSDConfig := *sdConfig
		integrationSDConfig.URL = sdConfig.URL + "?" + queryParams.Encode()
		sds := discovery.Configs{&integrationSDConfig}
		cfgs = append(cfgs, mi.i.ScrapeConfigs(sds)...)
	}
	sort.Slice(cfgs, func(i, j int) bool {
		return cfgs[i].JobName < cfgs[j].JobName
	})
	return cfgs
}
