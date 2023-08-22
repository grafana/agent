package common

import (
	"context"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/operator"
)

type Component struct {
	mut     sync.RWMutex
	config  *operator.Arguments
	manager *crdManager

	onUpdate  chan struct{}
	opts      component.Options
	healthMut sync.RWMutex
	health    component.Health

	kind string
}

func New(o component.Options, args component.Arguments, kind string) (*Component, error) {
	c := &Component{
		opts:     o,
		onUpdate: make(chan struct{}, 1),
		kind:     kind,
	}
	return c, c.Update(args)
}

func (c *Component) CurrentHealth() component.Health {
	c.healthMut.RLock()
	defer c.healthMut.RUnlock()
	return c.health
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

	c.reportHealth(nil)
	errChan := make(chan error, 1)
	for {
		select {
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}
			return nil
		case err := <-errChan:
			c.reportHealth(err)
		case <-c.onUpdate:
			c.mut.Lock()
			manager := newCrdManager(c.opts, c.opts.Logger, c.config, c.kind)
			c.manager = manager
			if cancel != nil {
				cancel()
			}
			innerCtx, cancel = context.WithCancel(ctx)
			go func() {
				if err := manager.Run(innerCtx); err != nil {
					level.Error(c.opts.Logger).Log("msg", "error running crd manager", "err", err)
					errChan <- err
				}
			}()
			c.mut.Unlock()
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	// TODO(jcreixell): Initialize manager here so we can return errors back early to the caller.
	// See https://github.com/grafana/agent/pull/2688#discussion_r1152384425
	c.mut.Lock()
	cfg := args.(operator.Arguments)
	c.config = &cfg
	c.mut.Unlock()
	select {
	case c.onUpdate <- struct{}{}:
	default:
	}
	return nil
}

// NotifyClusterChange implements component.ClusterComponent.
func (c *Component) NotifyClusterChange() {
	c.mut.RLock()
	defer c.mut.RUnlock()

	if !c.config.Clustering.Enabled {
		return // no-op
	}

	if c.manager != nil {
		c.manager.ClusteringUpdated()
	}
}

// DebugInfo returns debug information for this component.
func (c *Component) DebugInfo() interface{} {
	return c.manager.DebugInfo()
}

func (c *Component) reportHealth(err error) {
	c.healthMut.Lock()
	defer c.healthMut.Unlock()

	if err != nil {
		c.health = component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    err.Error(),
			UpdateTime: time.Now(),
		}
		return
	} else {
		c.health = component.Health{
			Health:     component.HealthTypeHealthy,
			UpdateTime: time.Now(),
		}
	}
}
