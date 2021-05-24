package integrations

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations/config"
)

// Config provides the configuration and constructor for an integration.
type Config interface {
	// Name returns the name of the integration and the key that will be used to
	// pull the configuration from the Agent config YAML.
	Name() string

	// CommonConfig returns the set of common configuration values present across
	// all integrations.
	CommonConfig() config.Common

	// NewIntegration returns an integration for the given with the given logger.
	NewIntegration(l log.Logger) (Integration, error)
}

// An Integration is a process that integrates with some external system and
// pulls telemetry data.
type Integration interface {
	// MetricsHandler returns an http.Handler that will return metrics.
	MetricsHandler() (http.Handler, error)

	// ScrapeConfigs returns a set of scrape configs that determine where metrics
	// can be scraped.
	ScrapeConfigs() []config.ScrapeConfig

	// Run should start the integration and do any required tasks, if necessary.
	// For example, an Integration that requires a persistent connection to a
	// database would establish that connection here. If the integration doesn't
	// need to do anything, it should wait for the ctx to be canceled.
	Run(ctx context.Context) error

	CustomHandlers() map[string]http.Handler
}
