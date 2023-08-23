package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/kafka"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/kafka_exporter"
)

func (b *IntegrationsV1ConfigBuilder) appendKafkaExporter(config *kafka_exporter.Config) discovery.Exports {
	args := toKafkaExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "kafka"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.kafka.%s.targets", compLabel))
}

func toKafkaExporter(config *kafka_exporter.Config) *kafka.Arguments {
	return &kafka.Arguments{
		KafkaURIs:               config.KafkaURIs,
		UseSASL:                 config.UseSASL,
		UseSASLHandshake:        config.UseSASLHandshake,
		SASLUsername:            config.SASLUsername,
		SASLPassword:            config.SASLPassword,
		SASLMechanism:           config.SASLMechanism,
		UseTLS:                  config.UseTLS,
		CAFile:                  config.CAFile,
		CertFile:                config.CertFile,
		KeyFile:                 config.KeyFile,
		InsecureSkipVerify:      config.InsecureSkipVerify,
		KafkaVersion:            config.KafkaVersion,
		UseZooKeeperLag:         config.UseZooKeeperLag,
		ZookeeperURIs:           config.ZookeeperURIs,
		ClusterName:             config.ClusterName,
		MetadataRefreshInterval: config.MetadataRefreshInterval,
		AllowConcurrent:         config.AllowConcurrent,
		MaxOffsets:              config.MaxOffsets,
		PruneIntervalSeconds:    config.PruneIntervalSeconds,
		TopicsFilter:            config.TopicsFilter,
		GroupFilter:             config.GroupFilter,
	}
}
