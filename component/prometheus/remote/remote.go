package remote

import (
	"context"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/prometheus/prometheus/tsdb/wlog"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.remote",
		Singleton: false,
		Args:      Arguments{},
		Exports:   Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewComponent(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
}

type Exports struct {
	Receiver wlog.WriteTo `river:"receiver,attr"`
}

type Component struct {
	mut  sync.Mutex
	args Arguments
}

func NewComponent(o component.Options, c Arguments) (*Component, error) {
	return &Component{}, nil
}

func (c *Component) Run(ctx context.Context) error {
	return nil
}

func (c *Component) Update(args component.Arguments) error {
	c.args = args.(Arguments)
	return nil
}
