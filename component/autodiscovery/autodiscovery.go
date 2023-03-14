package autodiscovery

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/autodiscovery/runner"
)

func init() {
	component.Register(component.Registration{
		Name:    "autodiscovery",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

func New(o component.Options, args component.Arguments) (*Component, error) {
	c := &Component{
		opts: o,
	}

	return c, c.Update(args)
}

type Exports struct {
	Config string `river:"config,attr"`
}

type Arguments struct {
	RefreshPeriod time.Duration `river:"refresh_period,attr,optional"`
	Enabled       []string      `river:"enabled,attr,optional"`
	Disabled      []string      `river:"disabled,attr,optional"`
	//TODO: When parsing "enabled" and "disabled", turn them to lowercase?
	//TODO: A bool attribute to toggle automatic integrations on/off?
	//TODO: What happens if a mechanism is listed in both "enabled" and "disabled"?
}

func getDefault() Arguments {
	return Arguments{RefreshPeriod: time.Hour * 24}
}

// UnmarshalRiver implements river.Unmarshaler.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = getDefault()

	type arguments Arguments
	return f((*arguments)(a))
}

var _ component.Component = (*Component)(nil)

type Component struct {
	opts component.Options

	mut           sync.RWMutex
	RefreshPeriod time.Duration
	runner        runner.Autodiscovery
}

// Run implements component.Compnoent.
func (c *Component) Run(ctx context.Context) error {
	ticker := time.NewTicker(c.RefreshPeriod)

	for {
		select {
		case <-ticker.C:
			c.mut.Lock()

			buf := new(bytes.Buffer)
			c.runner.Do(buf)
			c.opts.Logger.Log("msg", "autodiscovered river config", "contents", buf.String())

			c.mut.Unlock()
		case <-ctx.Done():
			return nil
		}
	}

}

// Update implements component.Compnoent.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	inputArgs := args.(Arguments)
	c.RefreshPeriod = inputArgs.RefreshPeriod
	c.runner.Enabled = runner.ConvertMechanismStringSliceToMap(inputArgs.Enabled)
	c.runner.Disabled = runner.ConvertMechanismStringSliceToMap(inputArgs.Disabled)

	return nil
}
