package remote

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/tsdb/wlog"

	"github.com/prometheus/prometheus/config"

	types "github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/pkg/river"
	common "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

// Defaults for config blocks.
var (
	DefaultArguments = Arguments{}

	DefaultQueueOptions = QueueOptions{
		Capacity:          10000,
		MaxShards:         50,
		MinShards:         1,
		MaxSamplesPerSend: 2000,
		BatchSendDeadline: 5 * time.Second,
		MinBackoff:        30 * time.Millisecond,
		MaxBackoff:        5 * time.Second,
		RetryOnHTTP429:    false,
	}

	DefaultMetadataOptions = MetadataOptions{
		Send:              true,
		SendInterval:      1 * time.Minute,
		MaxSamplesPerSend: 2000,
	}

	_ river.Unmarshaler = (*QueueOptions)(nil)
)

// Arguments represents the input state of the prometheus.remote_write
// component.
type Arguments struct {
	Endpoint *EndpointOptions `river:"endpoint,block,optional"`
}

// UnmarshalRiver implements river.Unmarshaler.
func (rc *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*rc = DefaultArguments

	type config Arguments
	return f((*config)(rc))
}

// EndpointOptions describes an individual location for where metrics in the WAL
// should be delivered to using the remote_write protocol.
type EndpointOptions struct {
	Name                 string                  `river:"name,attr,optional"`
	URL                  string                  `river:"url,attr"`
	RemoteTimeout        time.Duration           `river:"remote_timeout,attr,optional"`
	Headers              map[string]string       `river:"headers,attr,optional"`
	SendExemplars        bool                    `river:"send_exemplars,attr,optional"`
	SendNativeHistograms bool                    `river:"send_native_histograms,attr,optional"`
	HTTPClientConfig     *types.HTTPClientConfig `river:",squash"`
	QueueOptions         *QueueOptions           `river:"queue_config,block,optional"`
	MetadataOptions      *MetadataOptions        `river:"metadata_config,block,optional"`
}

func GetDefaultEndpointOptions() EndpointOptions {
	var defaultEndpointOptions = EndpointOptions{
		RemoteTimeout:    30 * time.Second,
		SendExemplars:    true,
		HTTPClientConfig: types.CloneDefaultHTTPClientConfig(),
	}

	return defaultEndpointOptions
}

// UnmarshalRiver implements river.Unmarshaler.
func (r *EndpointOptions) UnmarshalRiver(f func(v interface{}) error) error {
	*r = GetDefaultEndpointOptions()

	type arguments EndpointOptions
	err := f((*arguments)(r))
	if err != nil {
		return err
	}

	// We must explicitly Validate because HTTPClientConfig is squashed and it won't run otherwise
	if r.HTTPClientConfig != nil {
		return r.HTTPClientConfig.Validate()
	}

	return nil
}

// QueueOptions handles the low level queue config options for a remote_write
type QueueOptions struct {
	Capacity          int           `river:"capacity,attr,optional"`
	MaxShards         int           `river:"max_shards,attr,optional"`
	MinShards         int           `river:"min_shards,attr,optional"`
	MaxSamplesPerSend int           `river:"max_samples_per_send,attr,optional"`
	BatchSendDeadline time.Duration `river:"batch_send_deadline,attr,optional"`
	MinBackoff        time.Duration `river:"min_backoff,attr,optional"`
	MaxBackoff        time.Duration `river:"max_backoff,attr,optional"`
	RetryOnHTTP429    bool          `river:"retry_on_http_429,attr,optional"`
}

// UnmarshalRiver allows injecting of default values
func (r *QueueOptions) UnmarshalRiver(f func(v interface{}) error) error {
	*r = DefaultQueueOptions

	type arguments QueueOptions
	return f((*arguments)(r))
}

func (r *QueueOptions) toPrometheusType() config.QueueConfig {
	if r == nil {
		return config.DefaultQueueConfig
	}

	return config.QueueConfig{
		Capacity:          r.Capacity,
		MaxShards:         r.MaxShards,
		MinShards:         r.MinShards,
		MaxSamplesPerSend: r.MaxSamplesPerSend,
		BatchSendDeadline: model.Duration(r.BatchSendDeadline),
		MinBackoff:        model.Duration(r.MinBackoff),
		MaxBackoff:        model.Duration(r.MaxBackoff),
		RetryOnRateLimit:  r.RetryOnHTTP429,
	}
}

// MetadataOptions configures how metadata gets sent over the remote_write
// protocol.
type MetadataOptions struct {
	Send              bool          `river:"send,attr,optional"`
	SendInterval      time.Duration `river:"send_interval,attr,optional"`
	MaxSamplesPerSend int           `river:"max_samples_per_send,attr,optional"`
}

// UnmarshalRiver allows injecting of default values
func (o *MetadataOptions) UnmarshalRiver(f func(v interface{}) error) error {
	*o = DefaultMetadataOptions

	type options MetadataOptions
	return f((*options)(o))
}

func (o *MetadataOptions) toPrometheusType() config.MetadataConfig {
	if o == nil {
		return config.DefaultMetadataConfig
	}

	return config.MetadataConfig{
		Send:              o.Send,
		SendInterval:      model.Duration(o.SendInterval),
		MaxSamplesPerSend: o.MaxSamplesPerSend,
	}
}

type Exports struct {
	Receiver wlog.WriteTo `river:"receiver,attr"`
}

func convertConfigs(cfg Arguments) (*config.Config, error) {

	rw := cfg.Endpoint
	parsedURL, err := url.Parse(rw.URL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse remote_write url %q: %w", rw.URL, err)
	}
	rwCfg := &config.RemoteWriteConfig{
		URL:                  &common.URL{URL: parsedURL},
		RemoteTimeout:        model.Duration(rw.RemoteTimeout),
		Headers:              rw.Headers,
		WriteRelabelConfigs:  nil, // WriteRelabelConfigs are currently not supported
		Name:                 rw.Name,
		SendExemplars:        rw.SendExemplars,
		SendNativeHistograms: rw.SendNativeHistograms,

		HTTPClientConfig: *rw.HTTPClientConfig.Convert(),
		QueueConfig:      rw.QueueOptions.toPrometheusType(),
		MetadataConfig:   rw.MetadataOptions.toPrometheusType(),
		// TODO(rfratto): SigV4Config
	}

	return &config.Config{
		GlobalConfig:       config.GlobalConfig{},
		RemoteWriteConfigs: []*config.RemoteWriteConfig{rwCfg},
	}, nil
}

type maxTimestamp struct {
	mtx   sync.Mutex
	value float64
	prometheus.Gauge
}

func (m *maxTimestamp) Set(value float64) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if value > m.value {
		m.value = value
		m.Gauge.Set(value)
	}
}

func (m *maxTimestamp) Get() float64 {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.value
}

func (m *maxTimestamp) Collect(c chan<- prometheus.Metric) {
	if m.Get() > 0 {
		m.Gauge.Collect(c)
	}
}
