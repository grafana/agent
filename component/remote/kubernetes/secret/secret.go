package secret

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/remote/kubernetes"
)

func init() {
	component.Register(component.Registration{
		Name:    "remote.kubernetes.secret",
		Args:    kubernetes.Arguments{},
		Exports: kubernetes.Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return kubernetes.New(opts, args.(kubernetes.Arguments), kubernetes.TypeSecret)
		},
	})
}
