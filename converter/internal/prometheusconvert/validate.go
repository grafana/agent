package prometheusconvert

import (
	"fmt"

	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
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
	prom_ionos "github.com/prometheus/prometheus/discovery/ionos"
	prom_kubernetes "github.com/prometheus/prometheus/discovery/kubernetes"
	prom_linode "github.com/prometheus/prometheus/discovery/linode"
	prom_marathon "github.com/prometheus/prometheus/discovery/marathon"
	prom_moby "github.com/prometheus/prometheus/discovery/moby"
	prom_openstack "github.com/prometheus/prometheus/discovery/openstack"
	prom_scaleway "github.com/prometheus/prometheus/discovery/scaleway"
	prom_triton "github.com/prometheus/prometheus/discovery/triton"
	prom_kuma "github.com/prometheus/prometheus/discovery/xds"
	prom_zk "github.com/prometheus/prometheus/discovery/zookeeper"
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

	diags.AddAll(common.UnsupportedNotEquals(globalConfig.EvaluationInterval, prom_config.DefaultGlobalConfig.EvaluationInterval, "global evaluation_interval"))
	diags.AddAll(common.UnsupportedNotEquals(globalConfig.QueryLogFile, "", "global query_log_file"))

	return diags
}

func validateAlertingConfig(alertingConfig *prom_config.AlertingConfig) diag.Diagnostics {
	hasAlerting := len(alertingConfig.AlertmanagerConfigs) > 0 || len(alertingConfig.AlertRelabelConfigs) > 0
	return common.UnsupportedEquals(hasAlerting, true, "alerting")
}

func validateRuleFilesConfig(ruleFilesConfig []string) diag.Diagnostics {
	return common.UnsupportedEquals(len(ruleFilesConfig) > 0, true, "rule_files")
}

func validateScrapeConfigs(scrapeConfigs []*prom_config.ScrapeConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, scrapeConfig := range scrapeConfigs {
		diags.AddAll(validatePrometheusScrape(scrapeConfig))
		diags.AddAll(ValidateServiceDiscoveryConfigs(scrapeConfig.ServiceDiscoveryConfigs))
	}
	return diags
}

func ValidateServiceDiscoveryConfigs(serviceDiscoveryConfigs prom_discover.Configs) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, serviceDiscoveryConfig := range serviceDiscoveryConfigs {
		switch sdc := serviceDiscoveryConfig.(type) {
		case prom_discover.StaticConfig:
			diags.AddAll(validateScrapeTargets(sdc))
		case *prom_azure.SDConfig:
			diags.AddAll(validateDiscoveryAzure(sdc))
		case *prom_consul.SDConfig:
			diags.AddAll(validateDiscoveryConsul(sdc))
		case *prom_digitalocean.SDConfig:
			diags.AddAll(validateDiscoveryDigitalOcean(sdc))
		case *prom_dns.SDConfig:
			diags.AddAll(validateDiscoveryDns(sdc))
		case *prom_moby.DockerSDConfig:
			diags.AddAll(validateDiscoveryDocker(sdc))
		case *prom_aws.EC2SDConfig:
			diags.AddAll(validateDiscoveryEC2(sdc))
		case *prom_file.SDConfig:
			diags.AddAll(validateDiscoveryFile(sdc))
		case *prom_gce.SDConfig:
			diags.AddAll(validateDiscoveryGCE(sdc))
		case *prom_kubernetes.SDConfig:
			diags.AddAll(validateDiscoveryKubernetes(sdc))
		case *prom_aws.LightsailSDConfig:
			diags.AddAll(validateDiscoveryLightsail(sdc))
		case *prom_kuma.SDConfig:
			diags.AddAll(validateDiscoveryKuma(sdc))
		case *prom_linode.SDConfig:
			diags.AddAll(validateDiscoveryLinode(sdc))
		case *prom_triton.SDConfig:
			diags.AddAll(validateDiscoveryTriton(sdc))
		case *prom_scaleway.SDConfig:
			diags.AddAll(validateDiscoveryScaleway(sdc))
		case *prom_marathon.SDConfig:
			diags.AddAll(validateDiscoveryMarathon(sdc))
		case *prom_ionos.SDConfig:
			diags.AddAll(validateDiscoveryIonos(sdc))
		case *prom_zk.ServersetSDConfig:
			diags.AddAll(validateDiscoveryServerset(sdc))
		case *prom_zk.NerveSDConfig:
			diags.AddAll(validateDiscoveryNerve(sdc))
		case *prom_openstack.SDConfig:
			diags.AddAll(validateDiscoveryOpenstack(sdc))
		case *prom_moby.DockerSwarmSDConfig:
			diags.AddAll(validateDiscoveryDockerswarm(sdc))
		default:
			diags.Add(diag.SeverityLevelError, fmt.Sprintf("The converter does not support converting the provided %s service discovery.", serviceDiscoveryConfig.Name()))
		}
	}

	return diags
}

func validateStorageConfig(storageConfig *prom_config.StorageConfig) diag.Diagnostics {
	hasStorage := storageConfig.TSDBConfig != nil || storageConfig.ExemplarsConfig != nil
	return common.UnsupportedEquals(hasStorage, true, "storage")
}

func validateTracingConfig(tracingConfig *prom_config.TracingConfig) diag.Diagnostics {
	return common.UnsupportedNotDeepEquals(*tracingConfig, prom_config.TracingConfig{}, "tracing")
}

func validateRemoteWriteConfigs(remoteWriteConfigs []*prom_config.RemoteWriteConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, remoteWriteConfig := range remoteWriteConfigs {
		diags.AddAll(validateRemoteWriteConfig(remoteWriteConfig))
	}

	return diags
}

func validateRemoteReadConfigs(remoteReadConfigs []*prom_config.RemoteReadConfig) diag.Diagnostics {
	return common.UnsupportedEquals(len(remoteReadConfigs) > 0, true, "remote_read")
}
