package metricsforwarder

import (
	"context"
	"reflect"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/prometheus/prometheus/storage"
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
type MetricsReceiver struct{ storage.Appender }

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

	mut sync.RWMutex
	cfg Config
}

// NewComponent creates a new metrics_forwarder component.
func NewComponent(o component.Options, c Config) (*Component, error) {
	res := &Component{log: o.Logger, cfg: c}
	if err := res.Update(c); err != nil {
		return nil, err
	}
	return res, nil
}

var _ component.Component[Config] = (*Component)(nil)

// Run implements Component.
func (c *Component) Run(ctx context.Context, onStateChange func()) error {
	level.Info(c.log).Log("msg", "component starting")
	defer level.Info(c.log).Log("msg", "component shutting down")

	<-ctx.Done()
	return nil
}

// Update implements UdpatableComponent.
func (c *Component) Update(cfg Config) error {
	c.mut.Lock()
	defer c.mut.Unlock()
	spew.Dump(cfg)
	c.cfg = cfg
	return nil
}

// CurrentState implements Component.
func (c *Component) CurrentState() interface{} {
	return State{
		&MetricsReceiver{},
	}
}

// Config implements Component.
func (c *Component) Config() Config {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.cfg
}
