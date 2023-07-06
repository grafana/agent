package prometheusconvert

import (
	"reflect"

	"github.com/grafana/agent/converter/diag"
	prom_config "github.com/prometheus/prometheus/config"

	_ "github.com/prometheus/prometheus/discovery/install" // Register Prometheus SDs
)

func validate(promConfig *prom_config.Config) diag.Diagnostics {
	diags := validateGlobalConfig(&promConfig.GlobalConfig)

	newDiags := validateAlertingConfig(&promConfig.AlertingConfig)
	diags = append(diags, newDiags...)

	newDiags = validateRuleFilesConfig(promConfig.RuleFiles)
	diags = append(diags, newDiags...)

	newDiags = validateScrapeConfigs(promConfig.ScrapeConfigs)
	diags = append(diags, newDiags...)

	newDiags = validateStorageConfig(&promConfig.StorageConfig)
	diags = append(diags, newDiags...)

	newDiags = validateTracingConfig(&promConfig.TracingConfig)
	diags = append(diags, newDiags...)

	newDiags = validateRemoteWriteConfigs(promConfig.RemoteWriteConfigs)
	diags = append(diags, newDiags...)

	newDiags = validateRemoteReadConfigs(promConfig.RemoteReadConfigs)
	diags = append(diags, newDiags...)

	return diags
}

func validateGlobalConfig(globalConfig *prom_config.GlobalConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if globalConfig.EvaluationInterval != prom_config.DefaultGlobalConfig.EvaluationInterval {
		diags.Add(diag.SeverityLevelError, "unsupported global evaluation_interval config was provided")
	}

	if globalConfig.QueryLogFile != "" {
		diags.Add(diag.SeverityLevelError, "unsupported global query_log_file config was provided")
	}

	return diags
}

func validateAlertingConfig(alertingConfig *prom_config.AlertingConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(alertingConfig.AlertmanagerConfigs) > 0 || len(alertingConfig.AlertRelabelConfigs) > 0 {
		diags.Add(diag.SeverityLevelError, "unsupported alerting config was provided")
	}

	return diags
}

func validateRuleFilesConfig(ruleFilesConfig []string) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(ruleFilesConfig) > 0 {
		diags.Add(diag.SeverityLevelError, "unsupported rule_files config was provided")
	}

	return diags
}

func validateScrapeConfigs(scrapeConfigs []*prom_config.ScrapeConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, scrapeConfig := range scrapeConfigs {
		newDiags := validateHttpClientConfig(&scrapeConfig.HTTPClientConfig)
		diags = append(diags, newDiags...)
	}

	return diags
}

func validateStorageConfig(storageConfig *prom_config.StorageConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if storageConfig.TSDBConfig != nil || storageConfig.ExemplarsConfig != nil {
		diags.Add(diag.SeverityLevelError, "unsupported storage config was provided")
	}

	return diags
}

func validateTracingConfig(tracingConfig *prom_config.TracingConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if !reflect.DeepEqual(*tracingConfig, prom_config.TracingConfig{}) {
		diags.Add(diag.SeverityLevelError, "unsupported tracing config was provided")
	}

	return diags
}

func validateRemoteWriteConfigs(remoteWriteConfigs []*prom_config.RemoteWriteConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, remoteWriteConfig := range remoteWriteConfigs {
		if remoteWriteConfig.SigV4Config != nil {
			diags.Add(diag.SeverityLevelError, "unsupported remote_write sigv4 config was provided")
		}

		newDiags := validateHttpClientConfig(&remoteWriteConfig.HTTPClientConfig)
		diags = append(diags, newDiags...)
	}

	return diags
}

func validateRemoteReadConfigs(remoteReadConfigs []*prom_config.RemoteReadConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(remoteReadConfigs) > 0 {
		diags.Add(diag.SeverityLevelError, "unsupported remote_read config was provided")
	}

	return diags
}
