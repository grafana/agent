package syslog

import (
	"context"
	"reflect"
	"sync"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	st "github.com/grafana/agent/component/loki/source/syslog/internal/syslogtarget"
	"github.com/prometheus/prometheus/model/relabel"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.syslog",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the loki.source.syslog
// component.
type Arguments struct {
	SyslogListeners []ListenerConfig    `river:"listener,block"`
	ForwardTo       []loki.LogsReceiver `river:"forward_to,attr"`
	RelabelRules    *flow_relabel.Rules `river:"relabel_rules,attr,optional"`
}

// Component implements the loki.source.syslog component.
type Component struct {
	opts    component.Options
	metrics *st.Metrics

	mut       sync.RWMutex
	lc        []ListenerConfig
	fanout    []loki.LogsReceiver
	listeners []*st.SyslogTarget

	handler loki.LogsReceiver
}

// New creates a new loki.source.syslog component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:    o,
		metrics: st.NewMetrics(o.Registerer),
		handler: make(loki.LogsReceiver),
		fanout:  args.ForwardTo,

		listeners: []*st.SyslogTarget{},
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
		level.Info(c.opts.Logger).Log("msg", "loki.source.syslog component shutting down, stopping listeners")
		for _, l := range c.listeners {
			err := l.Stop()
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error while stopping syslog listener", "err", err)
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
	if newArgs.RelabelRules != nil && len(newArgs.RelabelRules.GetAll()) > 0 {
		rcs = flow_relabel.ComponentToPromRelabelConfigs(newArgs.RelabelRules.GetAll())
	}

	if configsChanged(c.lc, newArgs.SyslogListeners) {
		for _, l := range c.listeners {
			err := l.Stop()
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error while stopping syslog listener", "err", err)
			}
		}
		c.listeners = make([]*st.SyslogTarget, 0)
		entryHandler := loki.NewEntryHandler(c.handler, func() {})

		for _, cfg := range newArgs.SyslogListeners {
			t, err := st.NewSyslogTarget(c.metrics, c.opts.Logger, entryHandler, rcs, cfg.Convert())
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "failed to create syslog listener with provided config", "err", err)
				continue
			}
			c.listeners = append(c.listeners, t)
		}
	}

	return nil
}

// DebugInfo returns information about the status of listeners.
func (c *Component) DebugInfo() interface{} {
	var res readerDebugInfo

	for _, t := range c.listeners {
		res.ListenersInfo = append(res.ListenersInfo, listenerInfo{
			Type:          string(t.Type()),
			Ready:         t.Ready(),
			ListenAddress: t.ListenAddress().String(),
			Labels:        t.Labels().String(),
		})
	}
	return res
}

type readerDebugInfo struct {
	ListenersInfo []listenerInfo `river:"listeners_info,attr"`
}

type listenerInfo struct {
	Type          string `river:"type,attr"`
	Ready         bool   `river:"ready,attr"`
	ListenAddress string `river:"listen_address,attr"`
	Labels        string `river:"labels,attr"`
}

func configsChanged(prev, next []ListenerConfig) bool {
	if len(prev) != len(next) {
		return true
	}
	for i := range prev {
		if !reflect.DeepEqual(prev[i], next[i]) {
			return true
		}
	}
	return false
}
