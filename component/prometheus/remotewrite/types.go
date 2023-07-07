package remotewrite

import (
	"fmt"
	"net/url"
	"sort"
	"time"

	types "github.com/grafana/agent/component/common/config"
	flow_relabel "github.com/grafana/agent/component/common/relabel"

	common "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
)

// Defaults for config blocks.
var (
	DefaultArguments = Arguments{
		WALOptions: DefaultWALOptions,
	}

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

	DefaultWALOptions = WALOptions{
		TruncateFrequency: 2 * time.Hour,
		MinKeepaliveTime:  5 * time.Minute,
		MaxKeepaliveTime:  8 * time.Hour,
	}
)

// Arguments represents the input state of the prometheus.remote_write
// component.
type Arguments struct {
	ExternalLabels map[string]string  `river:"external_labels,attr,optional"`
	Endpoints      []*EndpointOptions `river:"endpoint,block,optional"`
	WALOptions     WALOptions         `river:"wal,block,optional"`
}

// SetToDefault implements river.Defaulter.
func (rc *Arguments) SetToDefault() {
	*rc = DefaultArguments
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
	WriteRelabelConfigs  []*flow_relabel.Config  `river:"write_relabel_config,block,optional"`
}

// SetToDefault implements river.Defaulter.
func (r *EndpointOptions) SetToDefault() {
	*r = EndpointOptions{
		RemoteTimeout:    30 * time.Second,
		SendExemplars:    true,
		HTTPClientConfig: types.CloneDefaultHTTPClientConfig(),
	}
}

// Validate implements river.Validator.
func (r *EndpointOptions) Validate() error {
	// We must explicitly Validate because HTTPClientConfig is squashed and it won't run otherwise
	if r.HTTPClientConfig != nil {
		if err := r.HTTPClientConfig.Validate(); err != nil {
			return err
		}
	}

	if r.WriteRelabelConfigs != nil {
		for _, relabelConfig := range r.WriteRelabelConfigs {
			if err := relabelConfig.Validate(); err != nil {
				return err
			}
		}
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

// SetToDefault implements river.Defaulter.
func (r *QueueOptions) SetToDefault() {
	*r = DefaultQueueOptions
}

func (r *QueueOptions) toPrometheusType() config.QueueConfig {
	if r == nil {
		var res QueueOptions
		res.SetToDefault()
		return res.toPrometheusType()
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

// SetToDefault implements river.Defaulter.
func (o *MetadataOptions) SetToDefault() {
	*o = DefaultMetadataOptions
}

func (o *MetadataOptions) toPrometheusType() config.MetadataConfig {
	if o == nil {
		var res MetadataOptions
		res.SetToDefault()
		return res.toPrometheusType()
	}

	return config.MetadataConfig{
		Send:              o.Send,
		SendInterval:      model.Duration(o.SendInterval),
		MaxSamplesPerSend: o.MaxSamplesPerSend,
	}
}

// WALOptions configures behavior within the WAL.
type WALOptions struct {
	TruncateFrequency time.Duration `river:"truncate_frequency,attr,optional"`
	MinKeepaliveTime  time.Duration `river:"min_keepalive_time,attr,optional"`
	MaxKeepaliveTime  time.Duration `river:"max_keepalive_time,attr,optional"`
}

// SetToDefault implements river.Defaulter.
func (o *WALOptions) SetToDefault() {
	*o = DefaultWALOptions
}

// Validate implements river.Validator.
func (o *WALOptions) Validate() error {
	switch {
	case o.TruncateFrequency == 0:
		return fmt.Errorf("truncate_frequency must not be 0")
	case o.MaxKeepaliveTime <= o.MinKeepaliveTime:
		return fmt.Errorf("min_keepalive_time must be smaller than max_keepalive_time")
	}

	return nil
}

// Exports are the set of fields exposed by the prometheus.remote_write
// component.
type Exports struct {
	Receiver storage.Appendable `river:"receiver,attr"`
}

func convertConfigs(cfg Arguments) (*config.Config, error) {
	var rwConfigs []*config.RemoteWriteConfig
	for _, rw := range cfg.Endpoints {
		parsedURL, err := url.Parse(rw.URL)
		if err != nil {
			return nil, fmt.Errorf("cannot parse remote_write url %q: %w", rw.URL, err)
		}

		rwConfigs = append(rwConfigs, &config.RemoteWriteConfig{
			URL:                  &common.URL{URL: parsedURL},
			RemoteTimeout:        model.Duration(rw.RemoteTimeout),
			Headers:              rw.Headers,
			Name:                 rw.Name,
			SendExemplars:        rw.SendExemplars,
			SendNativeHistograms: rw.SendNativeHistograms,

			WriteRelabelConfigs: flow_relabel.ComponentToPromRelabelConfigs(rw.WriteRelabelConfigs),
			HTTPClientConfig:    *rw.HTTPClientConfig.Convert(),
			QueueConfig:         rw.QueueOptions.toPrometheusType(),
			MetadataConfig:      rw.MetadataOptions.toPrometheusType(),
			// TODO(rfratto): SigV4Config
		})
	}

	return &config.Config{
		GlobalConfig: config.GlobalConfig{
			ExternalLabels: toLabels(cfg.ExternalLabels),
		},
		RemoteWriteConfigs: rwConfigs,
	}, nil
}

func toLabels(in map[string]string) labels.Labels {
	res := make(labels.Labels, 0, len(in))
	for k, v := range in {
		res = append(res, labels.Label{Name: k, Value: v})
	}
	sort.Sort(res)
	return res
}
