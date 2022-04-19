package integrations

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/scheduler"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/agentflow/config"
	"github.com/grafana/agent/pkg/agentflow/types"
	"github.com/grafana/agent/pkg/agentflow/types/actorstate"
	"github.com/grafana/agent/pkg/agentflow/types/exchange"
	gh_config "github.com/infinityworks/github-exporter/config"
	"github.com/infinityworks/github-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
	"time"
)

type Github struct {
	self     *actor.PID
	outs     []*actor.PID
	name     string
	cfg      *config.Github
	cancel   scheduler.CancelFunc
	exporter *exporter.Exporter
	registry *prometheus.Registry
	router   *mux.Router

	log  log.Logger
	root *actor.RootContext
}

func NewGithub(name string, cfg *config.Github, global types.Global) (actorstate.FlowActor, error) {
	if cfg.ApiURL == "" {
		cfg.ApiURL = "https://api.github.com"
	}
	conf := gh_config.Config{}
	err := conf.SetAPIURL(cfg.ApiURL)
	if err != nil {
		level.Error(global.Log).Log("msg", "api url is invalid", "err", err)
		return nil, err
	}
	conf.SetRepositories(cfg.Repositories)

	ghExporter := exporter.Exporter{
		APIMetrics: exporter.AddMetrics(),
		Config:     conf,
	}
	r := prometheus.NewRegistry()
	err = r.Register(&ghExporter)
	if err != nil {
		level.Error(global.Log).Log("error", err)
	}

	return &Github{
		name:     name,
		cfg:      cfg,
		exporter: &ghExporter,
		registry: r,
		router:   global.Mux,
		root:     global.RootContext,
	}, nil
}

func (g *Github) PID() *actor.PID {
	return g.self
}

func (g *Github) AllowableInputs() []actorstate.InOutType {
	return []actorstate.InOutType{actorstate.None}
}

func (g *Github) Output() actorstate.InOutType {
	return actorstate.Metrics
}

func (g *Github) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case actorstate.Init:
		g.outs = msg.Children
	case actorstate.Start:
		g.self = c.Self()
		sched := scheduler.NewTimerScheduler(c)
		g.cancel = sched.SendRepeatedly(1*time.Millisecond, 60*time.Second, c.Self(), "scrape")
		if g.cfg.EnableEndpoint {
			g.router.Handle("/"+g.name+"/metrics", promhttp.HandlerFor(g.registry, promhttp.HandlerOpts{}))
		}
	case actorstate.State:
		bb, _ := yaml.Marshal(&githubState{
			Cfg:    g.cfg,
			Status: "Healthy",
		})
		c.Respond(bb)
	case string:
		if msg != "scrape" {
			return
		}
		metrics, err := g.registry.Gather()
		if err != nil {
			level.Error(g.log).Log("error", err)
			return
		}
		ms := make([]exchange.Metric, 0)
		for _, m := range metrics {
			pogom := exchange.CopyMetricFromPrometheus(m)
			ms = append(ms, pogom)
		}
		if len(ms) == 0 {
			return
		}
		for _, o := range g.outs {
			c.Send(o, ms)
		}
	}
}

func (g *Github) Name() string {
	return g.name
}

type githubState struct {
	Cfg    *config.Github
	Status string
}
