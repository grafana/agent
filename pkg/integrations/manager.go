package integrations

import (
	"context"
	"errors"
	"fmt"
	config_util "github.com/prometheus/common/config"
	"path"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/pkg/relabel"
)

var (
	integrationAbnormalExits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agent_prometheus_integration_abnormal_exits_total",
		Help: "Total number of times an agent integration exited unexpectedly, causing it to be restarted.",
	}, []string{"integration_name"})
)

var (
	DefaultManagerConfig = ManagerConfig{
		ScrapeIntegrations:        true,
		IntegrationRestartBackoff: 5 * time.Second,
		UseHostnameLabel:          true,
		ReplaceInstanceLabel:      true,
	}
)

// ManagerConfig holds the configuration for all integrations.
type ManagerConfig struct {
	// Whether the Integration subsystem should be enabled.
	Enabled bool `yaml:"-"`

	// When true, scrapes metrics from integrations.
	ScrapeIntegrations bool `yaml:"scrape_integrations,omitempty"`
	// When true, replaces the instance label with the agent hostname.
	ReplaceInstanceLabel bool `yaml:"replace_instance_label,omitempty"`

	// DEPRECATED. When true, adds an agent_hostname label to all samples from integrations.
	// ReplaceInstanceLabel should be used instead.
	UseHostnameLabel bool `yaml:"use_hostname_label,omitempty"`

	// The integration configs is merged with the manager config struct so we
	// don't want to export it here; we'll manually unmarshal it in UnmarshalYAML.
	Integrations Configs `yaml:"-"`

	// Extra labels to add for all integration samples
	Labels model.LabelSet `yaml:"labels,omitempty"`

	// Prometheus RW configs to use for all integrations.
	PrometheusRemoteWrite []*instance.RemoteWriteConfig `yaml:"prometheus_remote_write,omitempty"`

	IntegrationRestartBackoff time.Duration `yaml:"integration_restart_backoff,omitempty"`

	// ListenPort tells the integration Manager which port the Agent is
	// listening on for generating Prometheus instance configs.
	ListenPort *int `yaml:"-"`

	// ListenHost tells the integration Manager which port the Agent is
	// listening on for generating Prometheus instance configs
	ListenHost *string `yaml:"-"`

	// Drives using HTTPS to read the integrations

	ClientCert string `yaml:"client_cert_file"`
	ClientKey  string `yaml:"client_key_file"`
	ServerName string `yaml:"server_name"`

	// If this is RequireAndVerifyClientCert then the ClientCert/Key/ServerName fields are required
	// comes from the http_tls_config in the server settings
	ClientAuthType string

	ClientCA string
}

// MarshalYAML implements yaml.Marshaler for ManagerConfig.
func (c ManagerConfig) MarshalYAML() (interface{}, error) {
	return MarshalYAML(c)
}

// UnmarshalYAML implements yaml.Unmarshaler for ManagerConfig.
func (c *ManagerConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultManagerConfig

	// If the ManagerConfig is unmarshaled, it's present in the config and should be
	// enabled.
	c.Enabled = true

	return UnmarshalYAML(c, unmarshal)
}

// DefaultRelabelConfigs returns the set of relabel configs that should be
// prepended to all RelabelConfigs for an integration.
func (c *ManagerConfig) DefaultRelabelConfigs() ([]*relabel.Config, error) {
	var cfgs []*relabel.Config

	if c.ReplaceInstanceLabel {
		hostname, err := instance.Hostname()
		if err != nil {
			return nil, err
		}

		replacement := fmt.Sprintf("%s:%d", hostname, *c.ListenPort)

		cfgs = append(cfgs, &relabel.Config{
			SourceLabels: model.LabelNames{model.AddressLabel},
			Action:       relabel.Replace,
			Separator:    ";",
			Regex:        relabel.MustNewRegexp("(.*)"),
			Replacement:  replacement,
			TargetLabel:  model.InstanceLabel,
		})
	}

	return cfgs, nil
}

// Manager manages a set of integrations and runs them.
type Manager struct {
	c                     ManagerConfig
	logger                log.Logger
	integrations          map[Config]Integration
	hostname              string
	defaultRelabelConfigs []*relabel.Config

	im     instance.Manager
	cancel context.CancelFunc
	done   chan bool
}

// NewManager creates a new integrations manager. NewManager must be given an
// InstanceManager which is responsible for accepting instance configs to
// scrape and send metrics from running integrations.
func NewManager(c ManagerConfig, logger log.Logger, im instance.Manager) (*Manager, error) {
	integrations := make(map[Config]Integration)

	for _, integrationCfg := range c.Integrations {
		if integrationCfg.CommonConfig().Enabled {
			l := log.With(logger, "integration", integrationCfg.Name())
			i, err := integrationCfg.NewIntegration(l)
			if err != nil {
				return nil, err
			}
			integrations[integrationCfg] = i
		}
	}

	return newManager(c, logger, im, integrations)
}

