package journal

import (
	"context"
	"sync"
	"time"

	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.journal",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

var _ component.Component = (*Component)(nil)

// Component
type Component struct {
	mut sync.RWMutex
}

// Run starts the component.
func (c *Component) Run(ctx context.Context) error {
	return nil
}

// Update updates the fields of the component.
func (c *Component) Update(args component.Arguments) error {

	c.mut.Lock()
	defer c.mut.Unlock()

	return nil
}

// Arguments are the arguments for the component.
type Arguments struct {
	FormatAsJson bool          `river:"format_as_json,attr,optional"`
	MaxAge       time.Duration `river:"max_age,attr,optional"`
	Path         string        `river:"path,attr,optional"`
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
	return c, nil
}
