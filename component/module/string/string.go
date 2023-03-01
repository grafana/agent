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
	component.Register(component.Registration{
		Name: "module.string",
		Args: Arguments{},
		Exports: Exports{Exports: func(name string) interface{} {
			// We have to instantiate the function with something so it's not nil and errors.
			// On load all values accessing the item will be nil and once exports start to get
			// filled in module.string will call OnStateChange which triggers a re-evaluation.
			return nil
		}},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

var _ component.Component = (*Component)(nil)
var _ component.SubgraphOwner = (*Component)(nil)

// Component is the module.string component
type Component struct {
	mut              sync.RWMutex
	args             Arguments
	values           map[string]interface{}
	opts             component.Options
	components       []component.Component
	argComponents    map[string]*argument.Component
	exportComponents map[string]*export.Component
	loadedOnce       bool
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
	return c, nil
}

// ID satisfies the DelegateComponent interface
func (c *Component) ID() string {
	return c.opts.ID
}

// IDs satisfies the DelegateComponent interface
func (c *Component) IDs() []string {
	return strings.Split(c.opts.ID, ".")
}

// Run starts the component.
func (c *Component) Run(ctx context.Context) error {
	err := c.Update(c.args)
	if err != nil {
		return err
	}
	<-ctx.Done()
	err = c.opts.Subgraph.UnloadSubgraph(c)
	return err
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
		components, _, err := c.opts.Subgraph.LoadSubgraph(c, []byte(c.args.Content))
		if err != nil {
			return err
		}
		c.components = components
		c.argComponents = make(map[string]*argument.Component)
		c.exportComponents = make(map[string]*export.Component)
		// We need to scan all the components that were loaded and see if they are an argument (input) or
		// an export (output). If they are an argument then we stuff the value. If it is an export then
		// we need to register a callback so that we can call onstatechange when they change.
		for _, x := range c.components {
			switch cc := x.(type) {
			case *argument.Component:
				c.argComponents[cc.Name] = cc
			case *export.Component:
				// Add the callback so that if the exports value changes if can notify the parent
				cc.UpdateInform(func(e export.Exports) {
					c.values[e.Name] = e.Value
					c.opts.OnStateChange(Exports{
						Exports: c.GetVal,
					})
				})
				c.exportComponents[cc.Name] = cc
			}
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
	// Fill out the export map so it exists
	exportMap := make(map[string]interface{})
	for _, v := range c.exportComponents {
		exportMap[v.Name] = v.Value
	}
	c.opts.OnStateChange(Exports{
		Exports: c.GetVal,
	})
	return c.opts.Subgraph.StartSubgraph(c)
}

// Inform is used by children export components to inform the parent to run.
func (c *Component) Inform(e component.Exports) {
	ex := e.(export.Exports)
	c.mut.Lock()
	defer c.mut.Unlock()
	c.values[ex.Name] = ex.Value
	c.opts.OnStateChange(Exports{
		Exports: c.GetVal,
	})
}

func (c *Component) GetVal(name string) interface{} {
	c.mut.Lock()
	defer c.mut.Unlock()
	v, found := c.values[name]
	if !found {
		return nil
	}
	return v
}

// Exports exports the exports! We use a func since a map would not have the keys filled in on initial
// load. So when accessing the map on load river loader throws an error of key not found. Alternative
// would be to the ability to mark a map as a lazy map and not throw if key not found.
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
