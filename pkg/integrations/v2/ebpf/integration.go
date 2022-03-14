//go:build linux && ebpf_enabled
// +build linux,ebpf_enabled

package ebpf

import (
	"context"
	"fmt"
	"net/http"
	"path"

	ebpf_config "github.com/cloudflare/ebpf_exporter/config"
	"github.com/cloudflare/ebpf_exporter/exporter"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	prom_config "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

type config struct {
	Programs []ebpf_config.Program `yaml:"programs,omitempty"`

	common  common.MetricsConfig
	globals integrations.Globals
}

type ebpfHandler struct {
	cfg *config
}

func init() {
	integrations.Register(&config{}, integrations.TypeSingleton)
}

var defaultConfig = config{
	Programs: []ebpf_config.Program{},
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = defaultConfig
	type plain config

	return unmarshal((*plain)(c))
}

func (c *config) ApplyDefaults(globals integrations.Globals) error {
	c.common.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	return nil
}

func (c *config) Identifier(globals integrations.Globals) (string, error) {
	return c.Name(), nil
}

func (c *config) Name() string { return "ebpf" }

func (c *config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	c.globals = globals
	ebpf := &ebpfHandler{}
	ebpf.cfg = c

	return ebpf, nil
}

// RunIntegration implements the Integration interface and is
// the entrypoint for our integration.
func (e *ebpfHandler) RunIntegration(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Handler implements the HTTPIntegration interface.
func (e *ebpfHandler) Handler(prefix string) (http.Handler, error) {
	r := mux.NewRouter()
	h, err := e.createHandler()
	if err != nil {
		return nil, err
	}

	r.Handle(path.Join(prefix, "metrics"), h)
	return r, nil
}

func (e *ebpfHandler) createHandler() (http.HandlerFunc, error) {
	exp, err := exporter.New(ebpf_config.Config{Programs: e.cfg.Programs})
	if err != nil {
		return nil, fmt.Errorf("failed to create ebpf exporter with input config: %s", err)
	}

	err = exp.Attach()
	if err != nil {
		return nil, fmt.Errorf("failed to attach ebpf exporter: %s", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		registry := prometheus.NewRegistry()
		registry.MustRegister(exp)
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
		return
	}, nil
}

// Targets implements the MetricsIntegration interface.
func (e *ebpfHandler) Targets(ep integrations.Endpoint) []*targetgroup.Group {
	name := e.cfg.Name()
	integrationNameValue := model.LabelValue("integrations/" + name)
	group := &targetgroup.Group{
		Labels: model.LabelSet{
			model.JobLabel:   integrationNameValue,
			"agent_hostname": model.LabelValue(e.cfg.globals.AgentIdentifier),

			// Meta labels that can be used during SD.
			"__meta_agent_integration_name":       model.LabelValue(name),
			"__meta_agent_integration_autoscrape": model.LabelValue(boolToString(*e.cfg.common.Autoscrape.Enable)),
			"__meta_agent_integration_instance":   model.LabelValue(e.cfg.Name()),
		},
		Source: fmt.Sprintf("%s/%s", name, name),
	}

	for _, lbl := range e.cfg.common.ExtraLabels {
		group.Labels[model.LabelName(lbl.Name)] = model.LabelValue(lbl.Value)
	}

	group.Targets = append(group.Targets, model.LabelSet{
		model.AddressLabel:     model.LabelValue(ep.Host),
		model.MetricsPathLabel: model.LabelValue(path.Join(ep.Prefix, "/metrics")),
	})

	return []*targetgroup.Group{group}
}

// ScrapeConfigs implements the MetricsIntegration interface.
func (e *ebpfHandler) ScrapeConfigs(sd discovery.Configs) []*autoscrape.ScrapeConfig {
	if !*e.cfg.common.Autoscrape.Enable {
		return nil
	}

	cfg := prom_config.DefaultScrapeConfig
	cfg.JobName = e.cfg.Name()
	cfg.Scheme = e.cfg.globals.AgentBaseURL.Scheme
	cfg.HTTPClientConfig = e.cfg.globals.SubsystemOpts.ClientConfig
	cfg.ServiceDiscoveryConfigs = sd
	cfg.ScrapeInterval = e.cfg.common.Autoscrape.ScrapeInterval
	cfg.ScrapeTimeout = e.cfg.common.Autoscrape.ScrapeTimeout
	cfg.RelabelConfigs = e.cfg.common.Autoscrape.RelabelConfigs
	cfg.MetricRelabelConfigs = e.cfg.common.Autoscrape.MetricRelabelConfigs

	return []*autoscrape.ScrapeConfig{{
		Instance: e.cfg.common.Autoscrape.MetricsInstance,
		Config:   cfg,
	}}
}

// ServeHTTP kicks off the integration's HTTP handler.
func (e *ebpfHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.createHandler()
}

func boolToString(b bool) string {
	switch b {
	case true:
		return "1"
	default:
		return "0"
	}
}
