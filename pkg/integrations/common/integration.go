package common

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/config"
)

type Integration interface {
	// Name returns the name of the integration. Each registered integration must
	// have a unique name.
	Name() string

	// CommonConfig returns the set of common configuration values present across
	// all integrations.
	CommonConfig() config.Common

	// RegisterRoutes should register any HTTP handlers used for the integration.
	//
	// The router provided to RegisterRoutes is a subrouter for the path
	// /integrations/<integration name>. All routes should register to the
	// relative root path and will be automatically combined to the subroute. For
	// example, if a metric "database" registers a /metrics endpoint, it will
	// be exposed as /integrations/database/metrics.
	RegisterRoutes(r *mux.Router) error

	// ScrapeConfigs should return a set of integration scrape configs that inform
	// the integration how samples should be collected.
	ScrapeConfigs() []config.ScrapeConfig

	// Run should start the integration and do any required tasks. Run should *not*
	// exit until context is canceled. If an integration doesn't need to do anything,
	// it should simply wait for ctx to be canceled.
	Run(ctx context.Context) error
}
