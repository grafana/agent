package metricsscraper

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	metricsforwarder "github.com/grafana/agent/component/metrics-forwarder"
	"github.com/hashicorp/hcl/v2"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"github.com/rfratto/gohcl"
)

func init() {
	component.Register(component.Registration{
		Name:   "metrics_scraper",
		Config: Config{},
		BuildComponent: func(o component.Options, c component.Config) (component.Component, error) {
			return NewComponent(o, c.(Config))
		},
	})
}

// Config represents the input state of the metrics_scraper component.
type Config struct {
	ScrapeInterval time.Duration                     `hcl:"scrape_interval,optional"`
	ScrapeTimeout  time.Duration                     `hcl:"scrape_timeout,optional"`
	Targets        []TargetGroup                     `hcl:"targets"`
	SendTo         *metricsforwarder.MetricsReceiver `hcl:"send_to"`
}

var DefaultConfig = Config{
	ScrapeInterval: time.Duration(60 * time.Second),
	ScrapeTimeout:  time.Duration(10 * time.Second),
}

var _ gohcl.Decoder = (*Config)(nil)

func (c *Config) DecodeHCL(body hcl.Body, ctx *hcl.EvalContext) error {
	*c = DefaultConfig

	type config Config
	return gohcl.DecodeBody(body, ctx, (*config)(c))
}

// TargetGroup is a set of targets that share a common set of labels.
type TargetGroup struct {
	Targets []LabelSet `hcl:"targets"`
	Labels  LabelSet   `hcl:"labels,optional"`
}

// LabelSet is a map of label names to values.
type LabelSet map[string]string

// Component is the metrics_scraper component.
type Component struct {
	log log.Logger
	id  string

	mut sync.RWMutex
	cfg Config

	newTargets chan struct{}
	scraper    *scrape.Manager
	app        *lazyAppendable
}

// NewComponent creates a new metrics_scraper component.
func NewComponent(o component.Options, c Config) (*Component, error) {
	app := &lazyAppendable{id: o.ComponentID}

	scrapeLogger := log.With(o.Logger, "subcomponent", "scrape")
	scraper := scrape.NewManager(&scrape.Options{}, scrapeLogger, app)

	res := &Component{
		log: o.Logger,
		id:  o.ComponentID,

		app:        app,
		scraper:    scraper,
		newTargets: make(chan struct{}, 1),
	}
	if err := res.Update(c); err != nil {
		return nil, err
	}
	return res, nil
}

var _ component.Component = (*Component)(nil)

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	defer c.scraper.Stop()

	level.Info(c.log).Log("msg", "component starting")
	defer level.Info(c.log).Log("msg", "component shutting down")

	targetChan := make(chan map[string][]*targetgroup.Group)

	go func() {
		// TODO(rfratto): how do we get targets to this thing?
		err := c.scraper.Run(targetChan)
		if err != nil {
			level.Error(c.log).Log("msg", "scraper failed", "err", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.newTargets:
			c.mut.RLock()
			targets := c.cfg.Targets
			c.mut.RUnlock()

			// Try to send the targets
			promTargets := c.convertTargets(targets)
			select {
			case <-ctx.Done():
			case targetChan <- promTargets:
				level.Debug(c.log).Log("msg", "passed targets to scrape manager", "count", len(targets))
			}
		}
	}
}

func (c *Component) convertTargets(groups []TargetGroup) map[string][]*targetgroup.Group {
	var promGroups []*targetgroup.Group

	for _, g := range groups {
		var promGroup targetgroup.Group
		for _, target := range g.Targets {
			promGroup.Targets = append(promGroup.Targets, convertLabelSet(target))
		}
		promGroup.Labels = convertLabelSet(g.Labels)
		promGroup.Source = c.id
		promGroups = append(promGroups, &promGroup)
	}

	return map[string][]*targetgroup.Group{c.id: promGroups}
}

func convertLabelSet(in LabelSet) model.LabelSet {
	out := make(model.LabelSet, len(in))
	for k, v := range in {
		out[model.LabelName(k)] = model.LabelValue(v)
	}
	return out
}

// Update implements UpdatableComponent.
func (c *Component) Update(newConfig component.Config) error {
	cfg := newConfig.(Config)

	c.mut.Lock()
	defer c.mut.Unlock()

	// TODO(rfratto): expose other config (HTTPClientConfig, Relabel)
	sc := config.DefaultScrapeConfig
	sc.JobName = c.id
	sc.ScrapeInterval = model.Duration(cfg.ScrapeInterval)
	sc.ScrapeTimeout = model.Duration(cfg.ScrapeTimeout)

	err := c.scraper.ApplyConfig(&config.Config{
		ScrapeConfigs: []*config.ScrapeConfig{&sc},
	})
	if err != nil {
		return fmt.Errorf("error applying targets: %w", err)
	}

	c.app.Set(cfg.SendTo)
	c.cfg = cfg

	select {
	case c.newTargets <- struct{}{}:
	default:
	}
	return nil
}

// CurrentState implements Component.
func (c *Component) CurrentState() interface{} {
	return nil
}

// Config implements Component.
func (c *Component) Config() Config {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.cfg
}

type lazyAppendable struct {
	id    string
	mut   sync.RWMutex
	inner storage.Appendable
}

var _ storage.Appendable = (*lazyAppendable)(nil)

func (la *lazyAppendable) Appender(ctx context.Context) storage.Appender {
	la.mut.RLock()
	defer la.mut.RUnlock()

	if la.inner == nil {
		return &failedAppender{id: la.id}
	}

	return la.inner.Appender(ctx)
}

func (la *lazyAppendable) Set(app storage.Appendable) {
	la.mut.Lock()
	defer la.mut.Unlock()
	la.inner = app
}

type failedAppender struct{ id string }

var _ storage.Appender = (*failedAppender)(nil)

func (fa *failedAppender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	return 0, fmt.Errorf("component %s does not have a configured MetricsReceiver to send samples to", fa.id)
}

func (fa *failedAppender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	return 0, fmt.Errorf("component %s does not have a configured MetricsReceiver to send examplars to", fa.id)
}

func (fa *failedAppender) Commit() error { return nil }

func (fa *failedAppender) Rollback() error { return nil }
