package prometheusconvert

import (
	"bytes"
	"fmt"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/token/builder"
	prom_config "github.com/prometheus/prometheus/config"
	prom_discover "github.com/prometheus/prometheus/discovery"
	prom_aws "github.com/prometheus/prometheus/discovery/aws"
	prom_azure "github.com/prometheus/prometheus/discovery/azure"
	prom_consul "github.com/prometheus/prometheus/discovery/consul"
	prom_digitalocean "github.com/prometheus/prometheus/discovery/digitalocean"
	prom_dns "github.com/prometheus/prometheus/discovery/dns"
	prom_file "github.com/prometheus/prometheus/discovery/file"
	prom_gce "github.com/prometheus/prometheus/discovery/gce"
	prom_kubernetes "github.com/prometheus/prometheus/discovery/kubernetes"
	prom_docker "github.com/prometheus/prometheus/discovery/moby"
	"github.com/prometheus/prometheus/storage"

	_ "github.com/prometheus/prometheus/discovery/install" // Register Prometheus SDs
)

// Convert implements a Prometheus config converter.
func Convert(in []byte) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	promConfig, err := prom_config.Load(string(in), false, log.NewNopLogger())
	if err != nil {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to parse Prometheus config: %s", err))
		return nil, diags
	}

	f := builder.NewFile()
	diags = AppendAll(f, promConfig, "")

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to render Flow config: %s", err.Error()))
		return nil, diags
	}

	if len(buf.Bytes()) == 0 {
		return nil, diags
	}

	prettyByte, newDiags := common.PrettyPrint(buf.Bytes())
	diags.AddAll(newDiags)
	return prettyByte, diags
}

// AppendAll analyzes the entire prometheus config in memory and transforms it
// into Flow Arguments. It then appends each argument to the file builder.
// Exports from other components are correctly referenced to build the Flow
// pipeline. A non-empty labelPrefix can be provided for label uniqueness when
// calling this function for the same builder.File multiple times.
func AppendAll(f *builder.File, promConfig *prom_config.Config, labelPrefix string) diag.Diagnostics {
	pb := newPrometheusBlocks()

	remoteWriteExports := appendPrometheusRemoteWrite(pb, promConfig.GlobalConfig, promConfig.RemoteWriteConfigs, labelPrefix)
	remoteWriteForwardTo := []storage.Appendable{remoteWriteExports.Receiver}

	for _, scrapeConfig := range promConfig.ScrapeConfigs {
		scrapeForwardTo := remoteWriteForwardTo
		label := scrapeConfig.JobName
		if labelPrefix != "" {
			label = labelPrefix + "_" + label
		}
		promMetricsRelabelExports := appendPrometheusRelabel(pb, scrapeConfig.MetricRelabelConfigs, remoteWriteForwardTo, label)
		if promMetricsRelabelExports != nil {
			scrapeForwardTo = []storage.Appendable{promMetricsRelabelExports.Receiver}
		}

		scrapeTargets := appendServiceDiscoveryConfigs(pb, scrapeConfig.ServiceDiscoveryConfigs, label)

		promDiscoveryRelabelExports := appendDiscoveryRelabel(pb, scrapeConfig.RelabelConfigs, scrapeTargets, label)
		if promDiscoveryRelabelExports != nil {
			scrapeTargets = promDiscoveryRelabelExports.Output
		}

		appendPrometheusScrape(pb, scrapeConfig, scrapeForwardTo, scrapeTargets, label)
	}

	diags := validate(promConfig)
	diags.AddAll(pb.getScrapeInfo())

	pb.appendToFile(f)

	return diags
}

// appendServiceDiscoveryConfigs will loop through the service discovery
// configs and append them to the file. This returns the scrape targets
// and discovery targets as a result.
func appendServiceDiscoveryConfigs(pb *prometheusBlocks, serviceDiscoveryConfig prom_discover.Configs, label string) []discovery.Target {
	var targets []discovery.Target
	labelCounts := make(map[string]int)
	for _, serviceDiscoveryConfig := range serviceDiscoveryConfig {
		var exports discovery.Exports
		switch sdc := serviceDiscoveryConfig.(type) {
		case prom_discover.StaticConfig:
			targets = append(targets, getScrapeTargets(sdc)...)
		case *prom_azure.SDConfig:
			labelCounts["azure"]++
			exports = appendDiscoveryAzure(pb, common.GetLabelWithIndex(labelCounts["azure"]-1, label), sdc)
		case *prom_consul.SDConfig:
			labelCounts["consul"]++
			exports = appendDiscoveryConsul(pb, common.GetLabelWithIndex(labelCounts["consul"]-1, label), sdc)
		case *prom_digitalocean.SDConfig:
			labelCounts["digitalocean"]++
			exports = appendDiscoveryDigitalOcean(pb, common.GetLabelWithIndex(labelCounts["digitalocean"]-1, label), sdc)
		case *prom_dns.SDConfig:
			labelCounts["dns"]++
			exports = appendDiscoveryDns(pb, common.GetLabelWithIndex(labelCounts["dns"]-1, label), sdc)
		case *prom_docker.DockerSDConfig:
			labelCounts["docker"]++
			exports = appendDiscoveryDocker(pb, common.GetLabelWithIndex(labelCounts["docker"]-1, label), sdc)
		case *prom_aws.EC2SDConfig:
			labelCounts["ec2"]++
			exports = appendDiscoveryEC2(pb, common.GetLabelWithIndex(labelCounts["ec2"]-1, label), sdc)
		case *prom_file.SDConfig:
			labelCounts["file"]++
			exports = appendDiscoveryFile(pb, common.GetLabelWithIndex(labelCounts["file"]-1, label), sdc)
		case *prom_gce.SDConfig:
			labelCounts["gce"]++
			exports = appendDiscoveryGCE(pb, common.GetLabelWithIndex(labelCounts["gce"]-1, label), sdc)
		case *prom_kubernetes.SDConfig:
			labelCounts["kubernetes"]++
			exports = appendDiscoveryKubernetes(pb, common.GetLabelWithIndex(labelCounts["kubernetes"]-1, label), sdc)
		case *prom_aws.LightsailSDConfig:
			labelCounts["lightsail"]++
			exports = appendDiscoveryLightsail(pb, common.GetLabelWithIndex(labelCounts["lightsail"]-1, label), sdc)
		}

		targets = append(exports.Targets, targets...)
	}

	return targets
}
