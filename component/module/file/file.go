package file

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

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
	mod module.ModuleComponent

	mut     sync.RWMutex
	args    Arguments
	content rivertypes.OptionalSecret

	managedLocalFile *file.Component
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
	_ component.HTTPComponent   = (*Component)(nil)
)

// New creates a new module.file component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		mod:  module.NewModuleComponent(o),
		args: args,
	}

	var err error
	c.managedLocalFile, err = c.NewManagedLocalComponent(o)
	if err != nil {
		return nil, err
	}

	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	ch := make(chan error, 1)
	go func() {
		err := c.managedLocalFile.Run(ctx)
		if err != nil {
			ch <- err
		}
	}()

	c.mod.RunFlowController(ctx)

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
	newArgs := args.(Arguments)

	c.setArgs(newArgs)

	err := c.managedLocalFile.Update(newArgs.LocalFileArguments)
	if err != nil {
		c.setHealth(component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    fmt.Sprintf("failed to update the managed local.file component: %s", err),
			UpdateTime: time.Now(),
		})
		return err
	}

	// Force a content load here and bubble up any error. This will catch problems
	// on initial load.
	return c.mod.LoadFlowContent(newArgs.Arguments, c.getContent().Value)
}

// NewManagedLocalComponent creates the new local.file managed component.
func (c *Component) NewManagedLocalComponent(o component.Options) (*file.Component, error) {
	localFileOpts := o
	localFileOpts.OnStateChange = func(e component.Exports) {
		c.setContent(e.(file.Exports).Content)

		// Any errors found here are reported via component health
		_ = c.mod.LoadFlowContent(c.getArgs().Arguments, c.getContent().Value)
	}

	return file.New(localFileOpts, c.getArgs().LocalFileArguments)
}

// Handler implements component.HTTPComponent.
func (c *Component) Handler() http.Handler {
	return c.mod.Handler()
}

// CurrentHealth implements component.HealthComponent.
func (c *Component) CurrentHealth() component.Health {
	return c.mod.CurrentHealth()
}

// setHealth updates the component health.
func (c *Component) setHealth(h component.Health) {
	c.mod.SetHealth(h)
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
