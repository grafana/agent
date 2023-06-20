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

	if len(buf.Bytes()) > 0 {
		prettyByte, newDiags := common.PrettyPrint(buf.Bytes())
		diags = append(diags, newDiags...)
		return prettyByte, diags
	}

	return nil, diags
}

// AppendAll analyzes the entire prometheus config in memory and transforms it
// into Flow Arguments. It then appends each argument to the file builder.
func AppendAll(f *builder.File, promConfig *promconfig.Config) diag.Diagnostics {
	var diags diag.Diagnostics
	remoteWriteExports := appendPrometheusRemoteWrite(f, promConfig)

	forwardTo := []storage.Appendable{remoteWriteExports.Receiver}
	for _, scrapeConfig := range promConfig.ScrapeConfigs {
		promMetricsRelabelExports := appendPrometheusRelabel(f, scrapeConfig.MetricRelabelConfigs, forwardTo, scrapeConfig.JobName)
		if promMetricsRelabelExports != nil {
			forwardTo = []storage.Appendable{promMetricsRelabelExports.Receiver}
		}

		var targets []discovery.Target
		var discoveryTargets []discovery.Target
		labelCounts := make(map[string]int)
		for _, serviceDiscoveryConfig := range scrapeConfig.ServiceDiscoveryConfigs {
			var exports discovery.Exports
			var newDiags diag.Diagnostics
			switch sdc := serviceDiscoveryConfig.(type) {
			case promdiscover.StaticConfig:
				targets = append(targets, getScrapeTargets(sdc)...)
			case *promazure.SDConfig:
				labelCounts["azure"]++
				exports, newDiags = appendDiscoveryAzure(f, common.GetUniqueLabel(scrapeConfig.JobName, labelCounts["azure"]), sdc)
			case *promconsul.SDConfig:
				labelCounts["consul"]++
				exports, newDiags = appendDiscoveryConsul(f, common.GetUniqueLabel(scrapeConfig.JobName, labelCounts["consul"]), sdc)
			case *promdigitalocean.SDConfig:
				labelCounts["digitalocean"]++
				exports, newDiags = appendDiscoveryDigitalOcean(f, common.GetUniqueLabel(scrapeConfig.JobName, labelCounts["digitalocean"]), sdc)
			case *promdns.SDConfig:
				labelCounts["dns"]++
				exports = appendDiscoveryDns(f, common.GetUniqueLabel(scrapeConfig.JobName, labelCounts["dns"]), sdc)
			case *promdocker.DockerSDConfig:
				labelCounts["docker"]++
				exports, newDiags = appendDiscoveryDocker(f, common.GetUniqueLabel(scrapeConfig.JobName, labelCounts["docker"]), sdc)
			case *promaws.EC2SDConfig:
				labelCounts["ec2"]++
				exports, newDiags = appendDiscoveryEC2(f, common.GetUniqueLabel(scrapeConfig.JobName, labelCounts["ec2"]), sdc)
			case *promgce.SDConfig:
				labelCounts["gce"]++
				exports = appendDiscoveryGCE(f, common.GetUniqueLabel(scrapeConfig.JobName, labelCounts["gce"]), sdc)
			case *promkubernetes.SDConfig:
				labelCounts["kubernetes"]++
				exports, newDiags = appendDiscoveryKubernetes(f, common.GetUniqueLabel(scrapeConfig.JobName, labelCounts["kubernetes"]), sdc)
			case *promaws.LightsailSDConfig:
				labelCounts["lightsail"]++
				exports, newDiags = appendDiscoveryLightsail(f, common.GetUniqueLabel(scrapeConfig.JobName, labelCounts["lightsail"]), sdc)
			default:
				diags.Add(diag.SeverityLevelWarn, fmt.Sprintf("unsupported service discovery %s was provided", serviceDiscoveryConfig.Name()))
			}

			diags = append(diags, newDiags...)
			discoveryTargets = append(discoveryTargets, exports.Targets...)
		}

		scrapeTargets := append(targets, discoveryTargets...)
		appendPrometheusScrape(f, scrapeConfig, forwardTo, scrapeTargets)
	}

	return diags
}
