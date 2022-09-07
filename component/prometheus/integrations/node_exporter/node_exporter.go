package node_exporter

import (
	"context"
	"net/http"
	"path"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	node_integration "github.com/grafana/agent/pkg/integrations/node_exporter"
	"github.com/prometheus/common/model"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.integration.node_exporter",
		Args:    node_integration.Config{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewComponent(opts, args.(node_integration.Config))
		},
	})
}

// Exports are simply a list of targets for a scraper to consume
type Exports struct {
	Targets []discovery.Target `river:"targets,attr"`
}

// Component for node_exporter.
type Component struct {
	log  log.Logger
	opts component.Options

	integration *node_integration.Integration

	mut sync.Mutex
}

// NewComponent creates a node_exporter component
func NewComponent(o component.Options, args node_integration.Config) (*Component, error) {
	c := &Component{
		log:  o.Logger,
		opts: o,
		mut:  sync.Mutex{},
	}

	// Call to Update() to set the output once at the start
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	var err error
	cfg := args.(node_integration.Config)
	c.mut.Lock()
	c.integration, err = node_integration.New(c.log, &cfg)
	c.mut.Unlock()
	if err != nil {
		return err
	}

	targets := []discovery.Target{{
		model.AddressLabel:     c.opts.HTTPListenAddr,
		model.SchemeLabel:      "http",
		model.MetricsPathLabel: path.Join(c.opts.HTTPPath, "metrics"),
		"name":                 "node_exporter",
	}}
	c.opts.OnStateChange(Exports{
		Targets: targets,
	})
	return err
}

// Handler serves node_exporter metrics endpoint
func (c *Component) Handler() http.Handler {
	c.mut.Lock()
	defer c.mut.Unlock()
	h, err := c.integration.MetricsHandler()
	if err != nil {
		level.Error(c.log).Log("msg", "failed to creating metrics handler", "err", err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
	}
	return h
}
