package metricsforwarder

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/metrics/wal"
	common "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/timestamp"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
)

// Options.
//
// TODO(rfratto): These should be flags. How do we want to handle static
// options for components?
var (
	walTruncateFrequency = 2 * time.Hour
	minWALTime           = 5 * time.Minute
	maxWALTime           = 8 * time.Hour
	remoteFlushDeadline  = 1 * time.Minute
)

func init() {
	remote.UserAgent = fmt.Sprintf("GrafanaAgent/%s", build.Version)
	config.DefaultRemoteWriteConfig.SendExemplars = true

	component.Register(component.Registration{
		Name:   "metrics_forwarder",
		Config: Config{},
		BuildComponent: func(o component.Options, c component.Config) (component.Component, error) {
			return NewComponent(o, c.(Config))
		},
	})

	component.RegisterComplexType("MetricsReceiver", reflect.TypeOf(MetricsReceiver{}))
}

// MetricsReceiver is the type used by the metrics_forwarder component to
// receive metrics to write to a WAL.
type MetricsReceiver struct{ storage.Appendable }

// Config represents the input state of the metrics_forwarder component.
type Config struct {
	ExternalLabels map[string]string    `hcl:"external_labels,optional"`
	RemoteWrite    []*RemoteWriteConfig `hcl:"remote_write,block"`
}

// RemoteWriteConfig is the metrics_fowarder's configuration for where to send
// metrics stored in the WAL.
type RemoteWriteConfig struct {
	Name      string           `hcl:"name,optional"`
	URL       string           `hcl:"url"`
	BasicAuth *BasicAuthConfig `hcl:"basic_auth,block"`
}

// BasicAuthConfig is the metrics_forwarder's configuration for authenticating
// against the remote system when sending metrics.
type BasicAuthConfig struct {
	Username string `hcl:"username"`
	Password string `hcl:"password"`
}

// State represents the output state of the metrics_forwarder component.
type State struct {
	Receiver *MetricsReceiver `hcl:"receiver"`
}

// Component is the metrics_forwarder component.
type Component struct {
	log  log.Logger
	opts component.Options

	walStore    *wal.Storage
	remoteStore *remote.Storage
	storage     storage.Storage

	mut sync.RWMutex
	cfg Config
}

// NewComponent creates a new metrics_forwarder component.
func NewComponent(o component.Options, c Config) (*Component, error) {
	// TODO(rfratto): don't hardcode base path
	walLogger := log.With(o.Logger, "subcomponent", "wal")
	dataPath := filepath.Join("data-agent", o.ComponentID)
	walStorage, err := wal.NewStorage(walLogger, nil, filepath.Join("data-agent", o.ComponentID))
	if err != nil {
		return nil, err
	}

	remoteLogger := log.With(o.Logger, "subcomponent", "rw")
	remoteStore := remote.NewStorage(remoteLogger, nil, startTime, dataPath, remoteFlushDeadline, nil)

	res := &Component{
		log:  o.Logger,
		opts: o,

		walStore:    walStorage,
		remoteStore: remoteStore,
		storage:     storage.NewFanout(o.Logger, walStorage, remoteStore),
	}
	if err := res.Update(c); err != nil {
		return nil, err
	}
	return res, nil
}

func startTime() (int64, error) { return 0, nil }

var _ component.Component = (*Component)(nil)

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	defer func() {
		level.Debug(c.log).Log("msg", "closing storage")
		err := c.storage.Close()
		level.Debug(c.log).Log("msg", "storage closed")
		if err != nil {
			level.Error(c.log).Log("msg", "error when closing storage", "err", err)
		}
	}()

	// Track the last timestamp we truncated for to prevent segments from getting
	// deleted until at least some new data has been sent.
	var lastTs = int64(math.MinInt64)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(walTruncateFrequency):
			// The timestamp ts is used to determine which series are not receiving
			// samples and may be deleted from the WAL. Their most recent append
			// timestamp is compared to ts, and if that timestamp is older then ts,
			// they are considered inactive and may be deleted.
			//
			// Subtracting a duration from ts will delay when it will be considered
			// inactive and scheduled for deletion.
			ts := c.remoteStore.LowestSentTimestamp() - minWALTime.Milliseconds()
			if ts < 0 {
				ts = 0
			}

			// Network issues can prevent the result of getRemoteWriteTimestamp from
			// changing. We don't want data in the WAL to grow forever, so we set a cap
			// on the maximum age data can be. If our ts is older than this cutoff point,
			// we'll shift it forward to start deleting very stale data.
			if maxTS := timestamp.FromTime(time.Now().Add(-maxWALTime)); ts < maxTS {
				ts = maxTS
			}

			if ts == lastTs {
				level.Debug(c.log).Log("msg", "not truncating the WAL, remote_write timestamp is unchanged", "ts", ts)
				continue
			}
			lastTs = ts

			level.Debug(c.log).Log("msg", "truncating the WAL", "ts", ts)
			err := c.walStore.Truncate(ts)
			if err != nil {
				// The only issue here is larger disk usage and a greater replay time,
				// so we'll only log this as a warning.
				level.Warn(c.log).Log("msg", "could not truncate WAL", "err", err)
			}
		}
	}
}

// getRemoteWriteTimestamp looks up the last successful remote write timestamp.
// This is passed to wal.Storage for its truncation. If no remote write
// sections are configured, getRemoteWriteTimestamp returns the current time.
func (c *Component) getRemoteWriteTimestamp() int64 {
	return c.remoteStore.LowestSentTimestamp()
}

// Update implements Component.
func (c *Component) Update(newConfig component.Config) error {
	cfg := newConfig.(Config)

	c.mut.Lock()
	defer c.mut.Unlock()

	var rwConfigs []*config.RemoteWriteConfig
	for i, rw := range cfg.RemoteWrite {
		parsedURL, err := url.Parse(rw.URL)
		if err != nil {
			return fmt.Errorf("cannot parse remote_write url %q: %w", rw.URL, err)
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

		// Default the name to the index in the file.
		if rwc.Name == "" {
			rwc.Name = fmt.Sprintf("%s[%d]", c.opts.ComponentID, i)
		}

		if rw.BasicAuth != nil {
			rwc.HTTPClientConfig.BasicAuth = &common.BasicAuth{
				Username: rw.BasicAuth.Username,
				Password: common.Secret(rw.BasicAuth.Password),
			}
		}

		rwConfigs = append(rwConfigs, rwc)
	}

	err := c.remoteStore.ApplyConfig(&config.Config{
		GlobalConfig: config.GlobalConfig{
			ExternalLabels: toLabels(cfg.ExternalLabels),
		},
		RemoteWriteConfigs: rwConfigs,
	})
	if err != nil {
		return err
	}

	c.cfg = cfg
	return nil
}

func toLabels(in map[string]string) labels.Labels {
	res := make(labels.Labels, 0, len(in))
	for k, v := range in {
		res = append(res, labels.Label{Name: k, Value: v})
	}
	sort.Sort(res)
	return res
}

// CurrentState implements Component.
func (c *Component) CurrentState() interface{} {
	return State{&MetricsReceiver{Appendable: c.storage}}
}

// Config implements Component.
func (c *Component) Config() Config {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.cfg
}
