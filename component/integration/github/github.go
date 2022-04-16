package github

import (
	"context"
	"net/http"
	"path"
	"sync"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/component"
	metricsscraper "github.com/grafana/agent/component/metrics-scraper"
	"github.com/hashicorp/hcl/v2"
	gh_config "github.com/infinityworks/github-exporter/config"
	gh_exporter "github.com/infinityworks/github-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/rfratto/gohcl"
)

func init() {
	component.Register(component.Registration{
		Name:   "integration.github",
		Config: Config{},
		BuildComponent: func(o component.Options, c component.Config) (component.Component, error) {
			return NewComponent(o, c.(Config))
		},
	})
}

// Config represents the input state of the integration.github component.
type Config struct {
	APIURL        string   `hcl:"api_url,optional"`
	Repositories  []string `hcl:"repositories,optional"`
	Organizations []string `hcl:"organizations,optional"`
	Users         []string `hcl:"users,optional"`
	APIToken      string   `hcl:"api_token,optional"`
}

var DefaultConfig = Config{
	APIURL: "https://api.github.com",
}

var _ gohcl.Decoder = (*Config)(nil)

func (c *Config) DecodeHCL(body hcl.Body, ctx *hcl.EvalContext) error {
	*c = DefaultConfig

	type config Config
	return gohcl.DecodeBody(body, ctx, (*config)(c))
}

// State represents the output state of the integration.github component.
type State struct {
	Targets []metricsscraper.TargetGroup `hcl:"targets"`
}

// Component is the integration.github component.
type Component struct {
	log  log.Logger
	reg  *prometheus.Registry
	opts component.Options

	mut           sync.RWMutex
	cfg           Config
	prevCollector prometheus.Collector
}

// NewComponent creates a new integration.github component.
func NewComponent(o component.Options, c Config) (*Component, error) {
	res := &Component{
		log:  o.Logger,
		reg:  prometheus.NewRegistry(),
		opts: o,
	}
	if err := res.Update(c); err != nil {
		return nil, err
	}
	return res, nil
}

var _ component.Component = (*Component)(nil)

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update implements UpdatableComponent.
func (c *Component) Update(newConfig component.Config) error {
	cfg := newConfig.(Config)

	c.mut.Lock()
	defer c.mut.Unlock()

	var exporterConf gh_config.Config

	if err := exporterConf.SetAPIURL(cfg.APIURL); err != nil {
		return err
	}
	exporterConf.SetRepositories(cfg.Repositories)
	exporterConf.SetUsers(cfg.Users)
	exporterConf.SetOrganisations(cfg.Organizations)
	exporterConf.SetAPIToken(cfg.APIToken)

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
				model.MetricsPathLabel: path.Join(component.HTTPPrefix(c.opts.ComponentID), "/metrics"),
			}},
			Labels: metricsscraper.LabelSet{
				model.InstanceLabel: c.opts.ComponentID,
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
