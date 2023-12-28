package probes

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/operator"
	"github.com/grafana/agent/component/prometheus/operator/common"
)

func init() {
	component.Register(component.Registration{
		Name: "prometheus.operator.probes",
		Args: operator.Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return common.New(opts, args, common.KindProbe)
		},
	})
}
