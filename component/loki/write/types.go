package write

import (
	"fmt"
	"net/url"
	"time"

	"github.com/grafana/agent/component/common/loki/client"

	"github.com/alecthomas/units"
	types "github.com/grafana/agent/component/common/config"
	"github.com/grafana/dskit/backoff"
	"github.com/grafana/dskit/flagext"
	lokiflagext "github.com/grafana/loki/pkg/util/flagext"
	"github.com/prometheus/common/model"
)

// EndpointOptions describes an individual location to send logs to.
type EndpointOptions struct {
	Name              string                  `river:"name,attr,optional"`
	URL               string                  `river:"url,attr"`
	BatchWait         time.Duration           `river:"batch_wait,attr,optional"`
	BatchSize         units.Base2Bytes        `river:"batch_size,attr,optional"`
	RemoteTimeout     time.Duration           `river:"remote_timeout,attr,optional"`
	Headers           map[string]string       `river:"headers,attr,optional"`
	MinBackoff        time.Duration           `river:"min_backoff_period,attr,optional"`  // start backoff at this level
	MaxBackoff        time.Duration           `river:"max_backoff_period,attr,optional"`  // increase exponentially to this level
	MaxBackoffRetries int                     `river:"max_backoff_retries,attr,optional"` // give up after this many; zero means infinite retries
	TenantID          string                  `river:"tenant_id,attr,optional"`
	HTTPClientConfig  *types.HTTPClientConfig `river:",squash"`
}

// GetDefaultEndpointOptions defines the default settings for sending logs to a
// remote endpoint.
// The backoff schedule with the default parameters:
// 0.5s, 1s, 2s, 4s, 8s, 16s, 32s, 64s, 128s, 256s(4.267m)
// For a total time of 511.5s (8.5m) before logs are lost.
func GetDefaultEndpointOptions() EndpointOptions {
	var defaultEndpointOptions = EndpointOptions{
		BatchWait:         1 * time.Second,
		BatchSize:         1 * units.MiB,
		RemoteTimeout:     10 * time.Second,
		MinBackoff:        500 * time.Millisecond,
		MaxBackoff:        5 * time.Minute,
		MaxBackoffRetries: 10,
		HTTPClientConfig:  types.CloneDefaultHTTPClientConfig(),
	}

	return defaultEndpointOptions
}

// UnmarshalRiver implements river.Unmarshaler.
func (r *EndpointOptions) UnmarshalRiver(f func(v interface{}) error) error {
	*r = GetDefaultEndpointOptions()

	type arguments EndpointOptions
	if err := f((*arguments)(r)); err != nil {
		return err
	}

	if _, err := url.Parse(r.URL); err != nil {
		return fmt.Errorf("failed to parse remote url %q: %w", r.URL, err)
	}

	// We must explicitly Validate because HTTPClientConfig is squashed and it won't run otherwise
	if r.HTTPClientConfig != nil {
		return r.HTTPClientConfig.Validate()
	}

	return nil
}

func (args Arguments) convertClientConfigs() []client.Config {
	var res []client.Config
	for _, cfg := range args.Endpoints {
		url, _ := url.Parse(cfg.URL)
		cc := client.Config{
			Name:      cfg.Name,
			URL:       flagext.URLValue{URL: url},
			Headers:   cfg.Headers,
			BatchWait: cfg.BatchWait,
			BatchSize: int(cfg.BatchSize),
			Client:    *cfg.HTTPClientConfig.Convert(),
			BackoffConfig: backoff.Config{
				MinBackoff: cfg.MinBackoff,
				MaxBackoff: cfg.MaxBackoff,
				MaxRetries: cfg.MaxBackoffRetries,
			},
			ExternalLabels: lokiflagext.LabelSet{LabelSet: toLabelSet(args.ExternalLabels)},
			Timeout:        cfg.RemoteTimeout,
			TenantID:       cfg.TenantID,
		}
		res = append(res, cc)
	}

	return res
}

func toLabelSet(in map[string]string) model.LabelSet {
	res := make(model.LabelSet, len(in))
	for k, v := range in {
		res[model.LabelName(k)] = model.LabelValue(v)
	}
	return res
}
