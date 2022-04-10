package metricsforwarder

import (
	"context"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/prometheus/prometheus/storage"
)

func init() {
	component.Register(component.Registration[Config]{
		Name: "metrics_forwarder",
		BuildComponent: func(l log.Logger, c Config) (component.Component[Config], error) {
			return NewComponent(l, c)
		},
	})

	component.RegisterComplexType("MetricsReceiver", reflect.TypeOf(MetricsReceiver{}))
}

// MetricsReceiver is the type used by the metrics_forwarder component to
// receive metrics to write to a WAL.
type MetricsReceiver struct{ storage.Appender }

// Config represents the input state of the metrics_forwarder component.
type Config struct {
	RemoteWrite []*RemoteWriteConfig `hcl:"remote_write,block" cty:"remote_write"`
}

// RemoteWriteConfig is the metrics_fowarder's configuration for where to send
// metrics stored in the WAL.
type RemoteWriteConfig struct {
	URL       string           `hcl:"url" cty:"url"`
	BasicAuth *BasicAuthConfig `hcl:"basic_auth,block" cty:"basic_auth"`
}

// BasicAuthConfig is the metrics_forwarder's configuration for authenticating
// against the remote system when sending metrics.
type BasicAuthConfig struct {
	Username string `hcl:"username" cty:"username"`
	Password string `hcl:"password" cty:"password"`
}

// State represents the output state of the metrics_forwarder component.
type State struct {
	Receiver *MetricsReceiver `hcl:"receiver" cty:"receiver"`
}

// Component is the metrics_forwarder component.
type Component struct {
	log log.Logger
}

// NewComponent creates a new metrics_forwarder component.
func NewComponent(l log.Logger, c Config) (*Component, error) {
	spew.Dump(c)
	return &Component{log: l}, nil
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
	spew.Dump(cfg)
	return nil
}

// CurrentState implements Component.
func (c *Component) CurrentState() interface{} {
	return State{}
}

// Config implements Component.
func (c *Component) Config() Config {
	return Config{}
}
