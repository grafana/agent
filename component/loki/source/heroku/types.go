package heroku

import (
	"github.com/grafana/agent/component/loki/source/heroku/internal/herokutarget"
	"github.com/grafana/agent/pkg/river"
	"github.com/prometheus/common/model"
	sv "github.com/weaveworks/common/server"
)

// ListenerConfig defines a heroku listener.
type ListenerConfig struct {
	ListenAddress string `river:"address,attr,optional"`
	ListenPort    int    `river:"port,attr"`
	// TODO - add the rest of the server config from Promtail
}

// DefaultListenerConfig provides the default arguments for a heroku listener.
var DefaultListenerConfig = ListenerConfig{
	ListenAddress: "0.0.0.0",
}

var _ river.Unmarshaler = (*ListenerConfig)(nil)

// UnmarshalRiver implements river.Unmarshaler.
func (sc *ListenerConfig) UnmarshalRiver(f func(interface{}) error) error {
	*sc = DefaultListenerConfig

	type herokucfg ListenerConfig
	err := f((*herokucfg)(sc))
	if err != nil {
		return err
	}

	return nil
}

// Convert is used to bridge between the River and Promtail types.
func (args *Arguments) Convert() *herokutarget.HerokuDrainTargetConfig {
	lbls := make(model.LabelSet, len(args.Labels))
	for k, v := range args.Labels {
		lbls[model.LabelName(k)] = model.LabelValue(v)
	}

	return &herokutarget.HerokuDrainTargetConfig{
		Server: sv.Config{
			HTTPListenAddress: args.HerokuListener.ListenAddress,
			HTTPListenPort:    args.HerokuListener.ListenPort,
		},
		Labels:               lbls,
		UseIncomingTimestamp: args.UseIncomingTimestamp,
	}
}
