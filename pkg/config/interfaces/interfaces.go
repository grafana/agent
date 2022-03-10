package interfaces

import (
	"context"
	"net/http"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"

	"github.com/grafana/agent/pkg/util"

	"github.com/weaveworks/common/logging"

	"github.com/grafana/loki/clients/pkg/promtail/targets/file"

	"github.com/grafana/loki/clients/pkg/promtail/positions"

	"github.com/grafana/loki/clients/pkg/promtail/client"

	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/grafana/dskit/kv"
	"github.com/grafana/dskit/ring"

	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/prometheus/config"
	promConfig "github.com/prometheus/prometheus/config"
)

type AgentConfig interface {
	ReloadPort() int
	ReloadAddress() string

	ServerConfig() ServerConfig
	MetricsConfig() MetricsConfig
	LogsConfig() LogsConfig
	TracesConfig() TracesConfig
	IntegrationsConfig() IntegrationsConfig

	EnableConfigEndpoints() bool
	LogDeprecations(log *util.Logger)
}

type ServerConfig interface {
	HTTPListenPort() int
	HTTPListenAddress() string
	HTTPTLSConfig() web.TLSStruct
	LogLevel() logging.Level
	LogFormat() logging.Format
	Log() logging.Interface
}

type MetricsConfig interface {
	GlobalRemoteWrite() []*config.RemoteWriteConfig
	GlobalConfig() *config.GlobalConfig
	InstanceGlobalConfig() *instance.GlobalConfig
	InstanceRestartBackoff() time.Duration
	InstanceMode() instance.Mode

	WALDir() string
	WALCleanupAge() time.Duration
	WALCleanupPeriod() time.Duration

	ClusterConfig() ClusterConfig
	Compare(metricsConfig MetricsConfig) bool
}

type ClusterConfig interface {
	Enabled() bool
	Lifecycler() ring.LifecyclerConfig
	KVStore() kv.Config
	ClusterReshardEventTimeout() time.Duration
	Client() client.Config
	APIEnableGetConfiguration() bool
	ReshardInterval() time.Duration
	ReshardTimeout() time.Duration
	DangerousAllowReadingFiles() bool
}

type LogsConfig interface {
	Configs() []LogInstanceConfig
}

type LogInstanceConfig interface {
	Name() string
	//TODO interface this
	PositionsConfig() positions.Config
	//TODO interface this
	ClientConfigs() []client.Config
	ScrapeConfigs() []scrapeconfig.Config
	TargetConfig() file.Config
}

type TracesConfig interface {
	InstanceConfigs() []TraceInstanceConfig
}

type TraceInstanceConfig interface {
	Name() string
}

type IntegrationsVersion int

const (
	IntegrationsVersion1 IntegrationsVersion = iota
	IntegrationsVersion2
)

type IntegrationsConfig interface {
	Version() IntegrationsVersion
	V1Config() V1Integration
	V2Config() V2Integration
}

type V1Integration interface {
	Configs() []V1Config
	Compare(cfg V1Integration) bool
	PrometheusConfig() promConfig.GlobalConfig
	ListenPort() int
}

type V1Config interface {
	Enabled() bool
	Name() string
	NewIntegration(log log.Logger) (Integration, error)
	CommonInstanceKey() string
	InstanceKey(string) (string, error)
}

type V2Integration interface {
	Configs() []V2Config
	Compare(cfg V2Integration) bool
	PrometheusConfig() promConfig.GlobalConfig
}

type V2Config interface {
}

// Config provides the configuration and constructor for an integration.
type Config interface {
	// Name returns the name of the integration and the key that will be used to
	// pull the configuration from the Agent config YAML.
	Name() string

	// InstanceKey should return the key the reprsents the config, which will be
	// used to populate the value of the `instance` label for metrics.
	//
	// InstanceKey is given an agentKey that represents the agent process. This
	// may be used if the integration being configured applies to an entire
	// machine.
	//
	// This method may not be invoked if the instance key for a Config is
	// overridden.
	InstanceKey(agentKey string) (string, error)

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
	ScrapeConfigs() []ScrapeConfig

	// Run should start the integration and do any required tasks, if necessary.
	// For example, an Integration that requires a persistent connection to a
	// database would establish that connection here. If the integration doesn't
	// need to do anything, it should wait for the ctx to be canceled.
	//
	// An error will be returned if the integration failed. Integrations should
	// not return the ctx error.
	Run(ctx context.Context) error
}

// ScrapeConfig is a subset of options used by integrations to inform how samples
// should be scraped. It is utilized by the integrations.Manager to define a full
// Prometheus-compatible ScrapeConfig.
type ScrapeConfig struct {
	// JobName should be a unique name indicating the collection of samples to be
	// scraped. It will be prepended by "integrations/" when used by the integrations
	// manager.
	JobName string

	// MetricsPath is the path relative to the integration where metrics are exposed.
	// It should match a route added to the router provided in Integration.RegisterRoutes.
	// The path will be prepended by "/integrations/<integration name>" when read by
	// the integrations manager.
	MetricsPath string
}
