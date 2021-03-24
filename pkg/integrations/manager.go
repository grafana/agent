package integrations

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	config_util "github.com/prometheus/common/config"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/grafana/agent/pkg/prom/instance/configstore"
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

// DefaultManagerConfig holds the default settings for integrations.
var DefaultManagerConfig = ManagerConfig{
	ScrapeIntegrations:        true,
	IntegrationRestartBackoff: 5 * time.Second,
	UseHostnameLabel:          true,
	ReplaceInstanceLabel:      true,
}

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

	TLSConfig config_util.TLSConfig `yaml:"http_tls_config"`

	// This is set to true if the Server TLSConfig Cert and Key path are set
	ServerUsingTLS bool `yaml:"-"`
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
	hostname              string
	defaultRelabelConfigs []*relabel.Config

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	im        instance.Manager
	done      chan bool
	validator configstore.Validator

	integrationsMut sync.RWMutex
	integrations    map[string]*integrationProcess
}

// NewManager creates a new integrations manager. NewManager must be given an
// InstanceManager which is responsible for accepting instance configs to
// scrape and send metrics from running integrations.
func NewManager(c ManagerConfig, logger log.Logger, im instance.Manager, validate configstore.Validator) (*Manager, error) {
	defaultRelabels, err := c.DefaultRelabelConfigs()
	if err != nil {
		return nil, fmt.Errorf("cannot get default relabel configs: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		c:                     c,
		logger:                logger,
		defaultRelabelConfigs: defaultRelabels,
		im:                    im,
		done:                  make(chan bool),
		validator:             validate,

		ctx:    ctx,
		cancel: cancel,
	}

	if c.UseHostnameLabel {
		var err error
		m.hostname, err = instance.Hostname()
		if err != nil {
			return nil, err
		}
	}

	if err := m.ApplyConfig(c); err != nil {
		return nil, fmt.Errorf("failed applying config: %w", err)
	}
	return m, nil
}

// ApplyConfig updates the configuration of the integrations subsystem.
func (m *Manager) ApplyConfig(c ManagerConfig) error {
	m.integrationsMut.Lock()
	defer m.integrationsMut.Unlock()

	if cmp.Equal(m.c, c) {
		return nil
	}

	// Iterate over our integrations. New or changed integrations will be
	// started, with their existing counterparts being shut down.
	for _, ic := range c.Integrations {
		// Key is used to identify the instance of this integration within the
		// instance manager and within our set of running integrations.
		key := instanceConfigKey(ic)

		if p, exist := m.integrations[key]; exist {
			// If the existing config hasn't changed, we don't want to re-start it.
			if cmp.Equal(p.cfg, ic) {
				continue
			}

			// ...Otherwise, we're going to replace the instance, so stop it before
			// creating a new one.
			p.stop()
		}

		l := log.With(m.logger, "integration", ic.Name())
		i, err := ic.NewIntegration(l)
		if err != nil {
			return fmt.Errorf("error initializing integration %q: %w", ic.Name(), err)
		}

		// Create, start, and register the new integration.
		ctx, cancel := context.WithCancel(m.ctx)
		p := &integrationProcess{
			log:  l,
			ctx:  ctx,
			stop: cancel,
			cfg:  ic,
			i:    i,
			m:    m,
		}
		go p.Run()
		m.integrations[key] = p

		// Configure the instance manager for the integration. This may include
		// deleting an existing config, which can happen when the previous config
		// was set to scrape but that config went away.
		shouldCollect := m.c.ScrapeIntegrations
		if common := ic.CommonConfig(); common.ScrapeIntegration != nil {
			shouldCollect = *common.ScrapeIntegration
		}
		if shouldCollect {
			instanceConfig := m.instanceConfigForIntegration(ic, i)
			if err := m.validator(&instanceConfig); err != nil {
				level.Error(p.log).Log("msg", "failed to apply integration. integration will not be scraped", "err", err, "integration", p.cfg.Name())
				continue
			}

			if err := m.im.ApplyConfig(instanceConfig); err != nil {
				level.Error(p.log).Log("msg", "failed to apply integration. integration will not be scraped", "err", err, "integration", p.cfg.Name())
				continue
			}
		} else {
			_ = m.im.DeleteConfig(key)
		}
	}

	// Remove configurations that have been removed and stop scraping them.
	for key, process := range m.integrations {
		foundConfig := false
		for _, ic := range c.Integrations {
			if instanceConfigKey(ic) == key {
				foundConfig = true
				break
			}
		}
		if foundConfig {
			continue
		}

		process.stop()
		_ = m.im.DeleteConfig(key)
		delete(m.integrations, key)
	}

	return nil
}

// integrationProcess is a running integration.
type integrationProcess struct {
	log  log.Logger
	ctx  context.Context
	stop context.CancelFunc
	cfg  Config
	i    Integration
	m    *Manager

	wg *sync.WaitGroup
}

// Run runs the integration until the process is canceled.
func (p *integrationProcess) Run() {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v", r)
			level.Error(p.log).Log("msg", "integration has panicked. THIS IS A BUG!", "err", err, "integration", p.cfg.Name())
		}
	}()

	p.wg.Add(1)
	defer p.wg.Done()

	for {
		err := p.i.Run(p.ctx)
		if err != nil && err != context.Canceled {
			integrationAbnormalExits.WithLabelValues(p.cfg.Name()).Inc()
			level.Error(p.log).Log("msg", "integration stopped abnormally, restarting after backoff", "err", err, "integration", p.cfg.Name(), "backoff", p.m.c.IntegrationRestartBackoff)
			time.Sleep(p.m.c.IntegrationRestartBackoff)
		} else {
			level.Info(p.log).Log("msg", "stopped integration", "integration", p.cfg.Name())
			break
		}
	}
}

