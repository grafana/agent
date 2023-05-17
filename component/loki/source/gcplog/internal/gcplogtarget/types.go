package gcplogtarget

import (
	"fmt"
	"time"

	"github.com/weaveworks/common/server"
	"github.com/prometheus/common/model"
)

// Target is a common interface implemented by both GCPLog targets.
type Target interface {
	Details() map[string]string
	Stop() error
}

// PullConfig configures a GCPLog target with the 'pull' strategy.
// struct is a derived of 'GcplogTargetConfig', but specific for Pull configuration, from Promtail code (https://github.com/grafana/loki/blob/main/clients/pkg/promtail/scrapeconfig/scrapeconfig.go)
type PullConfig struct {
	ProjectID            string            `river:"project_id,attr"`
	Subscription         string            `river:"subscription,attr"`
	Labels               model.LabelSet    `river:"labels,attr,optional"`
	UseIncomingTimestamp bool              `river:"use_incoming_timestamp,attr,optional"`
	UseFullLine          bool              `river:"use_full_line,attr,optional"`
}

// PushConfig configures a GCPLog target with the 'push' strategy.
// struct is a derived of 'GcplogTargetConfig', but specific for Push configuration from Promtail code (https://github.com/grafana/loki/blob/main/clients/pkg/promtail/scrapeconfig/scrapeconfig.go)
type PushConfig struct {
	Server               server.Config      `river:",squash"`
	PushTimeout          time.Duration      `river:"push_timeout,attr,optional"`
	Labels               model.LabelSet     `river:"labels,attr,optional"`
	UseIncomingTimestamp bool               `river:"use_incoming_timestamp,attr,optional"`
	UseFullLine          bool               `river:"use_full_line,attr,optional"`
}

// UnmarshalRiver implements the unmarshaller
func (p *PushConfig) UnmarshalRiver(f func(v interface{}) error) error {
	type pushCfg PushConfig
	err := f((*pushCfg)(p))
	if err != nil {
		return err
	}
	if p.PushTimeout < 0 {
		return fmt.Errorf("push_timeout must be greater than zero")
	}
	return nil
}
