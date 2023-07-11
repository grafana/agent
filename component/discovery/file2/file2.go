package file2

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/prometheus/common/model"
	prom_discovery "github.com/prometheus/prometheus/discovery/file"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.file2",
		Args:    Arguments{},
		Exports: discovery.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	Files           []string      `river:"files,attr"`
	RefreshInterval time.Duration `river:"refresh_interval,attr,optional"`
}

var DefaultArguments = Arguments{
	RefreshInterval: 5 * time.Minute,
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

func (a *Arguments) Convert() *prom_discovery.SDConfig {
	return &prom_discovery.SDConfig{
		Files:           a.Files,
		RefreshInterval: model.Duration(a.RefreshInterval),
	}
}

func New(opts component.Options, args Arguments) (*discovery.Component, error) {
	return discovery.New(opts, args, func(args component.Arguments) (discovery.Discoverer, error) {
		newArgs := args.(Arguments)
		return prom_discovery.NewDiscovery(newArgs.Convert(), opts.Logger), nil
	})
}
