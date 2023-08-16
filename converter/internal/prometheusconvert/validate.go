package prometheusconvert

import (
	"fmt"
	"reflect"

	"github.com/grafana/agent/converter/diag"
	prom_config "github.com/prometheus/prometheus/config"
	prom_discover "github.com/prometheus/prometheus/discovery"

	prom_aws "github.com/prometheus/prometheus/discovery/aws"
	prom_azure "github.com/prometheus/prometheus/discovery/azure"
	prom_consul "github.com/prometheus/prometheus/discovery/consul"
	prom_digitalocean "github.com/prometheus/prometheus/discovery/digitalocean"
	prom_dns "github.com/prometheus/prometheus/discovery/dns"
	prom_file "github.com/prometheus/prometheus/discovery/file"
	prom_gce "github.com/prometheus/prometheus/discovery/gce"
	_ "github.com/prometheus/prometheus/discovery/install" // Register Prometheus SDs
	prom_kubernetes "github.com/prometheus/prometheus/discovery/kubernetes"
	prom_docker "github.com/prometheus/prometheus/discovery/moby"
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
		diags.AddAll(validatePrometheusScrape(scrapeConfig))

		for _, serviceDiscoveryConfig := range scrapeConfig.ServiceDiscoveryConfigs {
			switch sdc := serviceDiscoveryConfig.(type) {
			case prom_discover.StaticConfig:
				diags.AddAll(validateScrapeTargets(sdc))
			case *prom_azure.SDConfig:
				diags.AddAll(ValidateDiscoveryAzure(sdc))
			case *prom_consul.SDConfig:
				diags.AddAll(validateDiscoveryConsul(sdc))
			case *prom_digitalocean.SDConfig:
				diags.AddAll(ValidateDiscoveryDigitalOcean(sdc))
			case *prom_dns.SDConfig:
				diags.AddAll(validateDiscoveryDns(sdc))
			case *prom_docker.DockerSDConfig:
				diags.AddAll(validateDiscoveryDocker(sdc))
			case *prom_aws.EC2SDConfig:
				diags.AddAll(ValidateDiscoveryEC2(sdc))
			case *prom_file.SDConfig:
				diags.AddAll(validateDiscoveryFile(sdc))
			case *prom_gce.SDConfig:
				diags.AddAll(ValidateDiscoveryGCE(sdc))
			case *prom_kubernetes.SDConfig:
				diags.AddAll(validateDiscoveryKubernetes(sdc))
			case *prom_aws.LightsailSDConfig:
				diags.AddAll(validateDiscoveryLightsail(sdc))
			default:
				diags.Add(diag.SeverityLevelError, fmt.Sprintf("unsupported service discovery %s was provided", serviceDiscoveryConfig.Name()))
			}
		}
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
		diags.AddAll(validateRemoteWriteConfig(remoteWriteConfig))
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
