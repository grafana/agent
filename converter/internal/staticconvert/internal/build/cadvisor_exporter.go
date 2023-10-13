package build

import (
	"fmt"
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/cadvisor"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	cadvisor_integration "github.com/grafana/agent/pkg/integrations/cadvisor"
)

func (b *IntegrationsV1ConfigBuilder) appendCadvisorExporter(config *cadvisor_integration.Config) discovery.Exports {
	args := toCadvisorExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "cadvisor"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.cadvisor.%s.targets", compLabel))
}

func toCadvisorExporter(config *cadvisor_integration.Config) *cadvisor.Arguments {
	return &cadvisor.Arguments{

		StoreContainerLabels:       config.StoreContainerLabels,
		AllowlistedContainerLabels: config.AllowlistedContainerLabels,
		EnvMetadataAllowlist:       config.EnvMetadataAllowlist,
		RawCgroupPrefixAllowlist:   config.RawCgroupPrefixAllowlist,
		PerfEventsConfig:           config.PerfEventsConfig,
		ResctrlInterval:            time.Duration(config.ResctrlInterval),
		DisabledMetrics:            config.DisabledMetrics,
		EnabledMetrics:             config.EnabledMetrics,
		StorageDuration:            config.StorageDuration,
		ContainerdHost:             config.Containerd,
		ContainerdNamespace:        config.ContainerdNamespace,
		DockerHost:                 config.Docker,
		UseDockerTLS:               config.DockerTLS,
		DockerTLSCert:              config.DockerTLSCert,
		DockerTLSKey:               config.DockerTLSKey,
		DockerTLSCA:                config.DockerTLSCA,
		DockerOnly:                 config.DockerOnly,
	}
}
