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
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	sv "github.com/weaveworks/common/server"
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

// ListenerConfig defines a heroku listener.
type ListenerConfig struct {
	ListenAddress string `river:"address,attr,optional"`
	ListenPort    int    `river:"port,attr"`
	// TODO - add the rest of the server config from Promtail
}

// DefaultListenerConfig provides the default arguments for a heroku listener.
var DefaultListenerConfig = ListenerConfig{
	ListenAddress: "0.0.0.0",
}

// UnmarshalRiver implements river.Unmarshaler.
func (lc *ListenerConfig) UnmarshalRiver(f func(interface{}) error) error {
	*lc = DefaultListenerConfig

	type herokucfg ListenerConfig
	err := f((*herokucfg)(lc))
	if err != nil {
		return err
	}

	return nil
}

// Component implements the loki.source.heroku component.
type Component struct {
	opts    component.Options
	metrics *ht.Metrics

	mut           sync.RWMutex
	args          Arguments
	fanout        []loki.LogsReceiver
	target        *ht.HerokuTarget
	serverMetrics *util.UncheckedCollector

	handler loki.LogsReceiver
}

// New creates a new loki.source.heroku component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:          o,
		metrics:       ht.NewMetrics(o.Registerer),
		mut:           sync.RWMutex{},
		args:          Arguments{},
		fanout:        args.ForwardTo,
		target:        nil,
		handler:       make(loki.LogsReceiver),
		serverMetrics: util.NewUncheckedCollector(nil),
	}

	o.Registerer.MustRegister(c.serverMetrics)

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
	if newArgs.RelabelRules != nil && len(newArgs.RelabelRules) > 0 {
		rcs = flow_relabel.ComponentToPromRelabelConfigs(newArgs.RelabelRules)
	}

	if listenerChanged(c.args.HerokuListener, newArgs.HerokuListener) || relabelRulesChanged(c.args.RelabelRules, newArgs.RelabelRules) {
		if c.target != nil {
			err := c.target.Stop()
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error while stopping heroku listener", "err", err)
			}
		}

		// [ht.NewHerokuTarget] registers new metrics every time it is called. To
		// avoid issues with re-registering metrics with the same name, we create a
		// new registry for the target every time we create one, and pass it to an
		// unchecked collector to bypass uniqueness checking.
		registry := prometheus.NewRegistry()
		c.serverMetrics.SetCollector(serverMetrics)

		entryHandler := loki.NewEntryHandler(c.handler, func() {})
		t, err := ht.NewHerokuTarget(c.metrics, c.opts.Logger, entryHandler, rcs, newArgs.Convert(), registry)
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "failed to create heroku listener with provided config", "err", err)
			return err
		}

		c.target = t
		c.args = newArgs
	}

	return nil
}

// Convert is used to bridge between the River and Promtail types.
func (args *Arguments) Convert() *ht.HerokuDrainTargetConfig {
	lbls := make(model.LabelSet, len(args.Labels))
	for k, v := range args.Labels {
		lbls[model.LabelName(k)] = model.LabelValue(v)
	}

	return &ht.HerokuDrainTargetConfig{
		Server: sv.Config{
			HTTPListenAddress: args.HerokuListener.ListenAddress,
			HTTPListenPort:    args.HerokuListener.ListenPort,
		},
		Labels:               lbls,
		UseIncomingTimestamp: args.UseIncomingTimestamp,
	}
}

// DebugInfo returns information about the status of listener.
func (c *Component) DebugInfo() interface{} {
	c.mut.RLock()
	defer c.mut.RUnlock()

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

func listenerChanged(prev, next ListenerConfig) bool {
	return !reflect.DeepEqual(prev, next)
}
func relabelRulesChanged(prev, next flow_relabel.Rules) bool {
	return !reflect.DeepEqual(prev, next)
}
