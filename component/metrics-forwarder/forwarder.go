package metricsforwarder

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/metrics/wal"
	common "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
)

func init() {
	component.Register(component.Registration[Config]{
		Name: "metrics_forwarder",
		BuildComponent: func(o component.Options, c Config) (component.Component[Config], error) {
			return NewComponent(o, c)
		},
	})

	component.RegisterComplexType("MetricsReceiver", reflect.TypeOf(MetricsReceiver{}))
}

// MetricsReceiver is the type used by the metrics_forwarder component to
// receive metrics to write to a WAL.
type MetricsReceiver struct{ storage.Appendable }

// Config represents the input state of the metrics_forwarder component.
type Config struct {
	RemoteWrite []*RemoteWriteConfig `hcl:"remote_write,block"`
}

// RemoteWriteConfig is the metrics_fowarder's configuration for where to send
// metrics stored in the WAL.
type RemoteWriteConfig struct {
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
	log log.Logger

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
	remoteStore := remote.NewStorage(remoteLogger, nil, startTime, dataPath, 10*time.Second, nil)

	res := &Component{
		log: o.Logger,

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

var _ component.Component[Config] = (*Component)(nil)

// Run implements Component.
func (c *Component) Run(ctx context.Context, onStateChange func()) error {
	defer func() {
		err := c.storage.Close()
		if err != nil {
			level.Error(c.log).Log("msg", "error when closing storage", "err", err)
		}
	}()

	level.Info(c.log).Log("msg", "component starting")
	defer level.Info(c.log).Log("msg", "component shutting down")

	// TODO(rfratto): truncate WAL / GC on a loop

	<-ctx.Done()
	return nil
}

// Update implements UdpatableComponent.
func (c *Component) Update(cfg Config) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	var rwConfigs []*config.RemoteWriteConfig
	for _, rw := range cfg.RemoteWrite {
		parsedURL, err := url.Parse(rw.URL)
		if err != nil {
			return fmt.Errorf("cannot parse remote_write url %q: %w", rw.URL, err)
		}

		rwc := &config.RemoteWriteConfig{
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
				Username: rw.BasicAuth.Username,
				Password: common.Secret(rw.BasicAuth.Password),
			}
		}

		rwConfigs = append(rwConfigs, rwc)
	}

	err := c.remoteStore.ApplyConfig(&config.Config{
		// TODO(rfratto): external labels
		RemoteWriteConfigs: rwConfigs,
	})
	if err != nil {
		return err
	}

	c.cfg = cfg
	return nil
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
