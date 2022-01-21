// Package v2 provides a way to run and manage Grafana Agent
// "integrations," which integrate some external system (such as MySQL) to
// Grafana Agent's existing metrics, logging, and tracing subsystems.
//
// Integrations are implemented in sub-packages. Every integration must
// have an implementation of Config that configures the integration. The Config
// interface is then used to instantiate an instance of the Integration
// interface.
//
// Implementations of integrations implement extra functionality by
// implementing interface extensions. The Integration interface is the most
// basic interface that all integrations must implement. Extensions like
// the MetricsIntegration interface define an integration that supports
// metrics.
//
// Extension interfaces are used by the integrations subsystem to enable
// common use cases. New behaviors can be implemented by manually using
// the other subsystems of the agent provided in IntegrationOptions.
package v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/shared"
	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

var (
	// ErrInvalidUpdate is returned by ApplyConfig when the config cannot
	// be dynamically applied.
	ErrInvalidUpdate = fmt.Errorf("invalid dynamic update")
)

// Config provides a configuration and constructor for an integration.
type Config interface {
	// Name returns the YAML field name of the integration. Name is used
	// when unmarshaling the Config from YAML.
	Name() string

	// ApplyDefaults should apply default settings to Config.
	ApplyDefaults(shared.Globals) error

	// Identifier returns a string to uniquely identify the integration created
	// by this Config. Identifier must be unique for each integration that shares
	// the same Name.
	//
	// If there is no reasonable identifier to use for an integration,
	// Globals.AgentIdentifier may be used by default.
	Identifier(shared.Globals) (string, error)

	// NewIntegration should return a new Integration using the provided
	// Globals to help initialize the Integration.
	//
	// NewIntegration must be idempotent for a Config. Use
	// Integration.RunIntegration to do anything with side effects, such as
	// opening a port.
	NewIntegration(log.Logger, shared.Globals) (Integration, error)
}

// ComparableConfig extends Config with an ConfigEquals method.
type ComparableConfig interface {
	Config

	// ConfigEquals should return true if c is equal to the ComparableConfig.
	ConfigEquals(c Config) bool
}

// An Integration integrates some external system with Grafana Agent's existing
// subsystems.
//
// All integrations must at least implement this interface. More behaviors
// can be added by implementing additional *Integration interfaces, such
// as HTTPIntegration.
type Integration interface {
	// RunIntegration starts the integration and performs background tasks. It
	// must not return until ctx is canceled, even if there is no work to do.
	RunIntegration(ctx context.Context) error
}

// UpdateIntegration is an Integration whose config can be updated
// dynamically. Integrations that do not implement this interface will be shut
// down and re-instantiated with the new Config.
type UpdateIntegration interface {
	Integration

	// ApplyConfig should apply the config c to the integration. An error can be
	// returned if the Config is invalid. When this happens, the old config will
	// continue to run.
	//
	// If ApplyConfig returns ErrInvalidUpdate, the integration will be
	// recreated.
	ApplyConfig(c Config, g shared.Globals) error
}

// HTTPIntegration is an integration that exposes an HTTP handler.
//
// Integrations are given a unique base path prefix where HTTP requests will be
// routed. The prefix chosen for an integration is not guaranteed to be
// predictable.
type HTTPIntegration interface {
	Integration

	// Handler returns an http.Handler. Handler will be invoked for any endpoint
	// under prefix. If Handler returns nil, nothing will be called. Handler
	// may be called multiple times.
	//
	// prefix will not be removed from the HTTP request by default.
	Handler(prefix string) (http.Handler, error)
}

// MetricsIntegration is an integration that exposes Prometheus scrape targets.
//
// It is assumed, but not required, that HTTPIntegration is also implemented
// to expose metrics. See HTTPIntegration for more information about how
// HTTP works with integrations.
type MetricsIntegration interface {
	Integration

	// Targets should return the current set of active targets exposed by this
	// integration. Targets may be called multiple times throughout the lifecycle
	// of the integration. Targets will not be called when the integration is not
	// running.
	//
	// prefix will be the same prefixed passed to HTTPIntegration.Handler and
	// can be used to update __metrics_path__ for targets.
	Targets(ep Endpoint) []*targetgroup.Group

	// ScrapeConfigs configures automatic scraping of targets. ScrapeConfigs
	// is optional if an integration should not scrape itself.
	//
	// Unlike Targets, ScrapeConfigs is only called once per config load, and may be
	// called before the integration runs. Use the provided discovery.Configs to
	// discover the targets exposed by this integration.
	ScrapeConfigs(discovery.Configs) []*autoscrape.ScrapeConfig
}

// Endpoint is a location where something is exposed.
type Endpoint struct {
	// Hostname (and optional port) where endpoint is exposed.
	Host string
	// Base prefix of the endpoint.
	Prefix string
}
