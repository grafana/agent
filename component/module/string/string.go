package string

import (
	"context"
	"strings"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/module/argument"
	"github.com/grafana/agent/component/module/export"
)

func init() {
	component.RegisterDelegate(component.Registration{
		Name: "module.string",
		Args: Arguments{},
		Exports: Exports{Exports: func(_ string) interface{} {
			return nil
		}},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

var _ component.Component = (*Component)(nil)
var _ component.DelegateComponent = (*Component)(nil)

// Component
type Component struct {
	mut              sync.RWMutex
	args             Arguments
	values           map[string]interface{}
	opts             component.Options
	components       []component.Component
	argComponents    map[string]*argument.Component
	exportComponents map[string]*export.Component
}

func (c *Component) ID() string {
	return c.opts.ID
}
func (c *Component) IDs() []string {
	return strings.Split(c.opts.ID, ".")
}

// Run starts the component.
func (c *Component) Run(ctx context.Context) error {
	c.Update(c.args)
	<-ctx.Done()
	return nil
}

// Update updates the fields of the component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()
	c.args = newArgs
	for k, v := range newArgs.Arguments {
		c.values[k] = v
	}
	// IF content has changed we need to rollback things
	if c.args.Content != newArgs.Content || len(c.components) == 0 {
		// TODO insert teardown here
		components, err := c.opts.Delegate.LoadSubgraph(c, []byte(c.args.Content))
		if err != nil {
			return err
		}
		c.components = components
		c.argComponents = make(map[string]*argument.Component)
		c.exportComponents = make(map[string]*export.Component)
		for _, x := range c.components {
			switch x.(type) {
			case *argument.Component:
				av := x.(*argument.Component)
				c.argComponents[av.Name] = av
			case *export.Component:
				xv := x.(*export.Component)
				// Add the callback so that if the exports value changes if can notify the parent
				xv.UpdateInform(func(e export.Exports) {
					c.values[e.Name] = e.Value
					c.opts.OnStateChange(Exports{
						Exports: func(name string) interface{} {
							c.mut.Lock()
							defer c.mut.Unlock()
							v, found := c.values[name]
							if !found {
								return nil
							}
							return v
						},
					})
				})
				c.exportComponents[xv.Name] = xv
			}
		}
	}
	for k, v := range c.argComponents {
		modVal, found := c.values[k]
		if !found {
			continue
		}
		v.Update(argument.Arguments{
			Value: modVal,
		})
	}
	// Fill out the export map so it exists
	exportMap := make(map[string]interface{})
	for _, v := range c.exportComponents {
		exportMap[v.Name] = v.Value
	}
	c.opts.OnStateChange(Exports{Exports: func(name string) interface{} {
		c.mut.Lock()
		defer c.mut.Unlock()
		v, found := c.values[name]
		if !found {
			return nil
		}
		return v

	}})
	return nil
}

// Inform is used by children export components to inform the parent to run.
func (c *Component) Inform(e component.Exports) {
	ex := e.(export.Exports)
	c.mut.Lock()
	defer c.mut.Unlock()
	c.values[ex.Name] = ex.Value
	c.opts.OnStateChange(Exports{
		Exports: func(name string) interface{} {
			c.mut.Lock()
			defer c.mut.Unlock()
			v, found := c.values[name]
			if !found {
				return nil
			}
			return v
		},
	})
}

type Exports struct {
	Exports func(name string) interface{} `river:"exports,attr"`
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

// New creates a new  component.
func New(o component.Options, args Arguments) (component.Component, error) {
	c := &Component{
		args:   args,
		values: make(map[string]interface{}),
		opts:   o,
	}
	return c, nil
}
