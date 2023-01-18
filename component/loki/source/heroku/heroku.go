package heroku

import (
	"context"
	"reflect"
	"sync"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	st "github.com/grafana/agent/component/loki/source/heroku/internal/herokutarget"
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
	HerokuListener ListenerConfig      `river:"listener,block"`
	ForwardTo      []loki.LogsReceiver `river:"forward_to,attr"`
	RelabelRules   flow_relabel.Rules  `river:"relabel_rules,attr,optional"`
}

// Component implements the loki.source.heroku component.
type Component struct {
	opts    component.Options
	metrics *st.Metrics

	mut      sync.RWMutex
	lc       ListenerConfig
	fanout   []loki.LogsReceiver
	listener *st.HerokuTarget

	handler loki.LogsReceiver
}

// New creates a new loki.source.heroku component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:     o,
		metrics:  st.NewMetrics(o.Registerer),
		mut:      sync.RWMutex{},
		lc:       ListenerConfig{},
		fanout:   args.ForwardTo,
		listener: nil,
		handler:  make(loki.LogsReceiver),
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
		level.Info(c.opts.Logger).Log("msg", "loki.source.heroku component shutting down, stopping listeners")
		if c.listener != nil {
			err := c.listener.Stop()
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
		if c.listener != nil {
			err := c.listener.Stop()
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error while stopping heroku listener", "err", err)
			}
		}

		entryHandler := loki.NewEntryHandler(c.handler, func() {})
		t, err := st.NewHerokuTarget(c.metrics, c.opts.Logger, entryHandler, "job_name_todo", rcs, newArgs.HerokuListener.Convert())
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "failed to create heroku listener with provided config", "err", err)
			return err
		}

		c.listener = t
	}

	return nil
}

// DebugInfo returns information about the status of listeners.
func (c *Component) DebugInfo() interface{} {
	var res readerDebugInfo

	res.ListenersInfo = listenerInfo{
		Type:          string(c.listener.Type()),
		Ready:         c.listener.Ready(),
		ListenAddress: c.listener.ListenAddress(),
		ListenPort:    c.listener.ListenPort(),
		Labels:        c.listener.Labels().String(),
	}

	return res
}

type readerDebugInfo struct {
	ListenersInfo listenerInfo `river:"listeners_info,attr"`
}

type listenerInfo struct {
	Type          string `river:"type,attr"`
	Ready         bool   `river:"ready,attr"`
	ListenAddress string `river:"listen_address,attr"`
	ListenPort    int    `river:"listen_port,attr"`
	Labels        string `river:"labels,attr"`
}

func configsChanged(prev, next ListenerConfig) bool {
	return !reflect.DeepEqual(prev, next)
}
