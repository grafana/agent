package gcplog

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/relabel"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	gt "github.com/grafana/agent/component/loki/source/gcplog/internal/gcplogtarget"
	"github.com/grafana/agent/pkg/util"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.gcplog",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the loki.source.gcplog
// component.
type Arguments struct {
	// TODO(@tpaschalis) Having these types defined in an internal package
	// means that an external caller cannot build this component's Arguments
	// by hand for now.
	PullTarget   *gt.PullConfig      `river:"pull,block,optional"`
	PushTarget   *gt.PushConfig      `river:"push,block,optional"`
	ForwardTo    []loki.LogsReceiver `river:"forward_to,attr"`
	RelabelRules flow_relabel.Rules  `river:"relabel_rules,attr,optional"`
}

// UnmarshalRiver implements the unmarshaller
func (a *Arguments) UnmarshalRiver(f func(v interface{}) error) error {
	*a = Arguments{}
	type arguments Arguments
	err := f((*arguments)(a))
	if err != nil {
		return err
	}

	if (a.PullTarget != nil) == (a.PushTarget != nil) {
		return fmt.Errorf("exactly one of 'push' or 'pull' must be provided")
	}
	return nil
}

// Component implements the loki.source.gcplog component.
type Component struct {
	opts          component.Options
	metrics       *gt.Metrics
	serverMetrics *util.UncheckedCollector

	mut    sync.RWMutex
	fanout []loki.LogsReceiver
	target gt.Target

	handler loki.LogsReceiver
}

// New creates a new loki.source.gcplog component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:          o,
		metrics:       gt.NewMetrics(o.Registerer),
		handler:       make(loki.LogsReceiver),
		fanout:        args.ForwardTo,
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
		_ = level.Info(c.opts.Logger).Log("msg", "loki.source.gcplog component shutting down, stopping the targets")
		c.mut.RLock()
		err := c.target.Stop()
		if err != nil {
			_ = level.Error(c.opts.Logger).Log("msg", "error while stopping gcplog target", "err", err)
		}
		c.mut.RUnlock()
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

	if c.target != nil {
		err := c.target.Stop()
		if err != nil {
			_ = level.Error(c.opts.Logger).Log("msg", "error while stopping gcplog target", "err", err)
		}
	}
	entryHandler := loki.NewEntryHandler(c.handler, func() {})
	jobName := strings.Replace(c.opts.ID, ".", "_", -1)

	if newArgs.PullTarget != nil {
		// TODO(@tpaschalis) Are there any options from "google.golang.org/api/option"
		// we should expose as configuration and pass here?
		t, err := gt.NewPullTarget(c.metrics, c.opts.Logger, entryHandler, jobName, newArgs.PullTarget, rcs)
		if err != nil {
			_ = level.Error(c.opts.Logger).Log("msg", "failed to create gcplog target with provided config", "err", err)
			return err
		}
		c.target = t
	}
	if newArgs.PushTarget != nil {
		// [gt.NewPushTarget] registers new metrics every time it is called. To
		// avoid issues with re-registering metrics with the same name, we create a
		// new registry for the target every time we create one, and pass it to an
		// unchecked collector to bypass uniqueness checking.
		registry := prometheus.NewRegistry()
		c.serverMetrics.SetCollector(registry)

		t, err := gt.NewPushTarget(c.metrics, c.opts.Logger, entryHandler, jobName, newArgs.PushTarget, rcs, registry)
		if err != nil {
			_ = level.Error(c.opts.Logger).Log("msg", "failed to create gcplog target with provided config", "err", err)
			return err
		}
		c.target = t
	}

	return nil
}

// DebugInfo returns information about the status of targets.
func (c *Component) DebugInfo() interface{} {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return targetDebugInfo{Details: c.target.Details()}
}

type targetDebugInfo struct {
	Details map[string]string `river:"target_info,attr"`
}
