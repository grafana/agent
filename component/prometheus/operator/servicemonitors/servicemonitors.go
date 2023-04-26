package servicemonitors

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/operator"
	"github.com/grafana/agent/component/prometheus/operator/common"
)

func init() {
	component.Register(component.Registration{
		Name: "prometheus.operator.servicemonitors",
		Args: operator.Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return common.New(opts, args, common.KindServiceMonitor)
		},
	})
}
