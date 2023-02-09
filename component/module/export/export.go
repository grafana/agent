package export

import (
	"context"
	"strings"
	"sync"

	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name: "module.export",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

var _ component.Component = (*Component)(nil)

// Component
type Component struct {
	mut    sync.RWMutex
	args   Arguments
	o      component.Options
	name   string
	inform func(e Exports)
}

// Run starts the component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update updates the fields of the component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()
	c.args = newArgs
	// Inform the parent module of the change, Export components are ONLY meant to be used within the context
	// of being in a submodule.
	if c.inform != nil {
		c.inform(Exports{
			Name:  c.name,
			Value: c.args.Value,
		})
	}
	return nil
}

func (c *Component) UpdateInform(f func(e Exports)) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.inform = f
	// Trigger an inform to be proactive.
	c.inform(Exports{
		Name:  c.name,
		Value: c.args.Value,
	})
}

// Arguments are the arguments for the component.
type Arguments struct {
	Value interface{} `river:"value,attr,required"`
}

func defaultArgs() Arguments {
	return Arguments{}
}

type Exports struct {
	Name  string      `river:"name,attr"`
	Value interface{} `river:"value,attr"`
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
	ids := strings.Split(o.ID, ".")
	return &Component{
		o:    o,
		args: args,
		name: ids[len(ids)-1],
	}, nil
}
