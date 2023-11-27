package process

import (
	"context"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/prometheus/prometheus/discovery/refresh"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.process",
		Args:    Arguments{},
		Exports: discovery.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
}

var DefaultConfig = Arguments{}

func (args *Arguments) SetToDefault() {
	*args = DefaultConfig
}

func New(opts component.Options, args Arguments) (*discovery.Component, error) {
	return discovery.New(opts, args, func(args component.Arguments) (discovery.Discoverer, error) {
		d := new(processDiscoverer)
		r := refresh.NewDiscovery(opts.Logger, "", 15*time.Second, d.Discover)
		return r, nil
	})
}

type processDiscoverer struct {
}

func (p *processDiscoverer) Discover(ctx context.Context) ([]*targetgroup.Group, error) {
	
	return nil, nil
}
