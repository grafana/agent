package staticconvert

import (
	"fmt"

	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/integrations/apache_http"
	"github.com/grafana/agent/pkg/integrations/blackbox_exporter"
	"github.com/grafana/agent/pkg/integrations/cloudwatch_exporter"
	"github.com/grafana/agent/pkg/integrations/node_exporter"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/server"
	"github.com/grafana/agent/pkg/traces"

	_ "github.com/prometheus/prometheus/discovery/install" // Register Prometheus SDs
)

func validate(staticConfig *config.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	_, grpcListenPort, _ := staticConfig.ServerFlags.GRPC.ListenHostPort()

	diags.AddAll(validateCommandLine())
	diags.AddAll(validateServer(staticConfig.Server))
	diags.AddAll(validateMetrics(staticConfig.Metrics, grpcListenPort))
	diags.AddAll(validateIntegrations(staticConfig.Integrations))
	diags.AddAll(validateTraces(staticConfig.Traces))
	diags.AddAll(validateLogs(staticConfig.Logs))
	diags.AddAll(validateAgentManagement(staticConfig.AgentManagement))

	return diags
}

func validateCommandLine() diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Add(diag.SeverityLevelWarn, "Please review your agent command line flags and ensure they are set in your Flow mode config file where necessary.")

	return diags
}

func validateServer(serverConfig *server.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	defaultServerConfig := server.DefaultConfig()
	diags.AddAll(common.UnsupportedNotDeepEquals(serverConfig.LogLevel.Level.Logrus, defaultServerConfig.LogLevel.Level.Logrus, "log_level server"))
	diags.AddAll(common.UnsupportedNotDeepEquals(serverConfig.LogFormat, defaultServerConfig.LogFormat, "log_format server"))
	diags.AddAll(common.UnsupportedNotDeepEquals(serverConfig.GRPC, defaultServerConfig.GRPC, "grpc_tls_config server"))
	diags.AddAll(common.UnsupportedNotDeepEquals(serverConfig.HTTP, defaultServerConfig.HTTP, "http_tls_config server"))

	return diags
}

func validateMetrics(metricsConfig metrics.Config, grpcListenPort int) diag.Diagnostics {
	var diags diag.Diagnostics

	defaultMetrics := config.DefaultConfig().Metrics
	defaultMetrics.ServiceConfig.Lifecycler.ListenPort = grpcListenPort
	diags.AddAll(common.UnsupportedNotDeepEquals(metricsConfig.WALCleanupAge, defaultMetrics.WALCleanupAge, "wal_cleanup_age metrics"))

	if metricsConfig.WALDir != defaultMetrics.WALDir {
		diags.Add(diag.SeverityLevelError, "unsupported wal_directory metrics config was provided. use the run command flag --storage.path for Flow mode instead.")
	}

	diags.AddAll(common.UnsupportedNotEquals(metricsConfig.WALCleanupPeriod, defaultMetrics.WALCleanupPeriod, "wal_cleanup_period metrics"))
	diags.AddAll(common.UnsupportedNotDeepEquals(metricsConfig.ServiceConfig, defaultMetrics.ServiceConfig, "scraping_service metrics"))
	diags.AddAll(common.UnsupportedNotDeepEquals(metricsConfig.ServiceClientConfig, defaultMetrics.ServiceClientConfig, "scraping_service_client metrics"))
	diags.AddAll(common.UnsupportedNotEquals(metricsConfig.InstanceRestartBackoff, defaultMetrics.InstanceRestartBackoff, "instance_restart_backoff metrics"))
	diags.AddAll(common.UnsupportedNotEquals(metricsConfig.InstanceMode, defaultMetrics.InstanceMode, "instance_mode metrics"))
	diags.AddAll(common.UnsupportedNotEquals(metricsConfig.DisableKeepAlives, defaultMetrics.DisableKeepAlives, "http_disable_keepalives metrics"))
	diags.AddAll(common.UnsupportedNotEquals(metricsConfig.IdleConnTimeout, defaultMetrics.IdleConnTimeout, "http_idle_conn_timeout metrics"))

	return diags
}

func validateIntegrations(integrationsConfig config.VersionedIntegrations) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(integrationsConfig.EnabledIntegrations()) == 0 {
		return diags
	}

	for _, integration := range integrationsConfig.ConfigV1.Integrations {
		switch itg := integration.Config.(type) {
		case *apache_http.Config:
		case *node_exporter.Config:
		case *blackbox_exporter.Config:
		case *cloudwatch_exporter.Config:
		default:
			diags.Add(diag.SeverityLevelError, fmt.Sprintf("unsupported integration %s was provided.", itg.Name()))
		}
	}

	return diags
}

func validateTraces(tracesConfig traces.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.AddAll(common.UnsupportedNotDeepEquals(tracesConfig, traces.Config{}, "traces"))

	return diags
}

func validateLogs(logsConfig *logs.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func validateAgentManagement(agentManagementConfig config.AgentManagementConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.AddAll(common.UnsupportedNotDeepEquals(agentManagementConfig, config.AgentManagementConfig{}, "agent_management"))

	return diags
}
