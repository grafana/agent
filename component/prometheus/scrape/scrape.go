package scrape

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	fa "github.com/grafana/agent/component/common/appendable"
	"github.com/grafana/agent/component/discovery"
	flow_prometheus "github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/pkg/build"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/scrape"
)

func init() {
	scrape.UserAgent = fmt.Sprintf("GrafanaAgent/%s", build.Version)

	component.Register(component.Registration{
		Name: "prometheus.scrape",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the prometheus.scrape
// component.
type Arguments struct {
	Targets   []discovery.Target          `river:"targets,attr"`
	ForwardTo []*flow_prometheus.Receiver `river:"forward_to,attr"`

	ScrapeConfig Config `river:"scrape_config,block"`

	// Scrape Options
	ExtraMetrics bool `river:"extra_metrics,attr,optional"`
}

// Component implements the prometheus.scrape component.
type Component struct {
	opts component.Options

	reloadTargets chan struct{}

	mut        sync.RWMutex
	args       Arguments
	scraper    *scrape.Manager
	appendable *fa.FlowAppendable
}

var (
	_ component.Component = (*Component)(nil)
)

// New creates a new prometheus.scrape component.
func New(o component.Options, args Arguments) (*Component, error) {
	flowAppendable := fa.NewFlowAppendable(args.ForwardTo...)

	scrapeOptions := &scrape.Options{ExtraMetrics: args.ExtraMetrics}
	scraper := scrape.NewManager(scrapeOptions, o.Logger, flowAppendable)
	c := &Component{
		opts:          o,
		reloadTargets: make(chan struct{}, 1),
		scraper:       scraper,
		appendable:    flowAppendable,
	}

	// Call to Update() to set the receivers and targets once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	defer c.scraper.Stop()

	targetSetsChan := make(chan map[string][]*targetgroup.Group)

	go func() {
		err := c.scraper.Run(targetSetsChan)
		level.Info(c.opts.Logger).Log("msg", "scrape manager stopped")
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "scrape manager failed", "err", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.reloadTargets:
			c.mut.RLock()
			tgs := c.args.Targets
			c.mut.RUnlock()
			promTargets := c.componentTargetsToProm(tgs)

			select {
			case targetSetsChan <- promTargets:
				level.Debug(c.opts.Logger).Log("msg", "passed new targets to scrape manager")
			case <-ctx.Done():
			}
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()
	c.args = newArgs

	c.appendable.SetReceivers(newArgs.ForwardTo)

	sc, err := newArgs.ScrapeConfig.getPromScrapeConfigs(c.opts.ID)
	if err != nil {
		return fmt.Errorf("invalid scrape_config: %w", err)
	}

	err = c.scraper.ApplyConfig(&config.Config{
		ScrapeConfigs: []*config.ScrapeConfig{sc},
	})
	if err != nil {
		return fmt.Errorf("error applying scrape configs: %w", err)
	}
	level.Debug(c.opts.Logger).Log("msg", "scrape config was updated")

	select {
	case c.reloadTargets <- struct{}{}:
	default:
	}

	return nil
}

// ScraperStatus reports the status of the scraper's jobs.
type ScraperStatus struct {
	TargetStatus []TargetStatus `river:"target,block,optional"`
}

// TargetStatus reports on the status of the latest scrape for a target.
type TargetStatus struct {
	JobName            string            `river:"job,attr"`
	URL                string            `river:"url,attr"`
	Health             string            `river:"health,attr"`
	Labels             map[string]string `river:"labels,attr"`
	LastError          string            `river:"last_error,attr,optional"`
	LastScrape         time.Time         `river:"last_scrape,attr"`
	LastScrapeDuration time.Duration     `river:"last_scrape_duration,attr,optional"`
}

// DebugInfo implements component.DebugComponent
func (c *Component) DebugInfo() interface{} {
	var res []TargetStatus

	for job, stt := range c.scraper.TargetsActive() {
		for _, st := range stt {
			var lastError string
			if st.LastError() != nil {
				lastError = st.LastError().Error()
			}
			if st != nil {
				res = append(res, TargetStatus{
					JobName:            job,
					URL:                st.URL().String(),
					Health:             string(st.Health()),
					Labels:             st.Labels().Map(),
					LastError:          lastError,
					LastScrape:         st.LastScrape(),
					LastScrapeDuration: st.LastScrapeDuration(),
				})
			}
		}
	}

	return ScraperStatus{TargetStatus: res}
}

func (c *Component) componentTargetsToProm(tgs []discovery.Target) map[string][]*targetgroup.Group {
	promGroup := &targetgroup.Group{Source: c.opts.ID}
	for _, tg := range tgs {
		promGroup.Targets = append(promGroup.Targets, convertLabelSet(tg))
	}

	return map[string][]*targetgroup.Group{c.opts.ID: {promGroup}}
}

func convertLabelSet(tg discovery.Target) model.LabelSet {
	lset := make(model.LabelSet, len(tg))
	for k, v := range tg {
		lset[model.LabelName(k)] = model.LabelValue(v)
	}
	return lset
}
