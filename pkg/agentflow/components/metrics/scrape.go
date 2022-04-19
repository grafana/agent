package metrics

import (
	"context"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/agentflow/config"
	"github.com/grafana/agent/pkg/agentflow/types"
	"github.com/grafana/agent/pkg/agentflow/types/actorstate"
	"github.com/grafana/agent/pkg/agentflow/types/exchange"
	"github.com/prometheus/client_golang/prometheus"
	cmnconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	promconfig "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"os"
)

type Scrape struct {
	name           string
	self           *actor.PID
	outs           []*actor.PID
	scraper        *scrape.Manager
	cfg            config.Scraper
	log            log.Logger
	metricRegistry prometheus.Registerer
	cancelFunc     context.CancelFunc
	sd             *discovery.Manager
	cachedMetrics  []exchange.Metric
	rootContext    *actor.RootContext
}

func NewScrape(name string, cfg config.Scraper, global *types.Global) (actorstate.FlowActor, error) {
	return &Scrape{
		cfg:            cfg,
		name:           name,
		log:            global.Log,
		metricRegistry: global.MetricRegistry,
		cachedMetrics:  make([]exchange.Metric, 0),
		rootContext:    global.RootContext,
	}, nil
}

func (s *Scrape) Output() actorstate.InOutType {
	return actorstate.Metrics
}

func (s *Scrape) AllowableInputs() []actorstate.InOutType {
	return []actorstate.InOutType{}
}

func (s *Scrape) Name() string {
	return s.name
}

func (s *Scrape) PID() *actor.PID {
	return s.self
}

func (s *Scrape) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case actorstate.Init:
		s.self = ctx.Self()
		s.outs = msg.Children
	case actorstate.Start:
		s.startScrape()
	case actorstate.Done:
		s.cancelFunc()
	case []exchange.Metric:
		for _, o := range s.outs {
			ctx.Send(o, msg)
		}
	}
}

func (s *Scrape) startScrape() {
	ctxBack := context.Background()
	ctxBack, s.cancelFunc = context.WithCancel(ctxBack)
	s.scraper = scrape.NewManager(nil, s.log, s)
	s.sd = discovery.NewManager(ctxBack, s.log, discovery.Name("autoscraper/"+s.name))
	cfgs := s.convertToScrapeConfig()

	scrapeConfigs := make([]*promconfig.ScrapeConfig, 0, 0)
	sdConfigs := make(map[string]discovery.Configs, 0)

	for _, job := range cfgs {
		sdConfigs[job.JobName] = job.ServiceDiscoveryConfigs
		scrapeConfigs = append(scrapeConfigs, job)
	}
	err := s.sd.ApplyConfig(sdConfigs)
	if err != nil {
		level.Error(s.log).Log("error", err, "msg", "error applying scraper service discovery config")
	}
	err = s.scraper.ApplyConfig(&promconfig.Config{
		GlobalConfig:  promconfig.DefaultGlobalConfig,
		ScrapeConfigs: cfgs,
		TracingConfig: promconfig.TracingConfig{},
	})
	if err != nil {
		level.Error(s.log).Log("error", err, "msg", "error applying scraper config")
	}
	go s.sd.Run()
	go s.scraper.Run(s.sd.SyncCh())
}

func (s *Scrape) convertToScrapeConfig() []*promconfig.ScrapeConfig {
	hs, _ := os.Hostname()
	scrapes := make([]*promconfig.ScrapeConfig, 0)
	for _, sc := range s.cfg.ScrapeConfigs {

		scrapes = append(scrapes, &promconfig.ScrapeConfig{
			JobName:                 sc.JobName,
			ScrapeInterval:          model.Duration(sc.ScrapeInterval),
			ScrapeTimeout:           model.Duration(sc.ScrapeTimeout),
			Scheme:                  "http",
			BodySizeLimit:           0,
			SampleLimit:             0,
			TargetLimit:             0,
			LabelLimit:              0,
			LabelNameLengthLimit:    0,
			LabelValueLengthLimit:   0,
			ServiceDiscoveryConfigs: convertToDiscoveryConfigs(sc.Targets, hs),
			HTTPClientConfig:        cmnconfig.DefaultHTTPClientConfig,
			RelabelConfigs:          nil,
			MetricRelabelConfigs:    nil,
			MetricsPath:             "/metrics",
		})
	}
	return scrapes
}

func convertToDiscoveryConfigs(target []string, hostname string) discovery.Configs {
	cfgs := make(discovery.Configs, 0)
	for _, t := range target {
		// A blank host somehow works, but it then requires a sever name to be set under tls.
		newHost := t
		if newHost == "" {
			newHost = "127.0.0.1"
		}
		labels := model.LabelSet{}
		labels[model.LabelName("agent_hostname")] = model.LabelValue(hostname)
		cfgs = append(cfgs, discovery.StaticConfig{{
			Targets: []model.LabelSet{{model.AddressLabel: model.LabelValue(t)}},
			Labels:  labels,
		}})

	}
	return cfgs
}

func (s *Scrape) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	m := exchange.NewMetricFromPromMetric(t, v, l)
	s.cachedMetrics = append(s.cachedMetrics, m)
	return 0, nil
}

func (s *Scrape) Commit() error {
	if len(s.cachedMetrics) == 0 {
		return nil
	}
	newMetrics := make([]exchange.Metric, len(s.cachedMetrics))
	copy(newMetrics, s.cachedMetrics)
	s.cachedMetrics = make([]exchange.Metric, 0)
	s.rootContext.Send(s.self, newMetrics)
	return nil
}

func (s *Scrape) Rollback() error {
	return nil
}

func (s *Scrape) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	return 0, nil
}

func (s *Scrape) Appender(ctx context.Context) storage.Appender {
	return s
}
