package integrations

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/scheduler"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/agentflow/config"
	"github.com/grafana/agent/pkg/agentflow/types"
	"github.com/grafana/agent/pkg/agentflow/types/actorstate"
	"github.com/grafana/agent/pkg/agentflow/types/exchange"
	gh_config "github.com/infinityworks/github-exporter/config"
	"github.com/infinityworks/github-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
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

	log log.Logger
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
	}, nil
}

func (a *Github) PID() *actor.PID {
	return a.self
}

func (a *Github) AllowableInputs() []actorstate.InOutType {
	return []actorstate.InOutType{actorstate.None}
}

func (a *Github) Output() actorstate.InOutType {
	return actorstate.Metrics
}

func (a *Github) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case actorstate.Init:
		a.outs = msg.Children
	case actorstate.Start:
		a.self = c.Self()
		sched := scheduler.NewTimerScheduler(c)
		a.cancel = sched.SendRepeatedly(1*time.Millisecond, 60*time.Second, c.Self(), "scrape")
	case string:
		if msg != "scrape" {
			return
		}
		metrics, err := a.registry.Gather()
		if err != nil {
			level.Error(a.log).Log("error", err)
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
		for _, o := range a.outs {
			c.Send(o, ms)
		}
	}
}

func (m *Github) Name() string {
	return m.name
}
