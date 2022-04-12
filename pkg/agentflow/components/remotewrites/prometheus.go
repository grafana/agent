package remotewrites

import (
	"context"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/agentflow/config"
	"github.com/grafana/agent/pkg/agentflow/types"
	"github.com/grafana/agent/pkg/agentflow/types/actorstate"
	"github.com/grafana/agent/pkg/agentflow/types/exchange"
	"github.com/grafana/agent/pkg/metrics/wal"
	"github.com/prometheus/client_golang/prometheus"
	cmnconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	prconfig "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/timestamp"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
	"net/url"

	"sync"
	"time"
)

type Prometheus struct {
	name    string
	self    *actor.PID
	storage storage.Storage
	wal     *wal.Storage
	log     log.Logger
}

func (f *Prometheus) AllowableInputs() []actorstate.InOutType {
	return []actorstate.InOutType{actorstate.Metrics}
}

func (f *Prometheus) Output() actorstate.InOutType {
	return actorstate.None
}

func NewPrometheus(name string, global *types.Global, cfg *config.PrometheusRemoteWrite) (actorstate.FlowActor, error) {
	r := prometheus.NewRegistry()
	walSt, err := wal.NewStorage(global.Log, r, cfg.WalDir)
	if err != nil {
		return nil, err
	}
	readyScrapeManager := &readyScrapeManager{}
	rwUrl, _ := url.Parse(cfg.URL)
	cmUrl := cmnconfig.URL{}
	cmUrl.URL = rwUrl
	httpCfg := cmnconfig.DefaultHTTPClientConfig
	secr := cmnconfig.Secret(cfg.Password)
	httpCfg.BasicAuth = &cmnconfig.BasicAuth{
		Username: cfg.Username,
		Password: secr,
	}
	st := remote.NewStorage(global.Log, r, walSt.StartTime, cfg.WalDir, 1*time.Minute, readyScrapeManager)
	err = st.ApplyConfig(&prconfig.Config{
		RemoteWriteConfigs: []*prconfig.RemoteWriteConfig{
			{
				Name:             "default",
				URL:              &cmUrl,
				HTTPClientConfig: httpCfg,
				QueueConfig:      prconfig.DefaultQueueConfig,
				MetadataConfig:   prconfig.DefaultMetadataConfig,
				RemoteTimeout:    model.Duration(30 * time.Second),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	stor := storage.NewFanout(global.Log, walSt, st)

	return &Prometheus{
		name:    name,
		storage: stor,
		log:     global.Log,
	}, nil
}

func (f *Prometheus) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case actorstate.Start:
		f.self = c.Self()
	case []exchange.Metric:
		appender := f.storage.Appender(context.Background())
		for _, m := range msg {
			promLbls := labels.FromMap(m.Labels())
			_, err := appender.Append(0, promLbls, timestamp.FromTime(m.Timestamp()), m.Value())
			if err != nil {
				level.Error(f.log).Log("msg", "error while writing to appender", "error", err)
			}
		}
		err := appender.Commit()
		if err != nil {
			level.Error(f.log).Log("msg", "error while committing to appender", "error", err)
		}

	}
}

func (f *Prometheus) Name() string {
	return f.name
}

func (f Prometheus) PID() *actor.PID {
	return f.PID()
}

// readyScrapeManager allows a scrape manager to be retrieved. Even if it's set at a later point in time.
type readyScrapeManager struct {
	mtx sync.RWMutex
	m   *scrape.Manager
}

// Set the scrape manager.
func (rm *readyScrapeManager) Set(m *scrape.Manager) {
	rm.mtx.Lock()
	defer rm.mtx.Unlock()

	rm.m = m
}

// Get the scrape manager. If is not ready, return an error.
func (rm *readyScrapeManager) Get() (*scrape.Manager, error) {
	rm.mtx.RLock()
	defer rm.mtx.RUnlock()

	if rm.m != nil {
		return rm.m, nil
	}

	return nil, fmt.Errorf("not ready")
}
