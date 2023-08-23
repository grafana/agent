package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/dnsmasq"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/dnsmasq_exporter"
)

func (b *IntegrationsV1ConfigBuilder) appendDnsmasqExporter(config *dnsmasq_exporter.Config) discovery.Exports {
	args := toDnsmasqExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "dnsmasq"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.dnsmasq.%s.targets", compLabel))
}

func toDnsmasqExporter(config *dnsmasq_exporter.Config) *dnsmasq.Arguments {
	return &dnsmasq.Arguments{
		Address:      config.DnsmasqAddress,
		LeasesFile:   config.LeasesPath,
		ExposeLeases: config.ExposeLeases,
	}
}
