package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.test.metrics",
		Singleton: false,
		Args:      Arguments{},
		Exports:   Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewComponent(opts, args.(Arguments))
		},
	})
}

type Component struct {
	mut  sync.Mutex
	args Arguments
}

func NewComponent(o component.Options, c Arguments) (*Component, error) {
	return &Component{
		args: c,
	}, nil
}

func (c *Component) Run(ctx context.Context) error {
	return nil

}

func (c *Component) Update(args component.Arguments) error {
	c.args = args.(Arguments)

	return nil
}

type Arguments struct {
	NumberOfInstances int           `river:"number_of_instances,attr,optional"`
	NumberOfMetrics   int           `river:"number_of_metrics,attr,optional"`
	NumberOfSeries    int           `river:"number_of_series,attr,optional"`
	MetricsRefresh    time.Duration `river:"metrics_refresh,attr,optional"`
}

type Exports struct {
	Targets []map[string]string `river:"targets,attr,optional"`
}
