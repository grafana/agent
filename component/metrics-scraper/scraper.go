package metricsscraper

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	metricsforwarder "github.com/grafana/agent/component/metrics-forwarder"
)

func init() {
	component.Register(component.Registration[Config]{
		Name: "metrics_scraper",
		BuildComponent: func(l log.Logger, c Config) (component.Component[Config], error) {
			return NewComponent(l, c)
		},
	})
}

type Config struct {
	Targets        []TargetGroup              `hcl:"targets" cty:"targets"`
	ScrapeInterval string                     `hcl:"scrape_interval,optional" cty:"scrape_interval"`
	ScrapeTimeout  string                     `hcl:"scrape_timeout,optional" cty:"scrape_timeout"`
	SendTo         *metricsforwarder.Appender `hcl:"send_to" cty:"send_to"`
}

// TargetGroup is a set of targets that share a common set of labels.
type TargetGroup struct {
	Targets []LabelSet `hcl:"targets" cty:"targets"`
	Labels  LabelSet   `hcl:"labels,optional" cty:"labels"`
}

// LabelSet is a map of label names to values.
type LabelSet map[string]string

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
	return nil
}

func (c *Component) Config() Config {
	return Config{}
}
