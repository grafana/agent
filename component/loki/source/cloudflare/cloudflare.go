package cloudflare

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/loki/positions"
	cft "github.com/grafana/agent/component/loki/source/cloudflare/internal/cloudflaretarget"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.cloudflare",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the
// loki.source.cloudflare component.
type Arguments struct {
	Config    Config              `river:"config,block"`
	ForwardTo []loki.LogsReceiver `river:"forward_to,attr"`
}

// Component implements the loki.source.cloudflare component.
type Component struct {
	opts    component.Options
	metrics *cft.Metrics

	mut    sync.RWMutex
	fanout []loki.LogsReceiver
	target *cft.Target

	posFile positions.Positions
	handler loki.LogsReceiver
}

// New creates a new loki.source.cloudflare component.
func New(o component.Options, args Arguments) (*Component, error) {
	err := os.MkdirAll(o.DataPath, 0750)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	positionsFile, err := positions.New(o.Logger, positions.Config{
		SyncPeriod:        10 * time.Second,
		PositionsFile:     filepath.Join(o.DataPath, "positions.yml"),
		IgnoreInvalidYaml: false,
		ReadOnly:          false,
	})
	if err != nil {
		return nil, err
	}

	c := &Component{
		opts:    o,
		metrics: cft.NewMetrics(o.Registerer),
		handler: make(loki.LogsReceiver),
		fanout:  args.ForwardTo,
		posFile: positionsFile,
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
		level.Info(c.opts.Logger).Log("msg", "loki.source.cloudflare component shutting down, stopping the target")
		c.target.Stop()
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

	if c.target != nil {
		c.target.Stop()
	}
	entryHandler := loki.NewEntryHandler(c.handler, func() {})

	t, err := cft.NewTarget(c.metrics, c.opts.Logger, entryHandler, c.posFile, newArgs.Config.Convert())
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "failed to create cloudflare target with provided config", "err", err)
		return err
	}
	c.target = t

	return nil
}

// DebugInfo returns information about the status of targets.
func (c *Component) DebugInfo() interface{} {
	c.mut.RLock()
	defer c.mut.RUnlock()

	lbls := make(map[string]string, len(c.target.Labels()))
	for k, v := range c.target.Labels() {
		lbls[string(k)] = string(v)
	}
	return targetDebugInfo{
		Ready:   c.target.Ready(),
		Labels:  lbls,
		Details: c.target.Details(),
	}
}

type targetDebugInfo struct {
	Ready   bool              `river:"ready,attr"`
	Labels  map[string]string `river:"labels,attr"`
	Details map[string]string `river:"target_info,attr"`
}
