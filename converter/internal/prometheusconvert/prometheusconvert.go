package prometheusconvert

import (
	"bytes"
	"fmt"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/remotewrite"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/river/token/builder"
	prom_config "github.com/prometheus/prometheus/config"
	prom_discover "github.com/prometheus/prometheus/discovery"
	prom_aws "github.com/prometheus/prometheus/discovery/aws"
	prom_azure "github.com/prometheus/prometheus/discovery/azure"
	prom_consul "github.com/prometheus/prometheus/discovery/consul"
	prom_digitalocean "github.com/prometheus/prometheus/discovery/digitalocean"
	prom_dns "github.com/prometheus/prometheus/discovery/dns"
	prom_file "github.com/prometheus/prometheus/discovery/file"
	prom_gce "github.com/prometheus/prometheus/discovery/gce"
	prom_ionos "github.com/prometheus/prometheus/discovery/ionos"
	prom_kubernetes "github.com/prometheus/prometheus/discovery/kubernetes"
	prom_linode "github.com/prometheus/prometheus/discovery/linode"
	prom_marathon "github.com/prometheus/prometheus/discovery/marathon"
	prom_docker "github.com/prometheus/prometheus/discovery/moby"
	prom_moby "github.com/prometheus/prometheus/discovery/moby"
	prom_openstack "github.com/prometheus/prometheus/discovery/openstack"
	prom_scaleway "github.com/prometheus/prometheus/discovery/scaleway"
	prom_triton "github.com/prometheus/prometheus/discovery/triton"
	prom_xds "github.com/prometheus/prometheus/discovery/xds"
	prom_nerve "github.com/prometheus/prometheus/discovery/zookeeper"
	prom_zk "github.com/prometheus/prometheus/discovery/zookeeper"
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
	diags = AppendAll(f, promConfig)
	diags.AddAll(common.ValidateNodes(f))

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
// pipeline.
func AppendAll(f *builder.File, promConfig *prom_config.Config) diag.Diagnostics {
	return AppendAllNested(f, promConfig, nil, []discovery.Target{}, nil)
}

// AppendAllNested analyzes the entire prometheus config in memory and transforms it
// into Flow Arguments. It then appends each argument to the file builder.
// Exports from other components are correctly referenced to build the Flow
// pipeline. Additional options can be provided overriding the job name, extra
// scrape targets, and predefined remote write exports.
func AppendAllNested(f *builder.File, promConfig *prom_config.Config, jobNameToCompLabelsFunc func(string) string, extraScrapeTargets []discovery.Target, remoteWriteExports *remotewrite.Exports) diag.Diagnostics {
	pb := NewPrometheusBlocks()

	if remoteWriteExports == nil {
		labelPrefix := ""
		if jobNameToCompLabelsFunc != nil {
			labelPrefix = jobNameToCompLabelsFunc("")
			if labelPrefix != "" {
				labelPrefix = common.SanitizeIdentifierPanics(labelPrefix)
			}
		}
		remoteWriteExports = appendPrometheusRemoteWrite(pb, promConfig.GlobalConfig, promConfig.RemoteWriteConfigs, labelPrefix)
	}
	remoteWriteForwardTo := []storage.Appendable{remoteWriteExports.Receiver}

	for _, scrapeConfig := range promConfig.ScrapeConfigs {
		scrapeForwardTo := remoteWriteForwardTo
		label := scrapeConfig.JobName
		if jobNameToCompLabelsFunc != nil {
			label = jobNameToCompLabelsFunc(scrapeConfig.JobName)
		}
		label = common.SanitizeIdentifierPanics(label)

		promMetricsRelabelExports := appendPrometheusRelabel(pb, scrapeConfig.MetricRelabelConfigs, remoteWriteForwardTo, label)
		if promMetricsRelabelExports != nil {
			scrapeForwardTo = []storage.Appendable{promMetricsRelabelExports.Receiver}
		}

		scrapeTargets := AppendServiceDiscoveryConfigs(pb, scrapeConfig.ServiceDiscoveryConfigs, label)
		scrapeTargets = append(scrapeTargets, extraScrapeTargets...)

		promDiscoveryRelabelExports := appendDiscoveryRelabel(pb, scrapeConfig.RelabelConfigs, scrapeTargets, label)
		if promDiscoveryRelabelExports != nil {
			scrapeTargets = promDiscoveryRelabelExports.Output
		}

		appendPrometheusScrape(pb, scrapeConfig, scrapeForwardTo, scrapeTargets, label)
	}

	diags := validate(promConfig)
	diags.AddAll(pb.getScrapeInfo())

	pb.AppendToFile(f)

	return diags
}

// AppendServiceDiscoveryConfigs will loop through the service discovery
// configs and append them to the file. This returns the scrape targets
// and discovery targets as a result.
func AppendServiceDiscoveryConfigs(pb *prometheusBlocks, serviceDiscoveryConfig prom_discover.Configs, label string) []discovery.Target {
	var targets []discovery.Target
	labelCounts := make(map[string]int)
	for _, serviceDiscoveryConfig := range serviceDiscoveryConfig {
		var exports discovery.Exports
		switch sdc := serviceDiscoveryConfig.(type) {
		case prom_discover.StaticConfig:
			targets = append(targets, getScrapeTargets(sdc)...)
		case *prom_azure.SDConfig:
			labelCounts["azure"]++
			exports = appendDiscoveryAzure(pb, common.LabelWithIndex(labelCounts["azure"]-1, label), sdc)
		case *prom_consul.SDConfig:
			labelCounts["consul"]++
			exports = appendDiscoveryConsul(pb, common.LabelWithIndex(labelCounts["consul"]-1, label), sdc)
		case *prom_digitalocean.SDConfig:
			labelCounts["digitalocean"]++
			exports = appendDiscoveryDigitalOcean(pb, common.LabelWithIndex(labelCounts["digitalocean"]-1, label), sdc)
		case *prom_dns.SDConfig:
			labelCounts["dns"]++
			exports = appendDiscoveryDns(pb, common.LabelWithIndex(labelCounts["dns"]-1, label), sdc)
		case *prom_docker.DockerSDConfig:
			labelCounts["docker"]++
			exports = appendDiscoveryDocker(pb, common.LabelWithIndex(labelCounts["docker"]-1, label), sdc)
		case *prom_aws.EC2SDConfig:
			labelCounts["ec2"]++
			exports = appendDiscoveryEC2(pb, common.LabelWithIndex(labelCounts["ec2"]-1, label), sdc)
		case *prom_file.SDConfig:
			labelCounts["file"]++
			exports = appendDiscoveryFile(pb, common.LabelWithIndex(labelCounts["file"]-1, label), sdc)
		case *prom_gce.SDConfig:
			labelCounts["gce"]++
			exports = appendDiscoveryGCE(pb, common.LabelWithIndex(labelCounts["gce"]-1, label), sdc)
		case *prom_kubernetes.SDConfig:
			labelCounts["kubernetes"]++
			exports = appendDiscoveryKubernetes(pb, common.LabelWithIndex(labelCounts["kubernetes"]-1, label), sdc)
		case *prom_aws.LightsailSDConfig:
			labelCounts["lightsail"]++
			exports = appendDiscoveryLightsail(pb, common.LabelWithIndex(labelCounts["lightsail"]-1, label), sdc)
		case *prom_marathon.SDConfig:
			labelCounts["marathon"]++
			exports = appendDiscoveryMarathon(pb, common.LabelWithIndex(labelCounts["marathon"]-1, label), sdc)
		case *prom_ionos.SDConfig:
			labelCounts["ionos"]++
			exports = appendDiscoveryIonos(pb, common.LabelWithIndex(labelCounts["ionos"]-1, label), sdc)
		case *prom_triton.SDConfig:
			labelCounts["triton"]++
			exports = appendDiscoveryTriton(pb, common.LabelWithIndex(labelCounts["triton"]-1, label), sdc)
		case *prom_xds.SDConfig:
			labelCounts["kuma"]++
			exports = appendDiscoveryKuma(pb, common.LabelWithIndex(labelCounts["kuma"]-1, label), sdc)
		case *prom_scaleway.SDConfig:
			labelCounts["scaleway"]++
			exports = appendDiscoveryScaleway(pb, common.LabelWithIndex(labelCounts["scaleway"]-1, label), sdc)
		case *prom_zk.ServersetSDConfig:
			labelCounts["serverset"]++
			exports = appendDiscoveryServerset(pb, common.LabelWithIndex(labelCounts["serverset"]-1, label), sdc)
		case *prom_linode.SDConfig:
			labelCounts["linode"]++
			exports = appendDiscoveryLinode(pb, common.LabelWithIndex(labelCounts["linode"]-1, label), sdc)
		case *prom_nerve.NerveSDConfig:
			labelCounts["nerve"]++
			exports = appendDiscoveryNerve(pb, common.LabelWithIndex(labelCounts["nerve"]-1, label), sdc)
		case *prom_openstack.SDConfig:
			labelCounts["openstack"]++
			exports = appendDiscoveryOpenstack(pb, common.LabelWithIndex(labelCounts["openstack"]-1, label), sdc)
		case *prom_moby.DockerSwarmSDConfig:
			labelCounts["dockerswarm"]++
			exports = appendDiscoveryDockerswarm(pb, common.LabelWithIndex(labelCounts["dockerswarm"]-1, label), sdc)
		}

		targets = append(targets, exports.Targets...)
	}

	return targets
}
