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

type MetricsReceiver struct{ storage.Appender }

type Config struct {
	RemoteWrite []*RemoteWriteConfig `hcl:"remote_write,block" cty:"remote_write"`
}

type RemoteWriteConfig struct {
	URL       string           `hcl:"url" cty:"url"`
	BasicAuth *BasicAuthConfig `hcl:"basic_auth,block" cty:"basic_auth"`
}

type BasicAuthConfig struct {
	Username string `hcl:"username" cty:"username"`
	Password string `hcl:"password" cty:"password"`
}

type State struct {
	Receiver *MetricsReceiver `hcl:"receiver" cty:"receiver"`
}

type Component struct {
	log log.Logger
}

func NewComponent(l log.Logger, c Config) (*Component, error) {
	spew.Dump(c)
	return &Component{log: l}, nil
}

var _ component.Component[Config] = (*Component)(nil)

func (c *Component) Run(ctx context.Context, onStateChange func()) error {
	level.Info(c.log).Log("msg", "component starting")
	defer level.Info(c.log).Log("msg", "component shutting down")

	<-ctx.Done()
	return nil
}

func (c *Component) Update(cfg Config) error {
	spew.Dump(cfg)
	return nil
}

func (c *Component) CurrentState() interface{} {
	return State{}
}

func (c *Component) Config() Config {
	return Config{}
}
