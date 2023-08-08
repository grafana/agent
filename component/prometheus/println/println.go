package println

import (
	"context"
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.println",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct{}

type Exports struct {
	Receiver storage.Appendable `river:"receiver,attr"`
}

type Component struct {
	opts     component.Options
	args     Arguments
	receiver *prometheus.Interceptor
}

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{opts: o}

	c.receiver = prometheus.NewInterceptor(
		nil,
		prometheus.WithAppendHook(func(ref storage.SeriesRef, l labels.Labels, t int64, v float64, next storage.Appender) (storage.SeriesRef, error) {
			for _, value := range l {
				fmt.Printf("%s %s %d %e\n", value.Name, value.Value, t, v)
			}
			fmt.Println("--------")
			return ref, nil
		}),
	)

	o.OnStateChange(Exports{Receiver: c.receiver})

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	return nil
}
