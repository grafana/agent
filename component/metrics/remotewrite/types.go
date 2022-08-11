package remotewrite

import (
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/prometheus/prometheus/model/labels"

	"github.com/prometheus/prometheus/config"

	types "github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/metrics"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/river"
	common "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

// Defaults for config blocks.
var (
	DefaultRemoteConfig = RemoteConfig{
		WALOptions: DefaultWALOptions,
	}

	DefaultQueueConfig = QueueConfig{
		Capacity:          2500,
		MaxShards:         200,
		MinShards:         1,
		MaxSamplesPerSend: 500,
		BatchSendDeadline: 5 * time.Second,
		MinBackoff:        30 * time.Millisecond,
		MaxBackoff:        5 * time.Second,
		RetryOn429:        false,
	}

	DefaultWALOptions = WALOptions{
		TruncateFrequency: 2 * time.Hour,
		MinKeepaliveTime:  5 * time.Minute,
		MaxKeepaliveTime:  8 * time.Hour,
	}

	_ river.Unmarshaler = (*QueueConfig)(nil)
)

// RemoteConfig represents the input state of the metrics_forwarder component.
type RemoteConfig struct {
	ExternalLabels map[string]string `river:"external_labels,attr,optional"`
	RemoteWrite    []*Config         `river:"remote_write,block,optional"`
	WALOptions     WALOptions        `river:"wal,block,optional"`
}

// UnmarshalRiver implements river.Unmarshaler.
func (rc *RemoteConfig) UnmarshalRiver(f func(interface{}) error) error {
	*rc = DefaultRemoteConfig

	type config RemoteConfig
	return f((*config)(rc))
}

// Config is the metrics_fowarder's configuration for where to send
// metrics stored in the WAL.
type Config struct {
	Name             string                  `river:"name,attr,optional"`
	URL              string                  `river:"url,attr"`
	SendExemplars    bool                    `river:"send_exemplars,attr,optional"`
	BasicAuth        *BasicAuthConfig        `river:"basic_auth,block,optional"`
	QueueConfig      *QueueConfig            `river:"queue_config,block,optional"`
	HTTPClientConfig *types.HTTPClientConfig `river:"client_config,block,optional"`
}

// QueueConfig handles the low level queue config options for a remote_write
type QueueConfig struct {
	Capacity          int           `river:"capacity,attr,optional"`
	MaxShards         int           `river:"max_shards,attr,optional"`
	MinShards         int           `river:"min_shards,attr,optional"`
	MaxSamplesPerSend int           `river:"max_samples_per_send,attr,optional"`
	BatchSendDeadline time.Duration `river:"batch_send_deadline,attr,optional"`
	MinBackoff        time.Duration `river:"min_backoff,attr,optional"`
	MaxBackoff        time.Duration `river:"max_backoff,attr,optional"`
	RetryOn429        bool          `river:"retry_on_http_429,attr,optional"`
}

// WALOptions configures behavior within the WAL.
type WALOptions struct {
	TruncateFrequency time.Duration `river:"truncate_frequency,attr,optional"`
	MinKeepaliveTime  time.Duration `river:"min_keepalive_time,attr,optional"`
	MaxKeepaliveTime  time.Duration `river:"max_keepalive_time,attr,optional"`
}

// UnmarshalRiver implements river.Unmarshaler.
func (o *WALOptions) UnmarshalRiver(f func(interface{}) error) error {
	*o = DefaultWALOptions

	type config WALOptions
	if err := f((*config)(o)); err != nil {
		return err
	}

	switch {
	case o.TruncateFrequency == 0:
		return fmt.Errorf("truncate_frequency must not be 0")
	case o.MaxKeepaliveTime <= o.MinKeepaliveTime:
		return fmt.Errorf("min_keepalive_time must be smaller than max_keepalive_time")
	}

	return nil
}

// Export is used to assign this to receive metrics
type Export struct {
	Receiver *metrics.Receiver `river:"receiver,attr"`
}

// BasicAuthConfig is the metrics_forwarder's configuration for authenticating
// against the remote system when sending metrics.
type BasicAuthConfig struct {
	Username     string            `river:"username,attr"`
	Password     rivertypes.Secret `river:"password,attr,optional"`
	PasswordFile string            `river:"password_file,attr,optional"`
}

// UnmarshalRiver allows injecting of default values
func (r *Config) UnmarshalRiver(f func(v interface{}) error) error {
	*r = Config{
		SendExemplars: true,
	}

	type arguments Config
	return f((*arguments)(r))
}

// UnmarshalRiver allows injecting of default values
func (r *QueueConfig) UnmarshalRiver(f func(v interface{}) error) error {
	*r = DefaultQueueConfig

	type arguments QueueConfig
	return f((*arguments)(r))
}

func convertConfigs(cfg RemoteConfig) (*config.Config, error) {
	var rwConfigs []*config.RemoteWriteConfig
	for _, rw := range cfg.RemoteWrite {
		parsedURL, err := url.Parse(rw.URL)
		if err != nil {
			return nil, fmt.Errorf("cannot parse remote_write url %q: %w", rw.URL, err)
		}

		rwc := &config.RemoteWriteConfig{
			Name:          rw.Name,
			URL:           &common.URL{URL: parsedURL},
			RemoteTimeout: model.Duration(30 * time.Second),
			QueueConfig:   config.DefaultQueueConfig,
			MetadataConfig: config.MetadataConfig{
				Send: false,
			},
			HTTPClientConfig: common.DefaultHTTPClientConfig,
		}

		if rw.BasicAuth != nil {
			rwc.HTTPClientConfig.BasicAuth = &common.BasicAuth{
				Username:     rw.BasicAuth.Username,
				Password:     common.Secret(rw.BasicAuth.Password),
				PasswordFile: rw.BasicAuth.PasswordFile,
			}
		}

		if rw.QueueConfig != nil {
			rwc.QueueConfig = config.QueueConfig{
				Capacity:          rw.QueueConfig.Capacity,
				MaxShards:         rw.QueueConfig.MaxShards,
				MinShards:         rw.QueueConfig.MinShards,
				MaxSamplesPerSend: rw.QueueConfig.MaxSamplesPerSend,
				BatchSendDeadline: model.Duration(rw.QueueConfig.BatchSendDeadline),
				MinBackoff:        model.Duration(rw.QueueConfig.MinBackoff),
				MaxBackoff:        model.Duration(rw.QueueConfig.MaxBackoff),
				RetryOnRateLimit:  rw.QueueConfig.RetryOn429,
			}
		}

		if rw.HTTPClientConfig != nil {
			rwc.HTTPClientConfig = *rw.HTTPClientConfig.Convert()
		}

		rwc.SendExemplars = rw.SendExemplars

		rwConfigs = append(rwConfigs, rwc)
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
