package integrations

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/metrics/instance"
	common_config "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	prom_config "github.com/prometheus/prometheus/config"
	http_sd "github.com/prometheus/prometheus/discovery/http"
)

// The endpoint to use for HTTP SD. The API uses query parameters to customize
// what gets returned by discovery.
const IntegrationsSDEndpoint = "/agent/api/v1/metrics/integrations/sd"

// DefaultSubsystemOptions holds the default settings for a Controller.
var DefaultSubsystemOptions = SubsystemOptions{
	ScrapeIntegrations: true,
}

// SubsystemOptions controls how the integrations subsystem behaves.
type SubsystemOptions struct {
	// When true, scrapes metrics from integrations.
	ScrapeIntegrations bool `yaml:"scrape_integrations,omitempty"`
	// Prometheus RW configs to use for self-scraping integrations.
	PrometheusRemoteWrite []*prom_config.RemoteWriteConfig `yaml:"prometheus_remote_write,omitempty"`

	// Configs are configurations of integration to create. Unmarshaled through
	// the custom UnmarshalYAML method of Controller.
	Configs []Config `yaml:"-"`

	// Extra labels to add for all integration telemetry.
	Labels model.LabelSet `yaml:"labels,omitempty"`

	// Override settings to self-communicate with agent.
	ClientConfig common_config.HTTPClientConfig `yaml:"client_config,omitempty"`
}

// MarshalYAML implements yaml.Marshaler for SubsystemOptions. Integrations
// will be marshaled inline.
func (o *SubsystemOptions) MarshalYAML() (interface{}, error) {
	return MarshalYAML(o)
}

// UnmarshalYAML implements yaml.Unmarshaler for SubsystemOptions. Inline
// integrations will be unmarshaled into o.Configs.
func (o *SubsystemOptions) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*o = DefaultSubsystemOptions
	return UnmarshalYAML(o, unmarshal)
}

// Subsystem runs the integrations subsystem, managing a set of integrations.
type Subsystem struct {
	logger log.Logger

	mut        sync.RWMutex
	sopts      SubsystemOptions
	iopts      Options
	apiHandler http.Handler // generated from controller

	ctrl             *controller
	stopController   context.CancelFunc
	controllerExited chan struct{}
}

// NewSubsystem creates and starts a new integrations Subsystem. Every field in
// IntegrationOptions must be filled out.
func NewSubsystem(sopts SubsystemOptions, iopts Options) (*Subsystem, error) {
	ctrl, err := newController(sopts.Configs, iopts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	ctrlExited := make(chan struct{})
	go func() {
		ctrl.run(ctx)
		close(ctrlExited)
	}()

	s := &Subsystem{
		logger: iopts.Logger,

		sopts: sopts,
		iopts: iopts,

		ctrl:             ctrl,
		stopController:   cancel,
		controllerExited: ctrlExited,
	}
	if err := s.ApplyConfig(sopts, iopts); err != nil {
		cancel()
		return nil, err
	}
	return s, nil
}

// ApplyConfig updates the configuration of the integrations Subsystem and
// options to pass to integrations.
func (s *Subsystem) ApplyConfig(sopts SubsystemOptions, opts Options) error {
	const prefix = "/integrations/"

	s.mut.Lock()
	defer s.mut.Unlock()

	if err := s.ctrl.UpdateController(sopts.Configs, opts); err != nil {
		return fmt.Errorf("error applying integrations: %w", err)
	}

	var firstErr error
	saveFirstErr := func(err error) {
		if firstErr == nil {
			firstErr = err
		}
	}

	// Set up HTTP wiring
	{
		handler, err := s.ctrl.Handler(prefix)
		if err != nil {
			saveFirstErr(fmt.Errorf("HTTP handler update failed: %w", err))
		}
		if handler != nil {
			s.apiHandler = handler
		}
	}

	// Set up integrations SD
	{
		// TODO(rfratto): cache results?
		// TODO(rfratto): track active targetsCh?
		targetsCh := s.ctrl.Targets(prefix)
		_ = targetsCh
	}

	// Set up self-scraping
	{
		httpSDConfig := http_sd.DefaultSDConfig
		httpSDConfig.HTTPClientConfig = opts.AgentHTTPClientConfig
		httpSDConfig.RefreshInterval = model.Duration(time.Second * 5) // TODO(rfratto): make configurable?

		apiURL := opts.CloneAgentBaseURL()
		apiURL.Path = IntegrationsSDEndpoint
		httpSDConfig.URL = apiURL.String()

		scrapeConfigs := s.ctrl.ScrapeConfigs(prefix, &httpSDConfig)
		if len(scrapeConfigs) == 0 {
			// We're not going to self scrape if there are no configs. Try to delete
			// the previous instance for self-scraping if one was running.
			_ = opts.Metrics.InstanceManager().DeleteConfig("integrations")
		} else {
			instanceCfg := instance.DefaultConfig
			instanceCfg.Name = "integrations"
			instanceCfg.ScrapeConfigs = scrapeConfigs
			instanceCfg.RemoteWrite = sopts.PrometheusRemoteWrite

			if err := opts.Metrics.Validate(&instanceCfg); err != nil {
				saveFirstErr(fmt.Errorf("failed to apply self-scraping configs: validation: %w", err))
			} else if err := opts.Metrics.InstanceManager().ApplyConfig(instanceCfg); err != nil {
				saveFirstErr(fmt.Errorf("failed to apply self-scraping configs: %w", err))
			}
		}
	}

	s.sopts = sopts
	s.iopts = opts
	return firstErr
}

// WireAPI hooks up integration endpoints to r.
func (s *Subsystem) WireAPI(r *mux.Router) {
	r.PathPrefix("/integrations").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.mut.RLock()
		handler := s.apiHandler
		s.mut.RUnlock()

		if handler == nil {
			rw.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(rw, "Integrations HTTP endpoints not yet available")
			return
		}
		handler.ServeHTTP(rw, r)
	})

	// TODO(rfratto): SD API
}

// Stop stops the manager and all running integrations. Blocks until all
// running integrations exit.
func (s *Subsystem) Stop() {
	s.stopController()
	<-s.controllerExited
}
