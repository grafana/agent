package autoscrape

import (
	"context"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/relabel"
	internal "github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/prometheus/common/model"
)

func init() {
	component.Register(component.Registration{
		Name:    "integrations.v2.autoscrape.global",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	enable          *bool         `river:"enable,bool,optional"`
	metricsInstance string        `river:"metrics_instance,string,optional"`
	scrapeInterval  time.Duration `river:"scrape_interval,attr,optional"`
	scrapeTimeout   time.Duration `river:"scrape_timeout,attr,optional"`
}

type Exports struct {
	Config internal.Global `river:"config,attr"`
}

type Component struct{}

func (c *Component) Run(ctx context.Context) error {
	return nil
}

func (c *Component) Update(args component.Arguments) error {
	return nil
}

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{}

	o.OnStateChange(Exports{Config: args.toInternalConfig()})

	return c, nil
}

func (args *Arguments) toInternalConfig() internal.Global {
	return internal.Global{
		Enable:          *args.enable,
		MetricsInstance: args.metricsInstance,
		ScrapeInterval:  model.Duration(args.scrapeInterval),
		ScrapeTimeout:   model.Duration(args.scrapeTimeout),
	}
}
