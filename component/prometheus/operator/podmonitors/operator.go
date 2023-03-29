package podmonitors

import (
	"context"
	"sync"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/scrape"
)

func init() {
	component.Register(component.Registration{
		Name: "prometheus.operator.podmonitors",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args)
		},
	})
}

type Component struct {
	mut     sync.Mutex
	config  *Arguments
	manager *CRDManager

	onUpdate chan struct{}
	opts     component.Options
}

func New(o component.Options, args component.Arguments) (*Component, error) {
	c := &Component{
		opts:     o,
		onUpdate: make(chan struct{}, 1),
	}
	return c, c.Update(args)
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	// innerCtx gets passed to things we create, so we can restart everything anytime we get an update.
	// Ideally, this component has very little dynamic config, and won't have frequent updates.
	var innerCtx context.Context
	// cancel is the func we use to trigger a stop to all downstream processors we create
	var cancel func()
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()

	errChan := make(chan error)
	for {
		select {
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}
			return nil
		case <-errChan:
			// TODO(jcreixell): Mark component as unhealthy here?
		case <-c.onUpdate:
			if cancel != nil {
				cancel()
			}
			innerCtx, cancel = context.WithCancel(ctx)
			c.mut.Lock()
			componentCfg := c.config
			manager := NewCRDManager(c.opts, c.opts.Logger, componentCfg)
			c.manager = manager
			c.mut.Unlock()
			go func() {
				if err := manager.Run(innerCtx); err != nil {
					level.Error(c.opts.Logger).Log("msg", "error running crd manager", "err", err)
					// TODO: anything else we need to do here? (component unhealthy?)
				}
			}()
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	cfg := args.(Arguments)
	c.config = &cfg
	c.mut.Unlock()
	select {
	case c.onUpdate <- struct{}{}:
	default:
	}
	return nil
}

// DebugInfo returns debug information for this component.
func (c *Component) DebugInfo() interface{} {
	var info DebugInfo
	for _, pm := range c.manager.DebugInfo {
		info.DiscoveredPodMonitors = append(info.DiscoveredPodMonitors, pm)
	}
	info.Targets = scrape.BuildTargetStatuses(c.manager.ScrapeManager.TargetsActive())
	return info
}