func (m *Manager) instanceConfigForIntegration(cfg Config, i Integration) instance.Config {
	common := cfg.CommonConfig()
	relabelConfigs := append(m.defaultRelabelConfigs, common.RelabelConfigs...)

	schema := "http"
	// Check for HTTPS support
	var httpClientConfig config_util.HTTPClientConfig
	if m.c.ServerUsingTLS {
		schema = "https"
		httpClientConfig.TLSConfig = m.c.TLSConfig
	}

	var scrapeConfigs []*config.ScrapeConfig

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
	instanceCfg.Name = instanceConfigKey(cfg)
	instanceCfg.ScrapeConfigs = scrapeConfigs
	instanceCfg.RemoteWrite = m.c.PrometheusRemoteWrite
	if common.WALTruncateFrequency > 0 {
		instanceCfg.WALTruncateFrequency = common.WALTruncateFrequency
	}
	return instanceCfg
}

// instanceConfigKey returns the instanceConfigKey for an integration Config.
func instanceConfigKey(cfg Config) string {
	return fmt.Sprintf("integration/%s", cfg.Name())
}

func (m *Manager) scrapeServiceDiscovery() discovery.Configs {
	// A blank host somehow works, but it then requires a sever name to be set under tls.
	newHost := *m.c.ListenHost
	if newHost == "" {
		newHost = "127.0.0.1"
	}
	localAddr := fmt.Sprintf("%s:%d", newHost, *m.c.ListenPort)
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

// WireAPI hooks up /metrics routes per-integration.
func (m *Manager) WireAPI(r *mux.Router) error {
	type handlerCacheEntry struct {
		handler http.Handler
		process *integrationProcess
	}
	var (
		handlerMut   sync.Mutex
		handlerCache = make(map[string]handlerCacheEntry)
	)

	// loadHandler will perform a dynamic lookup of an HTTP handler for an
	// integration. loadHandler should be called with a read lock on the
	// integrations mutex.
	loadHandler := func(name string) http.Handler {
		handlerMut.Lock()
		defer handlerMut.Unlock()

		key := fmt.Sprintf("integrations/%s", name)

		// Search the integration by name to see if it's still running.
		p, ok := m.integrations[key]
		if !ok {
			delete(handlerCache, name)
			return http.NotFoundHandler()
		}

		// Now look in the cache for a handler for the running process.
		cacheEntry, ok := handlerCache[key]
		if ok && cacheEntry.process == p {
			return cacheEntry.handler
		}

		// New integration process that hasn't been scraped before. Generate
		// a handler for it and cache it.
		handler, err := p.i.MetricsHandler()
		if err != nil {
			level.Error(m.logger).Log("msg", "could not create http handler for integration", "integration", name, "err", err)
			return http.HandlerFunc(internalServiceError)
		}

		cacheEntry = handlerCacheEntry{handler: handler, process: p}
		handlerCache[key] = cacheEntry
		return cacheEntry.handler
	}

	r.HandleFunc("/integrations/{name}/metrics", func(rw http.ResponseWriter, r *http.Request) {
		m.integrationsMut.RLock()
		defer m.integrationsMut.RUnlock()

		handler := loadHandler(mux.Vars(r)["name"])
		handler.ServeHTTP(rw, r)
	})

	return nil
}

func internalServiceError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
}

// Stop stops the manager and all of its integrations.
func (m *Manager) Stop() {
	m.cancel()
	m.wg.Done()
}
