package prometheusconvert

import (
	"bytes"
	"fmt"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/token/builder"
	promconfig "github.com/prometheus/prometheus/config"
	promdiscover "github.com/prometheus/prometheus/discovery"
	promaws "github.com/prometheus/prometheus/discovery/aws"
	promazure "github.com/prometheus/prometheus/discovery/azure"
	promconsul "github.com/prometheus/prometheus/discovery/consul"
	promdigitalocean "github.com/prometheus/prometheus/discovery/digitalocean"
	promdns "github.com/prometheus/prometheus/discovery/dns"
	promgce "github.com/prometheus/prometheus/discovery/gce"
	promkubernetes "github.com/prometheus/prometheus/discovery/kubernetes"
	promdocker "github.com/prometheus/prometheus/discovery/moby"
	"github.com/prometheus/prometheus/storage"

	_ "github.com/prometheus/prometheus/discovery/install" // Register Prometheus SDs
)

// Convert implements a Prometheus config converter.
func Convert(in []byte) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	promConfig, err := promconfig.Load(string(in), false, log.NewNopLogger())
	if err != nil {
		diags.Add(diag.SeverityLevelError, fmt.Sprintf("failed to parse Prometheus config: %s", err))
		return nil, diags
	}

	f := builder.NewFile()
	diags = AppendAll(f, promConfig)

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		diags.Add(diag.SeverityLevelError, fmt.Sprintf("failed to render Flow config: %s", err.Error()))
		return nil, diags
	}

	if len(buf.Bytes()) == 0 {
		return nil, diags
	}

	prettyByte, newDiags := common.PrettyPrint(buf.Bytes())
	diags = append(diags, newDiags...)
	return prettyByte, diags
}

// AppendAll analyzes the entire prometheus config in memory and transforms it
// into Flow Arguments. It then appends each argument to the file builder.
// Exports from other components are correctly referenced to build the Flow
// pipeline.
func AppendAll(f *builder.File, promConfig *promconfig.Config) diag.Diagnostics {
	var diags diag.Diagnostics
	remoteWriteExports := appendPrometheusRemoteWrite(f, promConfig)

	remoteWriteForwardTo := []storage.Appendable{remoteWriteExports.Receiver}
	scrapeForwardTo := remoteWriteForwardTo
	for _, scrapeConfig := range promConfig.ScrapeConfigs {
		promMetricsRelabelExports := appendPrometheusRelabel(f, scrapeConfig.MetricRelabelConfigs, remoteWriteForwardTo, scrapeConfig.JobName)
		if promMetricsRelabelExports != nil {
			scrapeForwardTo = []storage.Appendable{promMetricsRelabelExports.Receiver}
		}

		scrapeTargets, discoveryTargets, newDiags := appendServiceDiscoveryConfigs(f, scrapeConfig.ServiceDiscoveryConfigs, scrapeConfig.JobName)
		diags = append(diags, newDiags...)

		promDiscoveryRelabelExports := appendDiscoveryRelabel(f, scrapeConfig.RelabelConfigs, scrapeConfig.JobName, discoveryTargets)
		if promDiscoveryRelabelExports != nil {
			scrapeTargets = append(scrapeTargets, promDiscoveryRelabelExports.Output...)
		} else {
			scrapeTargets = append(scrapeTargets, discoveryTargets...)
		}

		appendPrometheusScrape(f, scrapeConfig, scrapeForwardTo, scrapeTargets)
	}

	return diags
}

// appendServiceDiscoveryConfigs will loop through the service discovery
// configs and append them to the file. This returns the scrape targets
// and discovery targets as a result.
func appendServiceDiscoveryConfigs(f *builder.File, serviceDiscoveryConfig promdiscover.Configs, label string) ([]discovery.Target, []discovery.Target, diag.Diagnostics) {
	var diags diag.Diagnostics
	var scrapeTargets []discovery.Target
	var discoveryTargets []discovery.Target
	labelCounts := make(map[string]int)
	for _, serviceDiscoveryConfig := range serviceDiscoveryConfig {
		var exports discovery.Exports
		var newDiags diag.Diagnostics
		switch sdc := serviceDiscoveryConfig.(type) {
		case promdiscover.StaticConfig:
			scrapeTargets = append(scrapeTargets, getScrapeTargets(sdc)...)
		case *promazure.SDConfig:
			labelCounts["azure"]++
			exports, newDiags = appendDiscoveryAzure(f, common.GetUniqueLabel(label, labelCounts["azure"]), sdc)
		case *promconsul.SDConfig:
			labelCounts["consul"]++
			exports, newDiags = appendDiscoveryConsul(f, common.GetUniqueLabel(label, labelCounts["consul"]), sdc)
		case *promdigitalocean.SDConfig:
			labelCounts["digitalocean"]++
			exports, newDiags = appendDiscoveryDigitalOcean(f, common.GetUniqueLabel(label, labelCounts["digitalocean"]), sdc)
		case *promdns.SDConfig:
			labelCounts["dns"]++
			exports = appendDiscoveryDns(f, common.GetUniqueLabel(label, labelCounts["dns"]), sdc)
		case *promdocker.DockerSDConfig:
			labelCounts["docker"]++
			exports, newDiags = appendDiscoveryDocker(f, common.GetUniqueLabel(label, labelCounts["docker"]), sdc)
		case *promaws.EC2SDConfig:
			labelCounts["ec2"]++
			exports, newDiags = appendDiscoveryEC2(f, common.GetUniqueLabel(label, labelCounts["ec2"]), sdc)
		case *promgce.SDConfig:
			labelCounts["gce"]++
			exports = appendDiscoveryGCE(f, common.GetUniqueLabel(label, labelCounts["gce"]), sdc)
		case *promkubernetes.SDConfig:
			labelCounts["kubernetes"]++
			exports, newDiags = appendDiscoveryKubernetes(f, common.GetUniqueLabel(label, labelCounts["kubernetes"]), sdc)
		case *promaws.LightsailSDConfig:
			labelCounts["lightsail"]++
			exports, newDiags = appendDiscoveryLightsail(f, common.GetUniqueLabel(label, labelCounts["lightsail"]), sdc)
		default:
			diags.Add(diag.SeverityLevelWarn, fmt.Sprintf("unsupported service discovery %s was provided", serviceDiscoveryConfig.Name()))
		}

		diags = append(diags, newDiags...)
		discoveryTargets = append(discoveryTargets, exports.Targets...)
	}

	return scrapeTargets, discoveryTargets, diags
}
