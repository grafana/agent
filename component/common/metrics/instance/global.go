package instance

import (
	"time"

	"github.com/grafana/agent/component/common/prometheus/config"
	internal "github.com/grafana/agent/pkg/metrics/instance"
)

type GlobalConfig struct {
	prometheus  config.GlobalConfig         `river:",squash"`
	remoteWrite []*config.RemoteWriteConfig `river:"remote_write_config,block,optional"`

	extraMetrics      bool
	disableKeepAlives bool
	idleConnTimeout   time.Duration
}

func (c *GlobalConfig) ToInternal() (*internal.GlobalConfig, error) {
	prometheus, err := c.prometheus.ToInternal()
	if err != nil {
		return nil, err
	}

	return &internal.GlobalConfig{
		Prometheus:        *prometheus,
		RemoteWrite:       config.RemoteWriteConfigsToInternals(c.remoteWrite),
		ExtraMetrics:      c.extraMetrics,
		DisableKeepAlives: c.disableKeepAlives,
		IdleConnTimeout:   c.idleConnTimeout,
	}, nil
}
