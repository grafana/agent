package scrape

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/metrics"
	"github.com/grafana/agent/pkg/build"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/scrape"
)

func init() {
	scrape.UserAgent = fmt.Sprintf("GrafanaAgent/%s", build.Version)

	component.Register(component.Registration{
		Name: "metrics.scrape",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
	component.RegisterGoStruct("MetricsReceiver", metrics.Receiver{})
}

// Arguments holds values which are used to configure the metrics.scrape
// component.
type Arguments struct {
	Targets   []Target            `hcl:"targets"`
	Receivers []*metrics.Receiver `hcl:"receivers"`

	// Scrape Options
	ExtraMetrics bool `hcl:"extra_metrics,optional"`
	// TODO(@tpaschalis) enable HTTPClientOptions []config_util.HTTPClientOption

	// Scrape Config
	ScrapeConfigs []Config `hcl:"scrape_config,block"`
}

// Target refers to a singular HTTP or HTTPS endpoint that will be used for
// scraping. Here, we're using a map[string]string instead of labels.Labels;
// if the label ordering is important, we can change to follow the upstream
// logic instead.
type Target map[string]string

// Component implements the metrics.Scrape component.
type Component struct {
	opts component.Options

	reloadTargets chan struct{}

	mut        sync.RWMutex
	args       Arguments
	scraper    *scrape.Manager
	appendable *flowAppendable
}

var (
	_ component.Component = (*Component)(nil)
)

// New creates a new metrics.scrape component.
func New(o component.Options, args Arguments) (*Component, error) {
	flowAppendable := newFlowAppendable(args.Receivers...)

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
			scs := c.args.ScrapeConfigs
			c.mut.RUnlock()
			promTargets := c.hclTargetsToProm(scs, tgs)

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

	c.appendable.receivers = newArgs.Receivers

	scs := c.getPromScrapeConfigs(newArgs.ScrapeConfigs)
	err := c.scraper.ApplyConfig(&config.Config{
		ScrapeConfigs: scs,
	})
	if err != nil {
		return fmt.Errorf("error applying scrape configs: %w", err)
	}
	level.Debug(c.opts.Logger).Log("msg", "applied new scrape configs", "len", len(scs))

	select {
	case c.reloadTargets <- struct{}{}:
	default:
	}

	return nil
}

// ScraperStatus reports the status of the scraper's jobs.
type ScraperStatus struct {
	TargetStatus []TargetStatus `hcl:"target,block"`
}

// TargetStatus reports on the status of the latest scrape for a target.
type TargetStatus struct {
	JobName            string            `hcl:"job"`
	URL                string            `hcl:"url"`
	Health             string            `hcl:"health"`
	Labels             map[string]string `hcl:"labels,attr"`
	LastError          string            `hcl:"last_error,optional"`
	LastScrape         time.Time         `hcl:"last_scrape"`
	LastScrapeDuration time.Duration     `hcl:"last_scrape_duration,optional"`
}

// DebugInfo implements component.DebugComponent
func (c *Component) DebugInfo() interface{} {
	var res []TargetStatus

	for job, stt := range c.scraper.TargetsActive() {
		for _, st := range stt {
			if st != nil {
				res = append(res, TargetStatus{
					JobName:            job,
					URL:                st.URL().String(),
					Health:             string(st.Health()),
					Labels:             st.Labels().Map(),
					LastError:          st.LastError().Error(),
					LastScrape:         st.LastScrape(),
					LastScrapeDuration: st.LastScrapeDuration(),
				})
			}
		}
	}

	return ScraperStatus{TargetStatus: res}
}

func (c *Component) hclTargetsToProm(scs []Config, tgs []Target) map[string][]*targetgroup.Group {
	res := make(map[string][]*targetgroup.Group, len(scs))
	for _, sc := range scs {
		promGroup := &targetgroup.Group{Source: c.opts.ID}
		for _, tg := range tgs {
			promGroup.Targets = append(promGroup.Targets, convertLabelSet(tg))
		}
		targetGroupName := c.opts.ID + "/" + sc.JobName
		res[targetGroupName] = []*targetgroup.Group{promGroup}
	}
	return res
}

func convertLabelSet(tg Target) model.LabelSet {
	lset := make(model.LabelSet, len(tg))
	for k, v := range tg {
		lset[model.LabelName(k)] = model.LabelValue(v)
	}
	return lset
}
