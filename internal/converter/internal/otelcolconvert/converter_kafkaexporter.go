package otelcolconvert

import (
	"fmt"

	"github.com/grafana/agent/internal/component/otelcol/exporter/kafka"
	"github.com/grafana/agent/internal/component/otelcol/exporter/otlp"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	"go.opentelemetry.io/collector/component"
)

func init() {
	converters = append(converters, kafkaExporterConverter{})
}

type kafkaExporterConverter struct{}

func (kafkaExporterConverter) Factory() component.Factory {
	return kafkaexporter.NewFactory()
}

func (kafkaExporterConverter) InputComponentName() string { return "otelcol.exporter.kafka" }

func (kafkaExporterConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	args := toKafkaExporterConfig(cfg.(*kafkaexporter.Config))
	block := common.NewBlockWithOverride([]string{"otelcol", "exporter", "kafka"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toKafkaExporterConfig(cfg *kafkaexporter.Config) *kafka.Arguments {
	return &kafka.Arguments{
		ProtocolVersion: cfg.ProtocolVersion,

		Brokers:  cfg.Brokers,
		ClientID: cfg.ClientID,
		Topic:    cfg.Topic,
		Encoding: cfg.Encoding,
		Timeout:  cfg.Timeout,

		Queue:        toQueueArguments(cfg.QueueSettings),
		Retry:        toRetryArguments(cfg.BackOffConfig),
		DebugMetrics: common.DefaultValue[otlp.Arguments]().DebugMetrics,

		Auth:     toKafkaAuthentication(encodeMapstruct(cfg.Authentication)),
		Metadata: toKafkaMetadata(cfg.Metadata),
		Producer: kafka.Producer{
			MaxMessageBytes:  cfg.Producer.MaxMessageBytes,
			RequiredAcks:     int(cfg.Producer.RequiredAcks),
			Compression:      cfg.Producer.Compression,
			FlushMaxMessages: cfg.Producer.FlushMaxMessages,
		},
		ResolveCanonicalBootstrapServersOnly: cfg.ResolveCanonicalBootstrapServersOnly,
		PartitionTracesByID:                  cfg.PartitionTracesByID,
	}
}
