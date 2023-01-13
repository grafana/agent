package heroku

import (
	"fmt"
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	"github.com/prometheus/common/model"
)

// ListenerConfig defines a heroku listener.
type ListenerConfig struct {
	ListenAddress        string            `river:"address,attr"`
	ListenProtocol       string            `river:"protocol,attr,optional"`
	IdleTimeout          time.Duration     `river:"idle_timeout,attr,optional"`
	LabelStructuredData  bool              `river:"label_structured_data,attr,optional"`
	Labels               map[string]string `river:"labels,attr,optional"`
	UseIncomingTimestamp bool              `river:"use_incoming_timestamp,attr,optional"`
	UseRFC5424Message    bool              `river:"use_rfc5424_message,attr,optional"`
	MaxMessageLength     int               `river:"max_message_length,attr,optional"`
	TLSConfig            config.TLSConfig  `river:"tls_config,block,optional"`
}

// DefaultListenerConfig provides the default arguments for a heroku listener.
/* TODO
var DefaultListenerConfig = ListenerConfig{
	ListenProtocol:   st.DefaultProtocol,
	IdleTimeout:      st.DefaultIdleTimeout,
	MaxMessageLength: st.DefaultMaxMessageLength,
}
*/

var _ river.Unmarshaler = (*ListenerConfig)(nil)

// UnmarshalRiver implements river.Unmarshaler.
func (sc *ListenerConfig) UnmarshalRiver(f func(interface{}) error) error {
	// TODO *sc = DefaultListenerConfig

	type herokucfg ListenerConfig
	err := f((*herokucfg)(sc))
	if err != nil {
		return err
	}

	if sc.ListenProtocol != "tcp" && sc.ListenProtocol != "udp" {
		return fmt.Errorf("heroku listener protocol should be either 'tcp' or 'udp', got %s", sc.ListenProtocol)
	}

	return nil
}

// Convert is used to bridge between the River and Promtail types.
func (sc ListenerConfig) Convert() *scrapeconfig.HerokuDrainTargetConfig {
	lbls := make(model.LabelSet, len(sc.Labels))
	for k, v := range sc.Labels {
		lbls[model.LabelName(k)] = model.LabelValue(v)
	}

	return &scrapeconfig.HerokuDrainTargetConfig{
		Labels:               lbls,
		UseIncomingTimestamp: sc.UseIncomingTimestamp,
	}
}
