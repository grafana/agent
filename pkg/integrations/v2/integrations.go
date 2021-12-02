// Package integrations provides a way to run and manage Grafana Agent
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
package integrations

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/traces"
	common_config "github.com/prometheus/common/config"
	prom_config "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

var (
	// ErrDisabled may be returned by NewIntegration to indicate that an
	// integration should not run.
	ErrDisabled = fmt.Errorf("integration disabled")
	// ErrInvalidUpdate is returned by ApplyConfig when the config cannot
	// be dynamically applied.
	ErrInvalidUpdate = fmt.Errorf("invalid dynamic update")
)

// Config provides a configuration and constructor for an integration.
type Config interface {
	// Name returns the YAML field name of the integration. Name is used
	// when unmarshaling the Config from YAML.
	Name() string

	// Identifier returns a string to uniquely identify the integration created
	// by this Config. Identifier must be unique for each integration that shares
	// the same Name.
	//
	// If there is no reasonable identifier to use for an integration,
	// Globals.AgentIdentifier may be used by default.
	Identifier(Globals) (string, error)

	// NewIntegration should return a new Integration using the provided
	// Globals to help initialize the Integration.
	//
	// NewIntegration must be idempotent for a Config. Use Integration.Run to do
	// anything with side effects, such as opening a port.
	//
	// NewIntegration may return ErrDisabled if the integration should not be
	// run.
	NewIntegration(log.Logger, Globals) (Integration, error)
}

// ComparableConfig extends Config with an ConfigEquals method.
type ComparableConfig interface {
	Config

	// ConfigEquals should return true if c is equal to the ComparableConfig.
	ConfigEquals(c Config) bool
}

// Globals are used to pass around subsystem-wide settings that integrations
// can take advantage of.
type Globals struct {
	// AgentIdentifier provides an identifier for the running agent. This can
	// be used for labelling whenever appropriate.
	//
	// AgentIdentifier will be set to the hostname:port of the running agent.
	// TODO(rfratto): flag to override identifier at agent level?
	AgentIdentifier string

	// Some integrations may wish to interact with various subsystems for their
	// implementation if the desired behavior is not supported natively by the
	// integration manager.

	Metrics *metrics.Agent // Metrics subsystem
	Logs    *logs.Logs     // Logs subsystem
	Tracing *traces.Traces // Traces subsystem

	// BaseURL to use to invoke methods against the embedded HTTP server.
	AgentBaseURL *url.URL
	// HTTP config to invoke methods against the embedded HTTP server.
	AgentHTTPClientConfig common_config.HTTPClientConfig
}

// CloneAgentBaseURL returns a copy of AgentBaseURL that can be modified.
func (g Globals) CloneAgentBaseURL() *url.URL {
	if g.AgentBaseURL == nil {
		return nil
	}
	rawUrl := g.AgentBaseURL.String()
	u, err := url.Parse(rawUrl)
	if err != nil {
		// The URL shouldn't be invalid at this point
		panic(err)
	}
	return u
}

// An Integration integrates some external system with Grafana Agent's existing
// subsystems.
//
// All integrations must at least implement this interface. More behaviors
// can be added by implementing additional *Integration interfaces, such
// as HTTPIntegration.
type Integration interface {
	// Run starts the integration and performs background tasks. Run must not
	// return until ctx is canceled, even if there is no work to do.
	RunIntegration(ctx context.Context) error
}

// ConfigurableIntegration is an Integration whose config can be updated
// dynamically. Integrations that do not implement this interface will be shut
// down and re-instantiated with the new Config.
type UpdateIntegration interface {
	Integration

	// ApplyConfig should apply the config c to the integration. An error can be
	// returned if the Config is invalid. When this happens, the old config will
	// continue to run.
	//
	// If ApplyConfig returns ErrDisabled, the integration will be stopped.
	// If ApplyConfig returns ErrInvalidUpdate, the integration will be recreatd.
	ApplyConfig(c Config, g Globals) error
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
	// of the integration.
	//
	// prefix will be the same prefixed passed to HTTPIntegration.Handler and
	// can be used to update __metrics_path__ for targets.
	Targets(prefix string) []*targetgroup.Group

	// ScrapeConfigs configures automatic scraping of targets. ScrapeConfigs
	// is optional if an integration should not scrape itself.
	//
	// Unlike Targets, ScrapeConfigs is only called once per config load. Use the
	// provided discovery.Configs to discover the targets exposed by this
	// integration.
	ScrapeConfigs(discovery.Configs) []*prom_config.ScrapeConfig
}
