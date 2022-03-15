//go:build linux
// +build linux

package ebpf

import (
	"fmt"
	"net/http"

	ebpf_config "github.com/cloudflare/ebpf_exporter/config"
	"github.com/cloudflare/ebpf_exporter/exporter"
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type config struct {
	Programs []ebpf_config.Program `yaml:"programs,omitempty"`
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
	return nil
}

func (c *config) Identifier(globals integrations.Globals) (string, error) {
	return c.Name(), nil
}

func (c *config) Name() string { return "ebpf" }

func (c *config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	var metricsCfg common.MetricsConfig
	metricsCfg.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)

	ebpf := &ebpfHandler{cfg: c}
	h, err := ebpf.createHandler()
	if err != nil {
		return nil, err
	}

	return metricsutils.NewMetricsHandlerIntegration(l, c, metricsCfg, globals, h)
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
