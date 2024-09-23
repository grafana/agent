package file

import (
	"time"

	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/prometheus/common/model"
	prom_discovery "github.com/prometheus/prometheus/discovery/file"
)

func init() {
	component.Register(component.Registration{
		Name:      "discovery.file",
		Stability: featuregate.StabilityStable,
		Args:      Arguments{},
		Exports:   discovery.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return discovery.NewFromConvertibleConfig(opts, args.(Arguments))
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

func (a Arguments) Convert() discovery.DiscovererConfig {
	return &prom_discovery.SDConfig{
		Files:           a.Files,
		RefreshInterval: model.Duration(a.RefreshInterval),
	}
}