func newManager(c ManagerConfig, logger log.Logger, im instance.Manager, integrations map[Config]Integration) (*Manager, error) {
	defaultRelabels, err := c.DefaultRelabelConfigs()
	if err != nil {
		return nil, fmt.Errorf("cannot get default relabel configs: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		c:                     c,
		logger:                logger,
		integrations:          integrations,
		defaultRelabelConfigs: defaultRelabels,
		im:                    im,
		cancel:                cancel,
		done:                  make(chan bool),
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

	for cfg, i := range m.integrations {
		go func(cfg Config, i Integration) {
			m.runIntegration(ctx, cfg, i)
			wg.Done()
		}(cfg, i)
	}

	wg.Wait()
	close(m.done)
}

func (m *Manager) runIntegration(ctx context.Context, cfg Config, i Integration) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v", r)
			level.Error(m.logger).Log("msg", "integration has panicked. THIS IS A BUG!", "err", err, "integration", cfg.Name())
		}
	}()

	shouldCollect := m.c.ScrapeIntegrations
	if common := cfg.CommonConfig(); common.ScrapeIntegration != nil {
		shouldCollect = *common.ScrapeIntegration
	}
	if shouldCollect {
		// Apply the config so an instance is launched to scrape our integration.
		instanceConfig, err := m.instanceConfigForIntegration(cfg, i)
		if err != nil {
			level.Error(m.logger).Log("msg", "failed to create config integration. integration will not run", "err", err, "integration", cfg.Name())
			return
		}

		if err := m.im.ApplyConfig(instanceConfig); err != nil {
			level.Error(m.logger).Log("msg", "failed to apply integration. integration will not run", "err", err, "integration", cfg.Name())
			return
		}
	}

	for {
		err := i.Run(ctx)
		if err != nil && err != context.Canceled {
			integrationAbnormalExits.WithLabelValues(cfg.Name()).Inc()
			level.Error(m.logger).Log("msg", "integration stopped abnormally, restarting after backoff", "err", err, "integration", cfg.Name(), "backoff", m.c.IntegrationRestartBackoff)
			time.Sleep(m.c.IntegrationRestartBackoff)
		} else {
			level.Info(m.logger).Log("msg", "stopped integration", "integration", cfg.Name())
			break
		}
	}
}

func (m *Manager) instanceConfigForIntegration(cfg Config, i Integration) (instance.Config, error) {
	prometheusName := fmt.Sprintf("integration/%s", cfg.Name())

	common := cfg.CommonConfig()
	relabelConfigs := append(m.defaultRelabelConfigs, common.RelabelConfigs...)

	var scrapeConfigs []*config.ScrapeConfig
	schema := "http"

	// If there is a TLS cert path then assume we need to use client certificates
	httpClientConfig := config_util.HTTPClientConfig{}
	if m.c.ClientAuthType != "" {
		// If the AuthType is to require a client cert/servername then one needs to be defined
		certRequired := m.c.ClientAuthType == "RequireAndVerifyClientCert" || m.c.ClientAuthType == "RequireAnyClientCert"
		certEmpty := m.c.ClientKey == "" || m.c.ClientCert == "" || m.c.ServerName == ""
		if certRequired && certEmpty {
			err := errors.New("client key or client cert or servername is not specific but a Client TLS Cert is required")
			level.Error(m.logger).Log("tls", "Client TLS", "err", err)
			return instance.Config{}, err
		}
		schema = "https"
		httpClientConfig.TLSConfig = config_util.TLSConfig{
			CAFile:     m.c.ClientCA,
			CertFile:   m.c.ClientCert,
			KeyFile:    m.c.ClientKey,
			ServerName: m.c.ServerName,
		}
	}
	for _, isc := range i.ScrapeConfigs() {
		sc := &config.ScrapeConfig{
			JobName:                 fmt.Sprintf("integrations/%s", isc.JobName),
			MetricsPath:             path.Join("/integrations", cfg.Name(), isc.MetricsPath),
			Scheme:                  schema,
			HonorLabels:             false,
			HonorTimestamps:         true,
			ScrapeInterval:          model.Duration(common.ScrapeInterval),
			ScrapeTimeout:           model.Duration(common.ScrapeTimeout),
			ServiceDiscoveryConfigs: m.scrapeServiceDiscovery(),
			RelabelConfigs:          relabelConfigs,
			MetricRelabelConfigs:    common.MetricRelabelConfigs,
			HTTPClientConfig:        httpClientConfig,
		}

		scrapeConfigs = append(scrapeConfigs, sc)
	}

	instanceCfg := instance.DefaultConfig
	instanceCfg.Name = prometheusName
	instanceCfg.ScrapeConfigs = scrapeConfigs
	instanceCfg.RemoteWrite = m.c.PrometheusRemoteWrite
	if common.WALTruncateFrequency > 0 {
		instanceCfg.WALTruncateFrequency = common.WALTruncateFrequency
	}
	return instanceCfg, nil
}

func (m *Manager) scrapeServiceDiscovery() discovery.Configs {

	localAddr := fmt.Sprintf("%s:%d", *m.c.ListenHost, *m.c.ListenPort)
	labels := model.LabelSet{}
	if m.c.UseHostnameLabel {
		labels[model.LabelName("agent_hostname")] = model.LabelValue(m.hostname)
	}
	for k, v := range m.c.Labels {
		labels[k] = v
	}

	return discovery.Configs{
		discovery.StaticConfig{{
			Targets: []model.LabelSet{{model.AddressLabel: model.LabelValue(localAddr)}},
			Labels:  labels,
		}},
	}
}

func (m *Manager) WireAPI(r *mux.Router) error {
	for c, i := range m.integrations {
		integrationsRoot := fmt.Sprintf("/integrations/%s", c.Name())
		subRouter := r.PathPrefix(integrationsRoot).Subrouter()

		err := i.RegisterRoutes(subRouter)
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop stops the manager and all of its integrations.
func (m *Manager) Stop() {
	m.cancel()
	<-m.done
}
