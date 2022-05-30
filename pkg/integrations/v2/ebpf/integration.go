//go:build linux && amd64 && !noebpf
// +build linux,amd64,!noebpf

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

type ebpfHandler struct {
	cfg *Config
}

// NewIntegration instantiates a new integrations.MetricsIntegration
// which will handle requests to the eBPF exporter.
func (c *Config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
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
	}, nil
}
