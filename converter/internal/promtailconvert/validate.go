package promtailconvert

import (
	"github.com/grafana/agent/converter/diag"
	promtailcfg "github.com/grafana/loki/clients/pkg/promtail/config"
	"github.com/grafana/loki/clients/pkg/promtail/targets/file"
)

// validateTopLevelConfig validates the top-level config for any unsupported features. There may still be some
// other unsupported features in scope of each config block, which are raised by their respective conversion code.
func validateTopLevelConfig(cfg *promtailcfg.Config, diags *diag.Diagnostics) {
	// We currently do not support the new global file watch config. It's an error, since setting it indicates
	// some advanced tuning which the user likely needs.
	if cfg.Global.FileWatch != file.DefaultWatchConig {
		diags.Add(diag.SeverityLevelError, "global/file_watch_config is not supported")
	}

	// The positions global config is not supported in Flow Mode.
	if cfg.PositionsConfig != DefaultPositionsConfig() {
		diags.Add(
			diag.SeverityLevelError,
			"global positions configuration is not supported - each Flow Mode's loki.source.file component "+
				"has its own positions file in the component's data directory",
		)
	}

	// The global and per-client stream lag labels is deprecated and has no effect.
	if len(cfg.Options.StreamLagLabels) > 0 {
		diags.Add(
			diag.SeverityLevelWarn,
			"stream_lag_labels is deprecated and the associated metric has been removed",
		)
	}

	// WAL support is still work in progress and not documented. Enabling it won't work, so it's an error.
	if cfg.WAL.Enabled {
		diags.Add(
			diag.SeverityLevelError,
			"Promtail's WAL is currently not supported in Flow Mode",
		)
	}

	// Not yet supported, see https://github.com/grafana/agent/issues/4342. It's an error since we want to
	// err on the safe side.
	//TODO(thampiotr): seems like it's possible to support this using loki.process component
	if cfg.LimitsConfig != DefaultLimitsConfig() {
		diags.Add(
			diag.SeverityLevelError,
			"limits_config is not yet supported in Flow Mode",
		)
	}

	// We cannot migrate the tracing config to Flow Mode, since in promtail it relies on
	// environment variables that can be set or not and depending on what is set, different
	// features of tracing are configured. We'd need to have conditionals in the
	// flow config to translate this. See https://www.jaegertracing.io/docs/1.16/client-features/
	if cfg.Tracing.Enabled {
		diags.Add(
			diag.SeverityLevelWarn,
			"tracing configuration cannot be migrated to Flow Mode automatically - please "+
				"refer to documentation on how to configure tracing in Flow Mode",
		)
	}

	if cfg.TargetConfig.Stdin {
		diags.Add(
			diag.SeverityLevelError,
			"reading targets from stdin is not supported in Flow Mode configuration file",
		)
	}
	if cfg.ServerConfig.ProfilingEnabled {
		diags.Add(diag.SeverityLevelWarn, "server.profiling_enabled is not supported - use Agent's "+
			"main HTTP server's profiling endpoints instead.")
	}

	if cfg.ServerConfig.RegisterInstrumentation {
		diags.Add(diag.SeverityLevelWarn, "server.register_instrumentation is not supported - Flow mode "+
			"components expose their metrics automatically in their own metrics namespace")
	}

	if cfg.ServerConfig.LogLevel.String() != "info" {
		diags.Add(diag.SeverityLevelWarn, "server.log_level is not supported - Flow mode "+
			"components may produce different logs")
	}

	if cfg.ServerConfig.PathPrefix != "" {
		diags.Add(diag.SeverityLevelError, "server.http_path_prefix is not supported")
	}

	if cfg.ServerConfig.HealthCheckTarget != nil && !*cfg.ServerConfig.HealthCheckTarget {
		diags.Add(diag.SeverityLevelWarn, "server.health_check_target disabling is not supported in Flow mode")
	}
}
