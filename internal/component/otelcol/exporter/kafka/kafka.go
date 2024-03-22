// Package kafka provides an otelcol.exporter.kafka component.
package kafka

import (
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/component/otelcol"
	"github.com/grafana/agent/internal/component/otelcol/exporter"
	"github.com/grafana/agent/internal/component/otelcol/receiver/kafka"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:      "otelcol.exporter.kafka",
		Stability: featuregate.StabilityStable,
		Args:      Arguments{},
		Exports:   otelcol.ConsumerExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := kafkaexporter.NewFactory()
			return exporter.New(opts, fact, args.(Arguments), exporter.TypeAll)
		},
	})
}

// Arguments configures the otelcol.exporter.otlp component.
type Arguments struct {
	ProtocolVersion string `river:"protocol_version,attr"`

	Brokers  []string      `river:"brokers,attr,optional"`
	ClientID string        `river:"client_id,attr,optional"`
	Topic    string        `river:"topic,attr,optional"`
	Encoding string        `river:"encoding,attr,optional"`
	Timeout  time.Duration `river:"timeout,attr,optional"`

	Queue        otelcol.QueueArguments        `river:"sending_queue,block,optional"`
	Retry        otelcol.RetryArguments        `river:"retry_on_failure,block,optional"`
	DebugMetrics otelcol.DebugMetricsArguments `river:"debug_metrics,block,optional"`

	Auth     kafka.AuthenticationArguments `river:"authentication,block,optional"`
	Metadata kafka.MetadataArguments       `river:"metadata,block,optional"`
	Producer Producer                      `river:"producer,block,optional"`

	ResolveCanonicalBootstrapServersOnly bool `river:"resolve_canonical_bootstrap_servers_only,attr,optional"`
	PartitionTracesByID                  bool `river:"partition_traces_by_id,attr,optional"`
}

var _ exporter.Arguments = Arguments{}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = Arguments{
		Brokers:  []string{"localhost:9092"},
		ClientID: "sarama",
		Encoding: "otlp_proto",
		Timeout:  otelcol.DefaultTimeout,
	}

	args.Queue.SetToDefault()
	args.Retry.SetToDefault()
	args.DebugMetrics.SetToDefault()
	args.Metadata.SetToDefault()
	args.Producer.SetToDefault()
}

// Convert implements exporter.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	input := make(map[string]interface{})
	input["auth"] = args.Auth.Convert()

	res := &kafkaexporter.Config{}
	err := mapstructure.Decode(input, res)
	if err != nil {
		return nil, err
	}
	res.ProtocolVersion = args.ProtocolVersion

	res.Brokers = args.Brokers
	res.ClientID = args.ClientID
	res.Topic = args.Topic
	res.Encoding = args.Encoding
	res.TimeoutSettings = exporterhelper.TimeoutSettings{
		Timeout: args.Timeout,
	}

	res.QueueSettings = *args.Queue.Convert()
	res.BackOffConfig = *args.Retry.Convert()

	res.Metadata = args.Metadata.Convert()
	res.Producer = args.Producer.Convert()

	res.ResolveCanonicalBootstrapServersOnly = args.ResolveCanonicalBootstrapServersOnly
	res.PartitionTracesByID = args.PartitionTracesByID

	return res, nil
}

// Extensions implements exporter.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return nil
}

// Exporters implements exporter.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// DebugMetricsConfig implements receiver.Arguments.
func (args Arguments) DebugMetricsConfig() otelcol.DebugMetricsArguments {
	return args.DebugMetrics
}

type Producer struct {
	MaxMessageBytes  int    `river:"max_message_bytes,attr,optional"`
	RequiredAcks     int    `river:"required_acks,attr,optional"`
	Compression      string `river:"compression,attr,optional"`
	FlushMaxMessages int    `river:"flush_max_messages,attr,optional"`
}

func (p Producer) Convert() kafkaexporter.Producer {
	return kafkaexporter.Producer{
		MaxMessageBytes:  p.MaxMessageBytes,
		RequiredAcks:     sarama.RequiredAcks(p.RequiredAcks),
		Compression:      p.Compression,
		FlushMaxMessages: p.FlushMaxMessages,
	}
}

func (p *Producer) SetToDefault() {
	*p = Producer{
		MaxMessageBytes:  1_000_000,
		RequiredAcks:     1,
		Compression:      "none",
		FlushMaxMessages: 0,
	}
}

var validCompressions = map[string]struct{}{
	"none":   {},
	"gzip":   {},
	"snappy": {},
	"lz4":    {},
	"zstd":   {},
}

func (p *Producer) Validate() error {
	if _, ok := validCompressions[p.Compression]; !ok {
		return fmt.Errorf("invalid compression value; must be one of 'none', 'gzip', 'snappy', 'lz4', 'zstd'")
	}

	if p.RequiredAcks < -1 || p.RequiredAcks > 1 {
		return fmt.Errorf("required_acks has to be between -1 and 1")
	}

	return nil
}
