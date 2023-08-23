package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/elasticsearch"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/elasticsearch_exporter"
)

func (b *IntegrationsV1ConfigBuilder) appendElasticsearchExporter(config *elasticsearch_exporter.Config) discovery.Exports {
	args := toElasticsearchExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "elasticsearch"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.elasticsearch.%s.targets", compLabel))
}

func toElasticsearchExporter(config *elasticsearch_exporter.Config) *elasticsearch.Arguments {
	return &elasticsearch.Arguments{
		Address:                   config.Address,
		Timeout:                   config.Timeout,
		AllNodes:                  config.AllNodes,
		Node:                      config.Node,
		ExportIndices:             config.ExportIndices,
		ExportIndicesSettings:     config.ExportIndicesSettings,
		ExportClusterSettings:     config.ExportClusterSettings,
		ExportShards:              config.ExportShards,
		IncludeAliases:            config.IncludeAliases,
		ExportSnapshots:           config.ExportSnapshots,
		ExportClusterInfoInterval: config.ExportClusterInfoInterval,
		CA:                        config.CA,
		ClientPrivateKey:          config.ClientPrivateKey,
		ClientCert:                config.ClientCert,
		InsecureSkipVerify:        config.InsecureSkipVerify,
		ExportDataStreams:         config.ExportDataStreams,
		ExportSLM:                 config.ExportSLM,
	}
}
