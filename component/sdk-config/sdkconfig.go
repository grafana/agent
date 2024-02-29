package sdkconfig

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/module"
	"github.com/grafana/agent/internal/featuregate"
)

func init() {
	component.Register(component.Registration{
		Name:      "sdkconfig",
		Stability: featuregate.StabilityExperimental,
		Args:      Arguments{},
		Exports:   module.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}
