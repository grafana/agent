package file

import (
	"context"
	"sync"
	"time"

	"github.com/grafana/agent/component/discovery"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.file",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the discovery.file
// component.
type Arguments struct {
	PathTargets []discovery.Target `river:"path_targets,attr"`
	SyncPeriod  time.Duration      `river:"sync_period,attr,optional"`
}

// Exports exposes targets.
type Exports struct {
	Targets []discovery.Target `river:"targets,attr"`
}

var _ component.Component = (*Component)(nil)

// Component implements the discovery.file component.
type Component struct {
	opts component.Options

	mut     sync.RWMutex
	args    Arguments
	watches []watch
}

// New creates a new discovery.file component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:    o,
		mut:     sync.RWMutex{},
		args:    args,
		watches: make([]watch, 0),
	}

	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

func getDefault() Arguments {
	return Arguments{SyncPeriod: 10 * time.Second}
}

// UnmarshalRiver implements river.Unmarshaler.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = getDefault()
	type arguments Arguments
	return f((*arguments)(a))
}

// Update satisfies the component interface.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.args = args.(Arguments)
	c.watches = c.watches[:0]
	for _, v := range c.args.PathTargets {
		c.watches = append(c.watches, watch{
			targets: v,
		})
	}
	return nil
}

// Run satisfies the component interface.
func (c *Component) Run(ctx context.Context) error {
	watchDog := time.NewTicker(c.args.SyncPeriod)
	timerDuration := c.args.SyncPeriod
	update := func() {
		c.mut.Lock()
		defer c.mut.Unlock()

		paths := c.getWatchedFiles()
		// The component node checks to see if exports have actually changed.
		c.opts.OnStateChange(Exports{Targets: paths})

		// Check to see if our ticker timer needs to be reset.
		if timerDuration != c.args.SyncPeriod {
			watchDog.Reset(c.args.SyncPeriod)
			timerDuration = c.args.SyncPeriod
		}
	}
	// Trigger initial check
	update()
	defer watchDog.Stop()
	for {
		select {
		case <-watchDog.C:
			// This triggers a check for any new paths, along with pushing new targets.
			update()
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *Component) getWatchedFiles() []discovery.Target {
	paths := make([]discovery.Target, 0)
	// See if there is anything new we need to check.
	for _, w := range c.watches {
		newPaths, err := w.getPaths()
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "error getting paths", "path", w.getPath(), "excluded", w.getExcludePath(), "err", err)
		}
		paths = append(paths, newPaths...)
	}
	return paths
}
