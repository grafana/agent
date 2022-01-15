package shared

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"

	v2 "github.com/grafana/agent/pkg/integrations/v2"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/prometheus/prometheus/discovery"
	http_sd "github.com/prometheus/prometheus/discovery/http"
	"go.uber.org/atomic"
)

// Controller manages a set of integrations.
type Controller struct {
	logger  log.Logger
	mut     sync.Mutex
	cfg     V2Integrations
	globals v2.Globals

	integrations       []*controlledIntegration // Integrations to Run
	reloadIntegrations chan struct{}            // Inform Controller.Run to re-read integrations

	// Next generation value to use for an integration.
	gen atomic.Uint64

	// onUpdateDone is used for testing and will be invoked when integrations
	// finish reloading.
	onUpdateDone func()
}

// NewController creates a new Controller. Controller is intended to be
// embedded inside of integrations that may want to multiplex other
// integrations.
func NewController(l log.Logger, cfg V2Integrations, globals v2.Globals) (*Controller, error) {
	c := &Controller{
		logger:             l,
		reloadIntegrations: make(chan struct{}, 1),
	}
	if err := c.UpdateController(cfg, globals); err != nil {
		return nil, err
	}
	return c, nil
}

// Run starts the Controller and blocks until ctx is canceled.
func (c *Controller) Run(ctx context.Context) {
	defer func() {
		level.Debug(c.logger).Log("msg", "stopping all integrations")

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
		level.Debug(c.logger).Log("msg", "updating running integrations", "prev_count", len(currentIntegrations), "new_count", len(c.integrations))

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

		var waitStarted sync.WaitGroup
		waitStarted.Add(len(newIntegrations))

		// Now all integrations can be launched.
		for _, current := range newIntegrations {
			go func(current *controlledIntegration) {
				waitStarted.Done()

				err := current.Run(ctx)
				if err != nil && !errors.Is(err, errIntegrationRunning) {
					level.Warn(c.logger).Log("msg", "integration exited with error", "instance", current.id, "err", err)
				}
			}(current)
		}

		// Wait for all integration goroutines to have been scheduled at least once.
		waitStarted.Wait()

		// Finally, store the current list of contolled integrations.
		currentIntegrations = newIntegrations
	}

	for {
		select {
		case <-ctx.Done():
			level.Debug(c.logger).Log("msg", "controller exiting")
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

	i v2.Integrations
	c v2.V2Config // Config that generated i. Used for changing to see if a shared changed.

	running atomic.Bool

	mut  sync.Mutex
	stop context.CancelFunc
}

func (ci *controlledIntegration) Running() bool {
	return ci.running.Load()
}

func (ci *controlledIntegration) Run(ctx context.Context) error {
	updatedRunningState := ci.running.CAS(false, true)
	if !updatedRunningState {
		// The CAS will fail if our integration was already running.
		return errIntegrationRunning
	}
	defer ci.running.Store(false)

	ci.mut.Lock()
	ctx, ci.stop = context.WithCancel(ctx)
	ci.mut.Unlock()

	// Early optimization: don't do anything if ctx has already been canceled
	if ctx.Err() != nil {
		return nil
	}
	return ci.i.RunIntegration(ctx)
}

var errIntegrationRunning = fmt.Errorf("already running")

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

func (id integrationID) String() string {
	return fmt.Sprintf("%s/%s", id.Name, id.Identifier)
}

// UpdateController updates the Controller with new Controller and
// IntegrationOptions.
//
// UpdateController updates running integrations. Extensions can be
// recalculated by calling relevant methods like Handler or Targets.
func (c *Controller) UpdateController(cfg V2Integrations, globals v2.Globals) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	integrationIDMap := map[integrationID]struct{}{}

	integrations := make([]*controlledIntegration, 0, len(cfg.ActiveConfigs()))

NextConfig:
	for _, ic := range cfg.ActiveConfigs() {
		name := ic.Name()

		identifier, err := ic.Identifier(globals)
		if err != nil {
			return fmt.Errorf("could not build identifier for integration %q: %w", name, err)
		}

		if err := ic.ApplyDefaults(globals); err != nil {
			return fmt.Errorf("failed to apply defaults for %s/%s: %w", name, identifier, err)
		}

		id := integrationID{Name: name, Identifier: identifier}
		if _, exist := integrationIDMap[id]; exist {
			return fmt.Errorf("multiple instance names %q in integration %q", identifier, name)
		}
		integrationIDMap[id] = struct{}{}

		// Now that we know the ID for an integration, we can check to see if it's
		// running and can be dynamically updated.
		for _, ci := range c.integrations {
			if ci.id != id {
				continue
			}

			// If the configs haven't changed, then we don't need to do anything.

			if v2.CompareConfigs(ci.c, ic) {
				integrations = append(integrations, ci)
				continue NextConfig
			}

			if ui, ok := ci.i.(v2.UpdateIntegration); ok {
				if err := ui.ApplyConfig(ic, globals); errors.Is(err, v2.ErrInvalidUpdate) {
					level.Warn(c.logger).Log("msg", "failed to dynamically update integration; will recreate", "integration", name, "instance", identifier, "err', err")
					break
				} else if err != nil {
					return fmt.Errorf("failed to update %s integration %q: %w", name, identifier, err)
				} else {
					// Update succeeded; re-use the running one and go to the next
					// integration to process.
					integrations = append(integrations, ci)
					continue NextConfig
				}
			}

			// We found the integration to update: we can stop this loop now.
			break
		}

		logger := log.With(c.logger, "integration", name, "instance", identifier)
		integration, err := ic.NewIntegration(logger, globals)
		if err != nil {
			return fmt.Errorf("failed to construct %s integration %q: %w", name, identifier, err)
		}

		// Create a new controlled integration.
		integrations = append(integrations, &controlledIntegration{
			id:  id,
			gen: c.gen.Inc(),
			i:   integration,
			c:   ic,
		})
	}

	// Update integrations and inform
	c.integrations = integrations
	c.reloadIntegrations <- struct{}{}

	c.cfg = cfg
	c.globals = globals
	return nil
}

// Handler returns an HTTP handler for the controller and its integrations.
// Handler will pass through requests to other running integrations. Handler
// always returns an http.Handler regardless of error.
//
// Handler is expensive to compute and should only be done after reloading the
// shared.
func (c *Controller) Handler(prefix string) (http.Handler, error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	var firstErr error
	saveFirstErr := func(err error) {
		if firstErr == nil {
			firstErr = err
		}
	}

	r := mux.NewRouter()

	err := forEachIntegration(c.integrations, prefix, func(ci *controlledIntegration, iprefix string) {
		id := ci.id

		i, ok := ci.i.(v2.HTTPIntegration)
		if !ok {
			return
		}

		handler, err := i.Handler(iprefix + "/")
		if err != nil {
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
	if err != nil {
		level.Warn(c.logger).Log("msg", "error when iterating over integrations to build HTTP handlers", "err", err)
	}

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

// Targets returns the current set of targets across all integrations. Use opts
// to customize which targets are returned.
func (c *Controller) Targets(ep v2.Endpoint, opts TargetOptions) []*TargetGroup {
	// Grab the integrations as fast as possible. We don't want to spend too much
	// time holding the mutex.
	type prefixedMetricsIntegration struct {
		id integrationID
		i  v2.MetricsIntegration
		ep v2.Endpoint
	}
	var mm []prefixedMetricsIntegration

	c.mut.Lock()
	err := forEachIntegration(c.integrations, ep.Prefix, func(ci *controlledIntegration, iprefix string) {
		// Best effort liveness check. They might stop running when we request
		// their targets, which is fine, but we should save as much work as we
		// can.
		if !ci.Running() {
			return
		}
		if mi, ok := ci.i.(v2.MetricsIntegration); ok {
			ep := v2.Endpoint{Host: ep.Host, Prefix: iprefix}
			mm = append(mm, prefixedMetricsIntegration{id: ci.id, i: mi, ep: ep})
		}
	})
	if err != nil {
		level.Warn(c.logger).Log("msg", "error when iterating over integrations to get targets", "err", err)
	}
	c.mut.Unlock()

	var tgs []*TargetGroup
	for _, mi := range mm {
		// If we're looking for a subset of integrations, filter out anything that doesn't match.
		if len(opts.Integrations) > 0 && !stringSliceContains(opts.Integrations, mi.id.Name) {
			continue
		}
		// If we're looking for a specific instance, filter out anything that doesn't match.
		if opts.Instance != "" && mi.id.Identifier != opts.Instance {
			continue
		}

		for _, tgt := range mi.i.Targets(mi.ep) {
			tgs = append(tgs, (*TargetGroup)(tgt))
		}
	}
	sort.Slice(tgs, func(i, j int) bool {
		return tgs[i].Source < tgs[j].Source
	})
	return tgs
}

func stringSliceContains(ss []string, s string) bool {
	for _, check := range ss {
		if check == s {
			return true
		}
	}
	return false
}

// TargetOptions controls which targets should be returned by the subsystem.
type TargetOptions struct {
	// Integrations is the set of integrations to return. An empty slice will
	// default to returning all integrations.
	Integrations []string
	// Instance matches a specific instance from all integrations. An empty
	// string will match any instance.
	Instance string
}

// TargetOptionsFromParams creates TargetOptions from parsed URL query parameters.
func TargetOptionsFromParams(u url.Values) (TargetOptions, error) {
	var to TargetOptions

	rawIntegrations := u.Get("integrations")
	if rawIntegrations != "" {
		rawIntegrations, err := url.QueryUnescape(rawIntegrations)
		if err != nil {
			return to, fmt.Errorf("invalid value for integrations: %w", err)
		}
		to.Integrations = strings.Split(rawIntegrations, ",")
	}

	rawInstance := u.Get("instance")
	if rawInstance != "" {
		rawInstance, err := url.QueryUnescape(rawInstance)
		if err != nil {
			return to, fmt.Errorf("invalid value for instance: %w", err)
		}
		to.Instance = rawInstance
	}

	return to, nil
}

// ToParams will convert to into URL query parameters.
func (to TargetOptions) ToParams() url.Values {
	p := make(url.Values)
	if len(to.Integrations) != 0 {
		p.Set("integrations", url.QueryEscape(strings.Join(to.Integrations, ",")))
	}
	if to.Instance != "" {
		p.Set("instance", url.QueryEscape(to.Instance))
	}
	return p
}

// ScrapeConfigs returns a set of scrape configs to use for self-scraping.
// sdConfig should contain the full URL where the integrations SD API is
// exposed. ScrapeConfigs will inject unique query parameters per integration
// to limit what will be discovered.
func (c *Controller) ScrapeConfigs(prefix string, sdConfig *http_sd.SDConfig) []*autoscrape.ScrapeConfig {
	// Grab the integrations as fast as possible. We don't want to spend too much
	// time holding the mutex.
	type prefixedMetricsIntegration struct {
		id     integrationID
		i      v2.MetricsIntegration
		prefix string
	}
	var mm []prefixedMetricsIntegration

	c.mut.Lock()
	err := forEachIntegration(c.integrations, prefix, func(ci *controlledIntegration, iprefix string) {
		if mi, ok := ci.i.(v2.MetricsIntegration); ok {
			mm = append(mm, prefixedMetricsIntegration{id: ci.id, i: mi, prefix: iprefix})
		}
	})
	if err != nil {
		level.Warn(c.logger).Log("msg", "error when iterating over integrations to get scrape configs", "err", err)
	}
	c.mut.Unlock()

	var cfgs []*autoscrape.ScrapeConfig
	for _, mi := range mm {
		// sdConfig will be pointing to the targets API. By default, this returns absolutely everything.
		// We want to use the query parmaeters to inform the API to only return
		// specific targets.
		opts := TargetOptions{
			Integrations: []string{mi.id.Name},
			Instance:     mi.id.Identifier,
		}

		integrationSDConfig := *sdConfig
		integrationSDConfig.URL = sdConfig.URL + "?" + opts.ToParams().Encode()
		sds := discovery.Configs{&integrationSDConfig}
		cfgs = append(cfgs, mi.i.ScrapeConfigs(sds)...)
	}
	sort.Slice(cfgs, func(i, j int) bool {
		return cfgs[i].Config.JobName < cfgs[j].Config.JobName
	})
	return cfgs
}
