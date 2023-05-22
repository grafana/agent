package remote

import (
	"context"
	"net/url"
	"sync"
	"time"

	"github.com/grafana/agent/component"
	"github.com/prometheus/client_golang/prometheus"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	promremote "github.com/prometheus/prometheus/storage/remote"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.remote",
		Singleton: false,
		Args:      Arguments{},
		Exports:   Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewComponent(opts, args.(Arguments))
		},
	})
}

type Component struct {
	mut  sync.Mutex
	args Arguments
	qm   *QueueManager
}

// TODO(rfratto): This should be exposed. How do we want to expose this?
var remoteFlushDeadline = 1 * time.Minute

func NewComponent(o component.Options, c Arguments) (*Component, error) {
	ew := newEWMARate(ewmaWeight, shardUpdateDuration)
	endUrl, err := url.Parse(c.Endpoint.URL)
	if err != nil {
		return nil, err
	}
	cfgURL := &config_util.URL{URL: endUrl}
	wr, err := promremote.NewWriteClient(o.ID, &promremote.ClientConfig{
		URL:              cfgURL,
		Timeout:          model.Duration(c.Endpoint.RemoteTimeout),
		HTTPClientConfig: *c.Endpoint.HTTPClientConfig.Convert(),
		SigV4Config:      nil,
		Headers:          c.Endpoint.Headers,
		RetryOnRateLimit: c.Endpoint.QueueOptions.toPrometheusType().RetryOnRateLimit,
	})
	if err != nil {
		return nil, err
	}
	met := newQueueManagerMetrics(o.Registerer, "", "")

	qm := NewQueueManager(
		met,
		o.Logger,
		ew,
		c.Endpoint.QueueOptions.toPrometheusType(),
		c.Endpoint.MetadataOptions.toPrometheusType(),
		wr,
		remoteFlushDeadline,
		&maxTimestamp{
			Gauge: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace:   "prometheus",
				Subsystem:   "remote_storage",
				Name:        "highest_timestamp_in_seconds",
				Help:        "Highest timestamp that has come into the remote storage via the Appender interface, in seconds since epoch.",
				ConstLabels: map[string]string{"component_id": o.ID},
			})},
		true,
		true,
		nil,
	)
	return &Component{
		args: c,
		qm:   qm,
	}, nil
}

func (c *Component) Run(ctx context.Context) error {
	return nil
}

func (c *Component) Update(args component.Arguments) error {
	c.args = args.(Arguments)
	return nil
}

const (
	// We track samples in/out and how long pushes take using an Exponentially
	// Weighted Moving Average.
	ewmaWeight          = 0.2
	shardUpdateDuration = 10 * time.Second

	// Allow 30% too many shards before scaling down.
	shardToleranceFraction = 0.3
)
