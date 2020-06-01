package integrations

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/agent"
	integrationCfg "github.com/grafana/agent/pkg/integrations/config"
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
	// When true, adds an agent_hostname label to all samples from integrations.
	UseHostnameLabel bool `yaml:"use_hostname_label"`

	Agent agent.Config `yaml:"agent"`

	// Extra labels to add for all integration samples
	Labels model.LabelSet `yaml:"labels"`

	// Prometheus RW configs to use for all integrations.
	PrometheusRemoteWrite []*config.RemoteWriteConfig `yaml:"prometheus_remote_write,omitempty"`

	// ListenPort tells the integration Manager which port the Agent is
	// listening on for generating Prometheus instance configs.
	ListenPort *int `yaml:"-"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.BoolVar(&c.UseHostnameLabel, "integrations.use-hostname-label", true, "When true, adds an agent_hostname label to all samples from integrations.")

	c.Agent.RegisterFlagsWithPrefix("integrations.", f)
}

type Integration interface {
	// Name returns the name of the integration. Each registered integration must
	// have a unique name.
	Name() string

	// CommonConfig returns the set of common configuration values present across
	// all integrations.
	CommonConfig() integrationCfg.Common

	// RegisterRoutes should register any HTTP handlers used for the integration.
	//
	// The router provided to RegisterRoutes is a subrouter for the path
	// /integrations/<integration name>. All routes should register to the
	// relative root path and will be automatically combined to the subroute. For
	// example, if a metric "database" registers a /metrics endpoint, it will
	// be exposed as /integrations/database/metrics.
	RegisterRoutes(r *mux.Router) error

	// ScrapeConfigs should return a set of integration scrape configs that inform
	// the integration how samples should be collected.
	ScrapeConfigs() []integrationCfg.ScrapeConfig

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
	hostname     string

	im     ha.InstanceManager
	cancel context.CancelFunc
	done   chan bool
}

// NewManager creates a new integrations manager. NewManager must be given an
// InstanceManager which is responsible for accepting instance configs to
// scrape and send metrics from running integrations.
func NewManager(c Config, logger log.Logger, im ha.InstanceManager) (*Manager, error) {
	var integrations []Integration
	if c.Agent.Enabled {
		integrations = append(integrations, agent.New(c.Agent))
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

	if c.UseHostnameLabel {
		var err error
		m.hostname, err = instance.Hostname()
		if err != nil {
			return nil, err
		}
	}

	go m.run(ctx)
	return m, nil
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

	common := i.CommonConfig()

	var scrapeConfigs []*config.ScrapeConfig
	for _, cfg := range i.ScrapeConfigs() {
		sc := &config.ScrapeConfig{
			JobName:                fmt.Sprintf("integrations/%s", cfg.JobName),
			MetricsPath:            filepath.Join("/integrations", i.Name(), cfg.MetricsPath),
			Scheme:                 "http",
			HonorLabels:            false,
			HonorTimestamps:        true,
			ScrapeInterval:         model.Duration(common.ScrapeInterval),
			ScrapeTimeout:          model.Duration(common.ScrapeTimeout),
			ServiceDiscoveryConfig: m.scrapeServiceDiscovery(),
		}

		scrapeConfigs = append(scrapeConfigs, sc)
	}

	return instance.Config{
		Name:          prometheusName,
		ScrapeConfigs: scrapeConfigs,
		RemoteWrite:   m.c.PrometheusRemoteWrite,
	}
}

func (m *Manager) scrapeServiceDiscovery() sd_config.ServiceDiscoveryConfig {
	localAddr := fmt.Sprintf("127.0.0.1:%d", *m.c.ListenPort)

	labels := model.LabelSet{}
	if m.c.UseHostnameLabel {
		labels[model.LabelName("agent_hostname")] = model.LabelValue(m.hostname)
	}
	for k, v := range m.c.Labels {
		labels[k] = v
	}

	return sd_config.ServiceDiscoveryConfig{
		StaticConfigs: []*targetgroup.Group{{
			Targets: []model.LabelSet{{model.AddressLabel: model.LabelValue(localAddr)}},
			Labels:  labels,
		}},
	}
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
