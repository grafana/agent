// Package mysqld_exporter embeds https://github.com/prometheus/mysqld_exporter
package mysqld_exporter //nolint:golint

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/mysqld_exporter/collector"
)

// Integration is the mysqld_exporter integration. The integration scrapes metrics
// from a mysqld process.
type Integration struct {
	c        Config
	logger   log.Logger
	exporter *collector.Exporter
}

// New creates a new mysqld_exporter integration.
func New(log log.Logger, c Config) (*Integration, error) {
	dsn := c.DataSourceName
	if len(dsn) == 0 {
		dsn = os.Getenv("MYSQLD_EXPORTER_DATA_SOURCE_NAME")
	}
	if len(dsn) == 0 {
		return nil, fmt.Errorf("cannot create mysqld_exporter; neither mysqld_exporter.data_source_name or $MYSQLD_EXPORTER_DATA_SOURCE_NAME is set")
	}

	scrapers := GetScrapers(c)
	exporter := collector.New(context.Background(), dsn, collector.NewMetrics(), scrapers, log, collector.Config{
		LockTimeout:   c.LockWaitTimeout,
		SlowLogFilter: c.LogSlowFilter,
	})

	level.Debug(log).Log("msg", "enabled mysqld_exporter scrapers")
	for _, scraper := range scrapers {
		level.Debug(log).Log("scraper", scraper.Name())
	}

	return &Integration{
		c:        c,
		logger:   log,
		exporter: exporter,
	}, nil
}

// CommonConfig satisfies Integration.CommonConfig.
func (i *Integration) CommonConfig() config.Common { return i.c.CommonConfig }

// Name satisfies Integration.Name.
func (i *Integration) Name() string { return "mysqld_exporter" }

// RegisterRoutes satisfies Integration.RegisterRoutes. The mux.Router provided
// here is expected to be a subrouter, where all registered paths will be
// registered within that subroute.
func (i *Integration) RegisterRoutes(r *mux.Router) error {
	handler, err := i.handler()
	if err != nil {
		return err
	}

	r.Handle("/metrics", handler)
	return nil
}

func (i *Integration) handler() (http.Handler, error) {
	r := prometheus.NewRegistry()
	if err := r.Register(i.exporter); err != nil {
		return nil, fmt.Errorf("couldn't register mysqld_exporter collector: %w", err)
	}
	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})
	return handler, nil
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (i *Integration) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{
		JobName:     i.Name(),
		MetricsPath: "/metrics",
	}}
}

// Run satisfies Integration.Run.
func (i *Integration) Run(ctx context.Context) error {
	// We don't need to do anything here, so we can just wait for the context to
	// finish.
	<-ctx.Done()
	return ctx.Err()
}
