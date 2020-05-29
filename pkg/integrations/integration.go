package integrations

import (
	"context"
	"flag"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/agent"
	"github.com/grafana/agent/pkg/prom/ha"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	sd_config "github.com/prometheus/prometheus/discovery/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

var (
	integrationAbnormalExits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agent_prometheus_integration_abnormal_exits_total",
		Help: "Total number of times an agent integration exited unexpectedly, causing it to be restarted.",
	}, []string{"integration_name"})
)

// Config holds the configuration for all integrations.
type Config struct {
	Agent agent.Config `yaml:"agent"`

	// Prometheus RW configs to use for all integrations.
	PrometheusRemoteWrite []*config.RemoteWriteConfig `yaml:"prometheus_remote_write,omitempty"`

	// ListenPort tells the integration Manager which port the Agent is
	// listening on for generating Prometheus instance configs.
	ListenPort *int `yaml:"-"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Agent.RegisterFlagsWithPrefix("integrations.", f)
}

type Integration interface {
	// Name returns the name of the integration. Each registered integration must
	// have a unique name.
	Name() string

	// RegisterRoutes should register any HTTP handlers used for the integration.
	//
	// The router provided to RegisterRoutes is a subrouter for the path
	// /integrations/<integration name>. All routes should register to the
	// relative root path and will be automatically combined to the subroute. For
	// example, if a metric "database" registers a /metrics endpoint, it will
	// be exposed as /integrations/database/metrics.
	RegisterRoutes(r *mux.Router) error

	// MetricsEndpoints should return any endpoints (as defined by RegisterRoutes)
	// that expose Prometheus/OpenMetrics metrics.
	MetricsEndpoints() []string

	// Run should start the integration and do any required tasks. Run should *not*
	// exit until context is canceled. If an integration doesn't need to do anything,
	// it should simply wait for ctx to be canceled.
	Run(ctx context.Context) error
}

// Manager manages a set of integrations and runs them.
type Manager struct {
	c            Config
	logger       log.Logger
	integrations []Integration

	im     ha.InstanceManager
	cancel context.CancelFunc
	done   chan bool
}

// NewManager creates a new integrations manager. NewManager must be given an
// InstanceManager which is responsible for accepting instance configs to
// scrape and send metrics from running integrations.
func NewManager(c Config, logger log.Logger, im ha.InstanceManager) *Manager {
	var integrations []Integration
	if c.Agent.Enabled {
		integrations = append(integrations, agent.New())
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		c:            c,
		logger:       logger,
		integrations: integrations,
		im:           im,
		cancel:       cancel,
		done:         make(chan bool),
	}

	go m.run(ctx)
	return m
}

func (m *Manager) run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(len(m.integrations))

	for _, i := range m.integrations {
		go func(i Integration) {
			m.runIntegration(ctx, i)
			wg.Done()
		}(i)
	}

	wg.Wait()
	close(m.done)
}

func (m *Manager) runIntegration(ctx context.Context, i Integration) {
	// Apply the config so an instance is launched to scrape our integration.
	instanceConfig := m.instanceConfigForIntegration(i)
	if err := m.im.ApplyConfig(instanceConfig); err != nil {
		level.Error(m.logger).Log("msg", "failed to apply integration. integration will not run. THIS IS A BUG!", "err", err)
		return
	}

	for {
		err := i.Run(ctx)
		if err != nil && err != context.Canceled {
			integrationAbnormalExits.WithLabelValues(i.Name()).Inc()
			level.Error(m.logger).Log("msg", "integration stopped abnormally, restarting after 5s", "err", err, "integration", i.Name())
			time.Sleep(5 * time.Second)
		} else {
			level.Info(m.logger).Log("msg", "stopped integration", "integration", i.Name())
		}
	}
}

func (m *Manager) instanceConfigForIntegration(i Integration) instance.Config {
	prometheusName := fmt.Sprintf("integration/%s", i.Name())

	var scrapeConfigs []*config.ScrapeConfig
	for idx, endpoint := range i.MetricsEndpoints() {
		metricsPath := path.Join("/integrations", i.Name(), endpoint)
		sc := &config.ScrapeConfig{
			JobName:         fmt.Sprintf("integration/%s/%d", i.Name(), idx),
			MetricsPath:     metricsPath,
			Scheme:          "http",
			HonorLabels:     false,
			HonorTimestamps: true,
			ServiceDiscoveryConfig: sd_config.ServiceDiscoveryConfig{
				StaticConfigs: []*targetgroup.Group{{
					Targets: []model.LabelSet{{
						model.AddressLabel: m.scrapeAddress(),
					}},
					Labels: model.LabelSet{
						"integration": model.LabelValue(i.Name()),
					},
				}},
			},
		}

		scrapeConfigs = append(scrapeConfigs, sc)
	}

	return instance.Config{
		Name:          prometheusName,
		ScrapeConfigs: scrapeConfigs,
		RemoteWrite:   m.c.PrometheusRemoteWrite,
	}
}

func (m *Manager) scrapeAddress() model.LabelValue {
	return model.LabelValue(
		fmt.Sprintf("127.0.0.1:%d", *m.c.ListenPort),
	)
}

func (m *Manager) WireAPI(r *mux.Router) error {
	for _, i := range m.integrations {
		integrationsRoot := fmt.Sprintf("/integrations/%s", i.Name())
		subRouter := r.PathPrefix(integrationsRoot).Subrouter()

		err := i.RegisterRoutes(subRouter)
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop stops the maanger and all of its integrations.
func (m *Manager) Stop() {
	m.cancel()
	<-m.done
}
