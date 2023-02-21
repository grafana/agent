package argument

import (
	"context"
	"strings"
	"sync"

	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name:    "module.argument",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

var _ component.Component = (*Component)(nil)

// Component is the argument component
type Component struct {
	mut  sync.RWMutex
	args Arguments
	opts component.Options
	Name string
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
	c.opts.OnStateChange(Exports{Value: c.args.Value})

	return nil
}

type Exports struct {
	Value interface{} `river:"value,attr,optional"`
}

// Arguments are the arguments for the component.
type Arguments struct {
	Value interface{} `river:"value,attr,optional"`
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
	splitName := strings.Split(o.ID, ".")
	return &Component{
		args: args,
		opts: o,
		Name: splitName[len(splitName)-1],
	}, nil
}
