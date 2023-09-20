package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/consul"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/consul_exporter"
)

func (b *IntegrationsV1ConfigBuilder) appendConsulExporter(config *consul_exporter.Config) discovery.Exports {
	args := toConsulExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "consul"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.consul.%s.targets", compLabel))
}

func toConsulExporter(config *consul_exporter.Config) *consul.Arguments {
	return &consul.Arguments{
		Server:             config.Server,
		CAFile:             config.CAFile,
		CertFile:           config.CertFile,
		KeyFile:            config.KeyFile,
		ServerName:         config.ServerName,
		Timeout:            config.Timeout,
		InsecureSkipVerify: config.InsecureSkipVerify,
		RequestLimit:       config.RequestLimit,
		AllowStale:         config.AllowStale,
		RequireConsistent:  config.RequireConsistent,
		KVPrefix:           config.KVPrefix,
		KVFilter:           config.KVFilter,
		HealthSummary:      config.HealthSummary,
	}
}
