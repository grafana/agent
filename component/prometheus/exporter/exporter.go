package exporter

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/prometheus/common/model"
)

// Creator is a function provided by an implementation to create a concrete exporter instance.
type Creator func(component.Options, component.Arguments) (integrations.Integration, error)

// Exports are simply a list of targets for a scraper to consume.
type Exports struct {
	Targets []discovery.Target `river:"targets,attr"`
}

type Component struct {
	opts component.Options

	mut sync.Mutex

	reload chan struct{}

	creator         Creator
	multiTargetFunc func(discovery.Target, component.Arguments) []discovery.Target
	baseTarget      discovery.Target

	exporter       integrations.Integration
	metricsHandler http.Handler
}

// New creates a new exporter component.
func New(creator Creator, name string, instance string) func(component.Options, component.Arguments) (component.Component, error) {
	return newExporter(creator, name, instance, nil)
}

// NewMultiTarget creates a new exporter component that supports multiple targets.
func NewMultiTarget(creator Creator, name string, instance string, multiTargetFunc func(discovery.Target, component.Arguments) []discovery.Target) func(component.Options, component.Arguments) (component.Component, error) {
	return newExporter(creator, name, instance, multiTargetFunc)
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	var cancel context.CancelFunc
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.reload:
			// cancel any previously running exporter
			if cancel != nil {
				cancel()
			}
			// create new context so we can cancel it if we get any future updates
			// since it is derived from the main run context, it only needs to be
			// canceled directly if we receive new updates
			newCtx, cancelFunc := context.WithCancel(ctx)
			cancel = cancelFunc

			// finally create and run new exporter
			c.mut.Lock()
			exporter := c.exporter
			c.metricsHandler = c.getHttpHandler(exporter)
			c.mut.Unlock()
			go func() {
				if err := exporter.Run(newCtx); err != nil {
					level.Error(c.opts.Logger).Log("msg", "error running exporter", "err", err)
				}
			}()
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	exporter, err := c.creator(c.opts, args)
	if err != nil {
		return err
	}
	c.mut.Lock()
	c.exporter = exporter

	var targets []discovery.Target
	if c.multiTargetFunc == nil {
		targets = []discovery.Target{c.baseTarget}
	} else {
		targets = c.multiTargetFunc(c.baseTarget, args)
	}

	c.opts.OnStateChange(Exports{
		Targets: targets,
	})
	c.mut.Unlock()
	select {
	case c.reload <- struct{}{}:
	default:
	}
	return err
}

// Handler serves metrics endpoint from the integration implementation.
func (c *Component) Handler() http.Handler {
	c.mut.Lock()
	defer c.mut.Unlock()
	return c.metricsHandler
}

func newExporter(creator Creator, name string, instance string, multiTargetFunc func(discovery.Target, component.Arguments) []discovery.Target) func(component.Options, component.Arguments) (component.Component, error) {
	return func(opts component.Options, args component.Arguments) (component.Component, error) {
		c := &Component{
			opts:            opts,
			reload:          make(chan struct{}, 1),
			creator:         creator,
			multiTargetFunc: multiTargetFunc,
		}
		jobName := fmt.Sprintf("integrations/%s", name)
		if instance == "" {
			instance = defaultInstance()
		}

		c.baseTarget = discovery.Target{
			model.AddressLabel:                  opts.HTTPListenAddr,
			model.SchemeLabel:                   "http",
			model.MetricsPathLabel:              path.Join(opts.HTTPPath, "metrics"),
			"instance":                          instance,
			"job":                               jobName,
			"__meta_agent_integration_name":     jobName,
			"__meta_agent_integration_instance": instance,
			"__meta_agent_integration_id":       opts.ID,
		}

		// Call to Update() to set the output once at the start.
		if err := c.Update(args); err != nil {
			return nil, err
		}

		return c, nil
	}
}

// get the http handler once and save it, so we don't create extra garbage
func (c *Component) getHttpHandler(integration integrations.Integration) http.Handler {
	h, err := integration.MetricsHandler()
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "failed to creating metrics handler", "err", err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
	}
	return h
}

// defaultInstance retrieves the hostname identifying the machine the process is
// running on. It will return the value of $HOSTNAME, if defined, and fall
// back to Go's os.Hostname. If that fails, it will return "unknown".
func defaultInstance() string {
	hostname := os.Getenv("HOSTNAME")
	if hostname != "" {
		return hostname
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
