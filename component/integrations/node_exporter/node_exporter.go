package node_exporter

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/metrics/scrape"
	node_integration "github.com/grafana/agent/pkg/integrations/node_exporter"
	"github.com/prometheus/common/model"
)

func init() {
	component.Register(component.Registration{
		Name:    "integration.node_exporter",
		Args:    node_integration.Config{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewComponent(opts, args.(node_integration.Config))
		},
	})
}

// Exports are simply a list of targets for a scraper to consume
type Exports struct {
	Output []scrape.Target `river:"output,attr"`
}

type Component struct {
	log  log.Logger
	opts component.Options

	mut sync.RWMutex
	cfg *node_integration.Config

	integration *node_integration.Integration
}

func NewComponent(o component.Options, args node_integration.Config) (*Component, error) {
	c := &Component{
		log:  o.Logger,
		cfg:  &args,
		opts: o,
	}

	// Call to Update() to set the output once at the start
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	c.log.Log("Msg", "Running")
	<-ctx.Done()
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.log.Log("Msg", "Update")
	var err error
	c.integration, err = node_integration.New(c.log, c.cfg)
	targets := []scrape.Target{{
		model.AddressLabel:     "127.0.0.1:12345",
		model.SchemeLabel:      "http",
		model.MetricsPathLabel: fmt.Sprintf("/component/%s/metrics", c.opts.ID),
		"name":                 "node_exporter",
	}}
	c.opts.OnStateChange(Exports{
		Output: targets,
	})
	return err
}

func (c *Component) Handler() http.Handler {
	if c.integration != nil {
		h, err := c.integration.MetricsHandler()
		if err != nil {
			c.log.Log(fmt.Errorf("Creating metrics handler: %e", err))
		}
		return h
	}
	return nil
}
