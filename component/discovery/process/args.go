package process

import (
	"time"

	"github.com/grafana/agent/component/discovery"
)

type Arguments struct {
	Join            []discovery.Target `river:"join,attr,optional"`
	RefreshInterval time.Duration      `river:"refresh_interval,attr,optional"`
}

var DefaultConfig = Arguments{
	Join:            nil,
	RefreshInterval: 14 * time.Second,
}

func (args *Arguments) SetToDefault() {
	*args = DefaultConfig
}
