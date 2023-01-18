package heroku

import (
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	"github.com/prometheus/common/model"
	sv "github.com/weaveworks/common/server"
)

// ListenerConfig defines a heroku listener.
type ListenerConfig struct {
	ListenAddress        string            `river:"address,attr"`
	ListenPort           int               `river:"port,attr"`
	Labels               map[string]string `river:"labels,attr,optional"`
	UseIncomingTimestamp bool              `river:"use_incoming_timestamp,attr,optional"`
}

// DefaultListenerConfig provides the default arguments for a heroku listener.
var DefaultListenerConfig = ListenerConfig{
	ListenAddress:        "localhost",
	ListenPort:           8080,
	UseIncomingTimestamp: true,
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
// func (sc ListenerConfig) Convert() *scrapeconfig.HerokuDrainTargetConfig {
// 	lbls := make(model.LabelSet, len(sc.Labels))
// 	for k, v := range sc.Labels {
// 		lbls[model.LabelName(k)] = model.LabelValue(v)
// 	}

// 	return &scrapeconfig.HerokuDrainTargetConfig{
// 		Labels:               lbls,
// 		UseIncomingTimestamp: sc.UseIncomingTimestamp,
// 	}
// }

// Convert is used to bridge between the River and Promtail types.
func (sc ListenerConfig) Convert() *scrapeconfig.HerokuDrainTargetConfig {
	lbls := make(model.LabelSet, len(sc.Labels))
	for k, v := range sc.Labels {
		lbls[model.LabelName(k)] = model.LabelValue(v)
	}

	return &scrapeconfig.HerokuDrainTargetConfig{
		Server: sv.Config{
			HTTPListenAddress: sc.ListenAddress,
			HTTPListenPort:    sc.ListenPort,
		},
		Labels:               lbls,
		UseIncomingTimestamp: sc.UseIncomingTimestamp,
	}
}
