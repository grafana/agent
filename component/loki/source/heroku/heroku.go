package heroku

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	ht "github.com/grafana/agent/component/loki/source/heroku/internal/herokutarget"
	"github.com/prometheus/prometheus/model/relabel"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.heroku",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the loki.source.heroku
// component.
type Arguments struct {
	HerokuListener       ListenerConfig      `river:"listener,block"`
	Labels               map[string]string   `river:"labels,attr,optional"`
	UseIncomingTimestamp bool                `river:"use_incoming_timestamp,attr,optional"`
	ForwardTo            []loki.LogsReceiver `river:"forward_to,attr"`
	RelabelRules         flow_relabel.Rules  `river:"relabel_rules,attr,optional"`
}

// Component implements the loki.source.heroku component.
type Component struct {
	opts    component.Options
	metrics *ht.Metrics

	mut    sync.RWMutex
	lc     ListenerConfig
	fanout []loki.LogsReceiver
	target *ht.HerokuTarget

	handler loki.LogsReceiver
}

// New creates a new loki.source.heroku component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:    o,
		metrics: ht.NewMetrics(o.Registerer),
		mut:     sync.RWMutex{},
		lc:      ListenerConfig{},
		fanout:  args.ForwardTo,
		target:  nil,
		handler: make(loki.LogsReceiver),
	}

	// Call to Update() to start readers and set receivers once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	defer func() {
		c.mut.Lock()
		defer c.mut.Unlock()

		level.Info(c.opts.Logger).Log("msg", "loki.source.heroku component shutting down, stopping listener")
		if c.target != nil {
			err := c.target.Stop()
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error while stopping heroku listener", "err", err)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.handler:
			c.mut.RLock()
			for _, receiver := range c.fanout {
				receiver <- entry
			}
			c.mut.RUnlock()
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)
	c.fanout = newArgs.ForwardTo

	var rcs []*relabel.Config
	if newArgs.RelabelRules != nil {
		rcs = flow_relabel.ComponentToPromRelabelConfigs(newArgs.RelabelRules())
	}

	if configsChanged(c.lc, newArgs.HerokuListener) {
		if c.target != nil {
			err := c.target.Stop()
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error while stopping heroku listener", "err", err)
			}
		}

		entryHandler := loki.NewEntryHandler(c.handler, func() {})
		t, err := ht.NewHerokuTarget(c.metrics, c.opts.Logger, entryHandler, rcs, newArgs.Convert(), c.opts.Registerer)
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "failed to create heroku listener with provided config", "err", err)
			return err
		}

		c.target = t
	}

	return nil
}

// DebugInfo returns information about the status of listener.
func (c *Component) DebugInfo() interface{} {
	var res readerDebugInfo = readerDebugInfo{
		Ready:   c.target.Ready(),
		Address: fmt.Sprintf("%s:%d", c.target.ListenAddress(), c.target.ListenPort()),
	}

	return res
}

type readerDebugInfo struct {
	Ready   bool   `river:"ready,attr"`
	Address string `river:"address,attr"`
}

func configsChanged(prev, next ListenerConfig) bool {
	return !reflect.DeepEqual(prev, next)
}
