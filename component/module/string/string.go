package string

import (
	"context"
	"strings"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/module/export"
)

func init() {
	component.RegisterDelegate(component.Registration{
		Name:    "module.string",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

var _ component.Component = (*Component)(nil)
var _ component.DelegateComponent = (*Component)(nil)

// Component
type Component struct {
	mut    sync.RWMutex
	args   Arguments
	values map[string]interface{}
	opts   component.Options
}

func (c *Component) ID() string {
	return c.opts.ID
}
func (c *Component) IDs() []string {
	return strings.Split(c.opts.ID, ".")
}

// Run starts the component.
func (c *Component) Run(ctx context.Context) error {
	return nil
}

// Update updates the fields of the component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()
	if c.args.Content == newArgs.Content {
		return nil
	}
	c.args = newArgs
	c.opts.Delegate.LoadSubgraph(c, []byte(c.args.Content))

	return nil
}

// Inform is used by children export components to inform the parent to run.
func (c *Component) Inform(e component.Exports) {
	ex := e.(export.Exports)
	c.mut.Lock()
	defer c.mut.Unlock()
	c.values[ex.Name] = ex.Value
	c.opts.OnStateChange(Exports{
		Exports: c.values,
	})
}

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

// New creates a new  component.
func New(o component.Options, args Arguments) (component.Component, error) {
	return &Component{
		args:   args,
		values: make(map[string]interface{}),
		opts:   o,
	}, nil
}
