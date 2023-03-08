package podmonitors

import (
	"context"
	"sync"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/prometheus/prometheus/discovery"
)

func init() {
	component.Register(component.Registration{
		Name: "prometheus.kubernetes.podmonitors",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args)
		},
	})
}

type Component struct {
	opts   component.Options
	config *Arguments

	onUpdate chan struct{}
	mut      sync.Mutex
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

	disc := discovery.NewManager(ctx, c.opts.Logger, discovery.Name(c.opts.ID))
	go func() {
		err := disc.Run()
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "discovery manager stopped", "err", err)
			// very unhelathy
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

		case <-c.onUpdate:
			if cancel != nil {
				cancel()
			}
			innerCtx, cancel = context.WithCancel(ctx)
			c.mut.Lock()
			componentCfg := c.config
			c.mut.Unlock()
			disc.ApplyConfig(nil)
			crdMan := newManager(c.opts, c.opts.Logger, componentCfg)
			go func() {
				if err := crdMan.run(innerCtx, disc); err != nil {
					level.Error(c.opts.Logger).Log("msg", "error running crd manager", "err", err)
				}
			}()
		}
	}
}

// Update implements component.Compnoent.
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
