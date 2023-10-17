package prometheusconvert

import (
	"reflect"

	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/prometheusconvert/component"
	prom_config "github.com/prometheus/prometheus/config"
	prom_discover "github.com/prometheus/prometheus/discovery"

	_ "github.com/prometheus/prometheus/discovery/install" // Register Prometheus SDs
)

func validate(promConfig *prom_config.Config) diag.Diagnostics {
	diags := validateGlobalConfig(&promConfig.GlobalConfig)
	diags.AddAll(validateAlertingConfig(&promConfig.AlertingConfig))
	diags.AddAll(validateRuleFilesConfig(promConfig.RuleFiles))
	diags.AddAll(validateScrapeConfigs(promConfig.ScrapeConfigs))
	diags.AddAll(validateStorageConfig(&promConfig.StorageConfig))
	diags.AddAll(validateTracingConfig(&promConfig.TracingConfig))
	diags.AddAll(validateRemoteWriteConfigs(promConfig.RemoteWriteConfigs))
	diags.AddAll(validateRemoteReadConfigs(promConfig.RemoteReadConfigs))

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
		diags.AddAll(component.ValidatePrometheusScrape(scrapeConfig))
		diags.AddAll(ValidateServiceDiscoveryConfigs(scrapeConfig.ServiceDiscoveryConfigs))
	}
	return diags
}

func ValidateServiceDiscoveryConfigs(serviceDiscoveryConfigs prom_discover.Configs) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, serviceDiscoveryConfig := range serviceDiscoveryConfigs {
		diags.AddAll(component.ValidateServiceDiscoveryConfig(serviceDiscoveryConfig))
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
		diags.AddAll(component.ValidateRemoteWriteConfig(remoteWriteConfig))
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
