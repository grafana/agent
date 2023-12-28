package build

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/elasticsearch"
	"github.com/grafana/agent/pkg/integrations/elasticsearch_exporter"
)

func (b *IntegrationsConfigBuilder) appendElasticsearchExporter(config *elasticsearch_exporter.Config, instanceKey *string) discovery.Exports {
	args := toElasticsearchExporter(config)
	return b.appendExporterBlock(args, config.Name(), instanceKey, "elasticsearch")
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
