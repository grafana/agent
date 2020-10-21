package integrations

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/agent"
	integrationCfg "github.com/grafana/agent/pkg/integrations/config"
	"github.com/grafana/agent/pkg/integrations/mysqld_exporter"
	"github.com/grafana/agent/pkg/integrations/node_exporter"
	"github.com/grafana/agent/pkg/integrations/process_exporter"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	sd_config "github.com/prometheus/prometheus/discovery/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/pkg/relabel"
)

var (
	integrationAbnormalExits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agent_prometheus_integration_abnormal_exits_total",
		Help: "Total number of times an agent integration exited unexpectedly, causing it to be restarted.",
	}, []string{"integration_name"})
)

var (
	DefaultConfig = Config{
		ScrapeIntegrations:        true,
		IntegrationRestartBackoff: 5 * time.Second,
		UseHostnameLabel:          true,
		ReplaceInstanceLabel:      true,
	}
)

// Config holds the configuration for all integrations.
type Config struct {
	// Whether the Integration subsystem should be enabled.
	Enabled bool `yaml:"-"`

	// When true, scrapes metrics from integrations.
	ScrapeIntegrations bool `yaml:"scrape_integrations"`
	// When true, replaces the instance label with the agent hostname.
	ReplaceInstanceLabel bool `yaml:"replace_instance_label"`

	// DEPRECATED. When true, adds an agent_hostname label to all samples from integrations.
	// ReplaceInstanceLabel should be used instead.
	UseHostnameLabel bool `yaml:"use_hostname_label"`

	Agent           agent.Config            `yaml:"agent"`
	NodeExporter    node_exporter.Config    `yaml:"node_exporter"`
	ProcessExporter process_exporter.Config `yaml:"process_exporter"`
	MysqldExporter  mysqld_exporter.Config  `yaml:"mysqld_exporter"`

	// Extra labels to add for all integration samples
	Labels model.LabelSet `yaml:"labels"`

	// Prometheus RW configs to use for all integrations.
	PrometheusRemoteWrite []*config.RemoteWriteConfig `yaml:"prometheus_remote_write,omitempty"`

	IntegrationRestartBackoff time.Duration `yaml:"integration_restart_backoff,omitempty"`

	// ListenPort tells the integration Manager which port the Agent is
	// listening on for generating Prometheus instance configs.
	ListenPort *int `yaml:"-"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	// If the Config is unmarshaled, it's present in the config and should be
	// enabled.
	c.Enabled = true

	type plain Config
	return unmarshal((*plain)(c))
}

func (c *Config) ApplyDefaults() error {
	if c.ReplaceInstanceLabel {
		hostname, err := instance.Hostname()
		if err != nil {
			return err
		}

		replacement := fmt.Sprintf("%s:%d", hostname, *c.ListenPort)

		for _, rwCfg := range c.PrometheusRemoteWrite {
			rwCfg.WriteRelabelConfigs = append([]*relabel.Config{
				{
					SourceLabels: model.LabelNames{model.InstanceLabel},
					Action:       relabel.Replace,
					Separator:    ";",
					Regex:        relabel.MustNewRegexp("(.*)"),
					Replacement:  replacement,
					TargetLabel:  model.InstanceLabel,
				},
			}, rwCfg.WriteRelabelConfigs...)
		}
	}

	return nil
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

	im     instance.Manager
	cancel context.CancelFunc
	done   chan bool
}

// NewManager creates a new integrations manager. NewManager must be given an
// InstanceManager which is responsible for accepting instance configs to
// scrape and send metrics from running integrations.
func NewManager(c Config, logger log.Logger, im instance.Manager) (*Manager, error) {
	var integrations []Integration

	if c.Agent.Enabled {
		integrations = append(integrations, agent.New(c.Agent))
	}
	if c.NodeExporter.Enabled {
		l := log.With(logger, "integration", "node_exporter")
		i, err := node_exporter.New(l, c.NodeExporter)
		if err != nil {
			return nil, err
		}
		integrations = append(integrations, i)
	}
	if c.ProcessExporter.Enabled {
		l := log.With(logger, "integration", "process_exporter")
		i, err := process_exporter.New(l, c.ProcessExporter)
		if err != nil {
			return nil, err
		}
		integrations = append(integrations, i)
	}
	if c.MysqldExporter.Enabled {
		l := log.With(logger, "integration", "mysqld_exporter")
		i, err := mysqld_exporter.New(l, c.MysqldExporter)
		if err != nil {
			return nil, err
		}
		integrations = append(integrations, i)
	}

	return newManager(c, logger, im, integrations)
}

func newManager(c Config, logger log.Logger, im instance.Manager, integrations []Integration) (*Manager, error) {
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
	shouldCollect := m.c.ScrapeIntegrations
	if common := i.CommonConfig(); common.ScrapeIntegration != nil {
		shouldCollect = *common.ScrapeIntegration
	}
	if shouldCollect {
		// Apply the config so an instance is launched to scrape our integration.
		instanceConfig := m.instanceConfigForIntegration(i)
		if err := m.im.ApplyConfig(instanceConfig); err != nil {
			level.Error(m.logger).Log("msg", "failed to apply integration. integration will not run. THIS IS A BUG!", "err", err, "integration", i.Name())
			return
		}
	}

	for {
		err := i.Run(ctx)
		if err != nil && err != context.Canceled {
			integrationAbnormalExits.WithLabelValues(i.Name()).Inc()
			level.Error(m.logger).Log("msg", "integration stopped abnormally, restarting after backoff", "err", err, "integration", i.Name(), "backoff", m.c.IntegrationRestartBackoff)
			time.Sleep(m.c.IntegrationRestartBackoff)
		} else {
			level.Info(m.logger).Log("msg", "stopped integration", "integration", i.Name())
			break
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
			MetricsPath:            path.Join("/integrations", i.Name(), cfg.MetricsPath),
			Scheme:                 "http",
			HonorLabels:            false,
			HonorTimestamps:        true,
			ScrapeInterval:         model.Duration(common.ScrapeInterval),
			ScrapeTimeout:          model.Duration(common.ScrapeTimeout),
			ServiceDiscoveryConfig: m.scrapeServiceDiscovery(),
			RelabelConfigs:         common.RelabelConfigs,
			MetricRelabelConfigs:   common.MetricRelabelConfigs,
		}

		scrapeConfigs = append(scrapeConfigs, sc)
	}

	instanceCfg := instance.DefaultConfig
	instanceCfg.Name = prometheusName
	instanceCfg.ScrapeConfigs = scrapeConfigs
	instanceCfg.RemoteWrite = m.c.PrometheusRemoteWrite
	return instanceCfg
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
