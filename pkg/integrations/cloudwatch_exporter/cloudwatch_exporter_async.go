package cloudwatch_exporter

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-kit/log"
	yace "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg"
	yaceClients "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/clients"
	yaceClientsV1 "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/clients/v1"
	yaceConf "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/grafana/agent/pkg/integrations/config"
)

// asyncExporter wraps YACE entrypoint around an Integration implementation
type asyncExporter struct {
	name         string
	logger       yaceLoggerWrapper
	sessionCache yaceClients.Factory
	scrapeConf   yaceConf.ScrapeConf
	registry     atomic.Pointer[prometheus.Registry]
	// scrapeInterval is the frequency in which a background go-routine collects new AWS metrics via YACE.
	scrapeInterval time.Duration
}

// NewDecoupledCloudwatchExporter creates a new YACE wrapper, that implements Integration. The decouple feature spawns a
// background go-routine to perform YACE metric collection allowing for a decoupled collection of AWS metrics from the
// ServerHandler.
func NewDecoupledCloudwatchExporter(name string, logger log.Logger, conf yaceConf.ScrapeConf, scrapeInterval time.Duration, fipsEnabled, debug bool) *asyncExporter {
	loggerWrapper := yaceLoggerWrapper{
		debug: debug,
		log:   logger,
	}
	return &asyncExporter{
		name:           name,
		logger:         loggerWrapper,
		sessionCache:   yaceClientsV1.NewFactory(conf, fipsEnabled, loggerWrapper),
		scrapeConf:     conf,
		registry:       atomic.Pointer[prometheus.Registry]{},
		scrapeInterval: scrapeInterval,
	}
}

func (e *asyncExporter) MetricsHandler() (http.Handler, error) {
	// Wrapping handler to have logging around handler
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		e.logger.Debug("Running collect in cloudwatch_exporter")
		reg := e.registry.Load()
		if reg == nil {
			e.logger.Warn("cloudwatch_exporter prometheus metric registry is empty")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, req)
	})
	return h, nil
}

func (e *asyncExporter) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{
		JobName:     e.name,
		MetricsPath: "/metrics",
	}}
}

func (e *asyncExporter) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.scrapeInterval)
	defer ticker.Stop()
	for {
		reg := prometheus.NewRegistry()
		err := yace.UpdateMetrics(
			ctx,
			e.logger,
			e.scrapeConf,
			reg,
			e.sessionCache,
			yace.MetricsPerQuery(metricsPerQuery),
			yace.LabelsSnakeCase(labelsSnakeCase),
			yace.CloudWatchAPIConcurrency(cloudWatchConcurrency),
			yace.TaggingAPIConcurrency(tagConcurrency),
			// Enable max-dimension-associator feature flag
			// https://github.com/nerdswords/yet-another-cloudwatch-exporter/blob/master/docs/feature_flags.md#new-associator-algorithm
			yace.EnableFeatureFlag(yaceConf.MaxDimensionsAssociator),
		)
		if err != nil {
			e.logger.Error(err, "Error collecting cloudwatch metrics")
		}
		// always update the registry even on error, to ensure we don't expose stale metrics from the previous
		// registry
		e.registry.Store(reg)

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
