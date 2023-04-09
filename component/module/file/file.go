package file

import (
	"context"
	"net/http"
	"sync"

	"go.uber.org/atomic"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/local/file"
	"github.com/grafana/agent/component/module"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/river"
)

func init() {
	component.Register(component.Registration{
		Name:    "module.file",
		Args:    Arguments{},
		Exports: module.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the module.file component.
type Arguments struct {
	LocalFileArguments file.Arguments `river:",squash"`

	// Arguments to pass into the module.
	Arguments map[string]any `river:"arguments,attr,optional"`
}

var _ river.Unmarshaler = (*Arguments)(nil)

// UnmarshalRiver implements river.Unmarshaler.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	a.LocalFileArguments = file.DefaultArguments

	type arguments Arguments
	err := f((*arguments)(a))
	if err != nil {
		return err
	}

	return nil
}

// Component implements the module.file component.
type Component struct {
	opts component.Options
	mod  *module.ModuleComponent

	mut     sync.RWMutex
	args    Arguments
	content rivertypes.OptionalSecret

	managedLocalFile *file.Component
	inUpdate         atomic.Bool
	isCreated        atomic.Bool
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
	_ component.HTTPComponent   = (*Component)(nil)
)

// New creates a new module.file component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts: o,
		mod:  module.NewModuleComponent(o),
		args: args,
	}
	defer c.isCreated.Store(true)

	var err error
	c.managedLocalFile, err = c.newManagedLocalComponent(o)
	if err != nil {
		return nil, err
	}
	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// NewManagedLocalComponent creates the new local.file managed component.
func (c *Component) newManagedLocalComponent(o component.Options) (*file.Component, error) {
	localFileOpts := o
	localFileOpts.OnStateChange = func(e component.Exports) {
		c.setContent(e.(file.Exports).Content)

		if !c.inUpdate.Load() && c.isCreated.Load() {
			// Any errors found here are reported via component health
			_ = c.mod.LoadFlowContent(c.getArgs().Arguments, c.getContent().Value)
		}
	}

	return file.New(localFileOpts, c.getArgs().LocalFileArguments)
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan error, 1)
	go func() {
		err := c.managedLocalFile.Run(ctx)
		if err != nil {
			ch <- err
		}
	}()

	go c.mod.RunFlowController(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-ch:
			return err
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.inUpdate.Store(true)
	defer c.inUpdate.Store(false)

	newArgs := args.(Arguments)
	c.setArgs(newArgs)

	err := c.managedLocalFile.Update(newArgs.LocalFileArguments)
	if err != nil {
		return err
	}

	// Force a content load here and bubble up any error. This will catch problems
	// on initial load.
	return c.mod.LoadFlowContent(newArgs.Arguments, c.getContent().Value)
}

// Handler implements component.HTTPComponent.
func (c *Component) Handler() http.Handler {
	return c.mod.Handler()
}

// CurrentHealth implements component.HealthComponent.
func (c *Component) CurrentHealth() component.Health {
	return c.mod.CurrentHealth()
}

// getArgs is a goroutine safe way to get args
func (c *Component) getArgs() Arguments {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.args
}

// setArgs is a goroutine safe way to set args
func (c *Component) setArgs(args Arguments) {
	c.mut.Lock()
	c.args = args
	c.mut.Unlock()
}

// getContent is a goroutine safe way to get content
func (c *Component) getContent() rivertypes.OptionalSecret {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.content
}

// setContent is a goroutine safe way to set content
func (c *Component) setContent(content rivertypes.OptionalSecret) {
	c.mut.Lock()
	c.content = content
	c.mut.Unlock()
}
