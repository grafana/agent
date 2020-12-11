// +build !linux

// Package process_exporter embeds https://github.com/ncabatoff/process-exporter
package process_exporter //nolint:golint

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/config"
)

// Integration is the process_exporter integration. On non-Linux platforms,
// this integration does nothing and will print a warning if enabled.
type Integration struct {
	c *Config
}

func New(logger log.Logger, c *Config) (*Integration, error) {
	level.Warn(logger).Log("msg", "the process_exporter only works on Linux; enabling it otherwise will do nothing")
	return &Integration{c: c}, nil
}

// RegisterRoutes satisfies Integration.RegisterRoutes.
func (i *Integration) RegisterRoutes(r *mux.Router) error {
	return nil
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (i *Integration) ScrapeConfigs() []config.ScrapeConfig {
	// No-op: nothing to scrape.
	return []config.ScrapeConfig{}
}

// Run satisfies Integration.Run.
func (i *Integration) Run(ctx context.Context) error {
	// We don't need to do anything here, so we can just wait for the context to
	// finish.
	<-ctx.Done()
	return ctx.Err()
}
