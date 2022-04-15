package github

import (
	"context"
	"net/http"
	"path"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/component"
	metricsscraper "github.com/grafana/agent/component/metrics-scraper"
	gh_config "github.com/infinityworks/github-exporter/config"
	gh_exporter "github.com/infinityworks/github-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
)

func init() {
	component.Register(component.Registration[Config]{
		Name: "integration.github",
		BuildComponent: func(o component.Options, c Config) (component.Component[Config], error) {
			return NewComponent(o, c)
		},
	})
}

// Config represents the input state of the integration.github component.
type Config struct {
	Repositories []string `hcl:"repositories,optional"`
}

// State represents the output state of the integration.github component.
type State struct {
	Targets []metricsscraper.TargetGroup `hcl:"targets"`
}

// Component is the integration.github component.
type Component struct {
	log log.Logger
	reg *prometheus.Registry
	id  string

	mut           sync.RWMutex
	cfg           Config
	prevCollector prometheus.Collector
}

// NewComponent creates a new integration.github component.
func NewComponent(o component.Options, c Config) (*Component, error) {
	res := &Component{
		log: o.Logger,
		reg: prometheus.NewRegistry(),
		id:  o.ComponentID,
	}
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

// Update implements UpdatableComponent.
func (c *Component) Update(cfg Config) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	var exporterConf gh_config.Config

	apiURL := "https://api.github.com"
	if err := exporterConf.SetAPIURL(apiURL); err != nil {
		return err
	}
	exporterConf.SetRepositories(cfg.Repositories)

	exporter := &gh_exporter.Exporter{
		APIMetrics: gh_exporter.AddMetrics(),
		Config:     exporterConf,
	}

	if c.prevCollector != nil {
		c.reg.Unregister(c.prevCollector)
	}

	c.reg.Register(exporter)
	c.prevCollector = exporter
	c.cfg = cfg
	return nil
}

// CurrentState implements Component.
func (c *Component) CurrentState() interface{} {
	// TODO(rfratto): remove hard coding
	return State{
		Targets: []metricsscraper.TargetGroup{{
			Targets: []metricsscraper.LabelSet{{
				model.AddressLabel:     "127.0.0.1:12345",
				model.MetricsPathLabel: path.Join(component.HTTPPrefix(c.id), "/metrics"),
			}},
			Labels: metricsscraper.LabelSet{
				model.InstanceLabel: c.id,
				model.JobLabel:      "integration.github",
			},
		}},
	}
}

// Config implements Component.
func (c *Component) Config() Config {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.cfg
}

// ComponentHandler implements HTTPComponent.
func (c *Component) ComponentHandler() (http.Handler, error) {
	r := mux.NewRouter()

	metricsHandler := promhttp.HandlerFor(c.reg, promhttp.HandlerOpts{})

	r.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// We grab a read lock on the mutex to prevent issues when swapping out the
		// collector.
		c.mut.RLock()
		defer c.mut.RUnlock()
		metricsHandler.ServeHTTP(w, r)
	})

	return r, nil
}
