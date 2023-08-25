package simple

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/prometheus/storage"

	types "github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/pkg/river"
)

// Defaults for config blocks.
var (
	DefaultQueueOptions = QueueOptions{
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

func defaultArgs() Arguments {
	return Arguments{
		TTL:   2 * time.Hour,
		Evict: 15 * time.Minute,
	}
}

type Arguments struct {
	TTL   time.Duration `river:"ttl,attr,optional"`
	Evict time.Duration `river:"evict_interval,attr,optional"`

	Endpoint *EndpointOptions `river:"endpoint,block,optional"`
}

type Exports struct {
	Receiver storage.Appendable `river:"receiver,attr"`
}

// UnmarshalRiver implements river.Unmarshaler.
func (rc *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*rc = defaultArgs()

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
	// Capacity          int           `river:"capacity,attr,optional"`
	// MaxShards         int           `river:"max_shards,attr,optional"`
	// MinShards         int           `river:"min_shards,attr,optional"`
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

type Bookmark struct {
	Key uint64
}
