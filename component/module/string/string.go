package string

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/grafana/agent/pkg/flow"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/module/argument"
	"github.com/grafana/agent/component/module/export"
	"github.com/grafana/agent/component/prometheus/exporter"
)

func init() {
	component.Register(component.Registration{
		Name:    "module.string",
		Args:    Arguments{},
		Exports: Exports{Exports: make(map[string]interface{})},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

var _ component.Component = (*Component)(nil)

// Component is the module.string component
type Component struct {
	mut              sync.RWMutex
	args             Arguments
	values           map[string]interface{}
	opts             component.Options
	argComponents    map[string]*argument.Component
	exportComponents map[string]*export.Component
	loadedOnce       bool
	flow             *flow.Flow
}

// New creates a new  component.
func New(o component.Options, args Arguments) (component.Component, error) {
	c := &Component{
		args:             args,
		values:           make(map[string]interface{}),
		argComponents:    make(map[string]*argument.Component),
		exportComponents: make(map[string]*export.Component),
		opts:             o,
	}
	c.flow = flow.New(flow.Options{
		Logger:         c.opts.Logger,
		Tracer:         c.opts.Tracer,
		DataPath:       filepath.Join(c.opts.DataPath, c.opts.ID),
		Reg:            c.opts.RootRegister,
		HTTPListenAddr: c.opts.HTTPListenAddr,
		NamespaceID:    c.opts.ID,
		Callback:       c.newComponent,
		Notify:         c.opts.Notify,
	})
	err := c.Update(args)
	return c, err
}

// Run starts the component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return c.flow.Close()
}

// Update updates the fields of the component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()

	for k, v := range newArgs.Arguments {
		c.values[k] = v
	}

	applyChange := c.args.Content != newArgs.Content
	if !applyChange {
		// We need to have loaded at least once
		applyChange = !c.loadedOnce
	}
	c.args = newArgs

	// If content has changed we need to apply the new string via the delegate
	if applyChange {
		// TODO insert teardown here
		// TODO do something with the diags
		c.argComponents = make(map[string]*argument.Component)
		c.exportComponents = make(map[string]*export.Component)

		flowFile, err := flow.ReadFile(c.opts.ID, []byte(c.args.Content))
		if len(flowFile.ConfigBlocks) > 0 {
			return fmt.Errorf("cannot have config blocks in a module")
		}
		if err != nil {
			return err
		}
		err = c.flow.LoadFile(flowFile)
		if err != nil {
			return err
		}

		c.loadedOnce = true
	}
	for k, v := range c.argComponents {
		modVal, found := c.values[k]
		if !found {
			continue
		}
		err := v.Update(argument.Arguments{
			Value: modVal,
		})
		if err != nil {
			return err
		}
	}
	exportMap := make(map[string]interface{})
	for k, v := range c.values {
		exportMap[k] = v
	}
	c.opts.OnStateChange(Exports{
		Exports: exportMap,
	})
	return nil
}

// newComponent is called for each component when it is evaluated by the loader.
// If this returns an error it will bubble up to the c.flow.LoadFile in this files Update function.
// This is why no locks are needed because its only called implicitly by something locking.
func (c *Component) newComponent(cmp component.Component) error {
	switch cc := cmp.(type) {
	case *argument.Component:
		cc.UpdateValue(c.argComponents[cc.Name])
		c.argComponents[cc.Name] = cc
	case *export.Component:
		// Add the callback so that if the exports value changes if can notify the parent
		name, val := cc.UpdateInform(func(e export.Exports) {
			c.values[e.Name] = e.Value
			exportMap := make(map[string]interface{})
			for k, v := range c.values {
				exportMap[k] = v
			}
			c.opts.OnStateChange(Exports{
				Exports: exportMap,
			})
		})
		// Go ahead and fill in the value
		c.values[name] = val
		c.exportComponents[cc.Name] = cc
	case *exporter.Component:
		if cc.GetId() == "prometheus.exporter.unix" {
			return fmt.Errorf("prometheus.exporter.unix is a singleton and cannot be a component in a module")
		}
	}
	return nil
}

// Exports exports the exports! We use a func since a map would not have the keys filled in on initial
// load. So when accessing the map on load river loader throws an error of key not found. Alternative
// would be to the ability to mark a map as a lazy map and not throw if key not found.
type Exports struct {
	Exports map[string]interface{} `river:"exports,attr"`
}

// Arguments are the arguments for the component.
type Arguments struct {
	Content   string                 `river:"content,attr"`
	Arguments map[string]interface{} `river:"arguments,attr,optional"`
}

func defaultArgs() Arguments {
	return Arguments{}
}

// UnmarshalRiver implements river.Unmarshaler.
func (r *Arguments) UnmarshalRiver(f func(v interface{}) error) error {
	*r = defaultArgs()

	type arguments Arguments
	if err := f((*arguments)(r)); err != nil {
		return err
	}

	return nil
}
