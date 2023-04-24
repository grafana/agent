package gcplogtarget

import (
	"fmt"
	"time"

	lhttp "github.com/grafana/agent/component/common/loki/http"
)

// Target is a common interface implemented by both GCPLog targets.
type Target interface {
	Details() map[string]string
	Stop() error
}

// PullConfig configures a GCPLog target with the 'pull' strategy.
type PullConfig struct {
	ProjectID            string            `river:"project_id,attr"`
	Subscription         string            `river:"subscription,attr"`
	Labels               map[string]string `river:"labels,attr,optional"`
	UseIncomingTimestamp bool              `river:"use_incoming_timestamp,attr,optional"`
}

// PushConfig configures a GCPLog target with the 'push' strategy.
type PushConfig struct {
	Server               *lhttp.ServerConfig `river:",squash"`
	PushTimeout          time.Duration       `river:"push_timeout,attr,optional"`
	Labels               map[string]string   `river:"labels,attr,optional"`
	UseIncomingTimestamp bool                `river:"use_incoming_timestamp,attr,optional"`
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
