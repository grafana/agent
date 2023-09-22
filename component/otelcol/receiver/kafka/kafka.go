// Package kafka provides an otelcol.receiver.kafka component.
package kafka

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/receiver"
	otel_service "github.com/grafana/agent/service/otel"
	"github.com/grafana/river/rivertypes"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:          "otelcol.receiver.kafka",
		Args:          Arguments{},
		NeedsServices: []string{otel_service.ServiceName},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := kafkareceiver.NewFactory()
			return receiver.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.receiver.kafka component.
type Arguments struct {
	Brokers         []string `river:"brokers,attr"`
	ProtocolVersion string   `river:"protocol_version,attr"`
	Topic           string   `river:"topic,attr,optional"`
	Encoding        string   `river:"encoding,attr,optional"`
	GroupID         string   `river:"group_id,attr,optional"`
	ClientID        string   `river:"client_id,attr,optional"`
	InitialOffset   string   `river:"initial_offset,attr,optional"`

	Authentication AuthenticationArguments `river:"authentication,block,optional"`
	Metadata       MetadataArguments       `river:"metadata,block,optional"`
	AutoCommit     AutoCommitArguments     `river:"autocommit,block,optional"`
	MessageMarking MessageMarkingArguments `river:"message_marking,block,optional"`

	// DebugMetrics configures component internal metrics. Optional.
	DebugMetrics otelcol.DebugMetricsArguments `river:"debug_metrics,block,optional"`

	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var _ receiver.Arguments = Arguments{}

// DefaultArguments holds default values for Arguments.
var DefaultArguments = Arguments{
	// We use the defaults from the upstream OpenTelemetry Collector component
	// for compatibility, even though that means using a client and group ID of
	// "otel-collector".

	Topic:         "otlp_spans",
	Encoding:      "otlp_proto",
	Brokers:       []string{"localhost:9092"},
	ClientID:      "otel-collector",
	GroupID:       "otel-collector",
	InitialOffset: "latest",
	Metadata: MetadataArguments{
		IncludeAllTopics: true,
		Retry: MetadataRetryArguments{
			MaxRetries: 3,
			Backoff:    250 * time.Millisecond,
		},
	},
	AutoCommit: AutoCommitArguments{
		Enable:   true,
		Interval: time.Second,
	},
	MessageMarking: MessageMarkingArguments{
		AfterExecution:      false,
		IncludeUnsuccessful: false,
	},
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Convert implements receiver.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	return &kafkareceiver.Config{
		Brokers:         args.Brokers,
		ProtocolVersion: args.ProtocolVersion,
		Topic:           args.Topic,
		Encoding:        args.Encoding,
		GroupID:         args.GroupID,
		ClientID:        args.ClientID,
		InitialOffset:   args.InitialOffset,

		Authentication: args.Authentication.Convert(),
		Metadata:       args.Metadata.Convert(),
		AutoCommit:     args.AutoCommit.Convert(),
		MessageMarking: args.MessageMarking.Convert(),
	}, nil
}

// Extensions implements receiver.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return nil
}

// Exporters implements receiver.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// NextConsumers implements receiver.Arguments.
func (args Arguments) NextConsumers() *otelcol.ConsumerArguments {
	return args.Output
}

// AuthenticationArguments configures how to authenticate to the Kafka broker.
type AuthenticationArguments struct {
	Plaintext *PlaintextArguments         `river:"plaintext,block,optional"`
	SASL      *SASLArguments              `river:"sasl,block,optional"`
	TLS       *otelcol.TLSClientArguments `river:"tls,block,optional"`
	Kerberos  *KerberosArguments          `river:"kerberos,block,optional"`
}

// Convert converts args into the upstream type.
func (args AuthenticationArguments) Convert() kafkaexporter.Authentication {
	var res kafkaexporter.Authentication

	if args.Plaintext != nil {
		conv := args.Plaintext.Convert()
		res.PlainText = &conv
	}
	if args.SASL != nil {
		conv := args.SASL.Convert()
		res.SASL = &conv
	}
	if args.TLS != nil {
		res.TLS = args.TLS.Convert()
	}
	if args.Kerberos != nil {
		conv := args.Kerberos.Convert()
		res.Kerberos = &conv
	}

	return res
}

// PlaintextArguments configures plaintext authentication against the Kafka
// broker.
type PlaintextArguments struct {
	Username string            `river:"username,attr"`
	Password rivertypes.Secret `river:"password,attr"`
}

// Convert converts args into the upstream type.
func (args PlaintextArguments) Convert() kafkaexporter.PlainTextConfig {
	return kafkaexporter.PlainTextConfig{
		Username: args.Username,
		Password: string(args.Password),
	}
}

// SASLArguments configures SASL authentication against the Kafka broker.
type SASLArguments struct {
	Username  string            `river:"username,attr"`
	Password  rivertypes.Secret `river:"password,attr"`
	Mechanism string            `river:"mechanism,attr"`
	AWSMSK    AWSMSKArguments   `river:"aws_msk,block,optional"`
}

// Convert converts args into the upstream type.
func (args SASLArguments) Convert() kafkaexporter.SASLConfig {
	return kafkaexporter.SASLConfig{
		Username:  args.Username,
		Password:  string(args.Password),
		Mechanism: args.Mechanism,
		AWSMSK:    args.AWSMSK.Convert(),
	}
}

// AWSMSKArguments exposes additional SASL authentication measures required to
// use the AWS_MSK_IAM mechanism.
type AWSMSKArguments struct {
	Region     string `river:"region,attr"`
	BrokerAddr string `river:"broker_addr,attr"`
}

// Convert converts args into the upstream type.
func (args AWSMSKArguments) Convert() kafkaexporter.AWSMSKConfig {
	return kafkaexporter.AWSMSKConfig{
		Region:     args.Region,
		BrokerAddr: args.BrokerAddr,
	}
}

// KerberosArguments configures Kerberos authentication against the Kafka
// broker.
type KerberosArguments struct {
	ServiceName string            `river:"service_name,attr,optional"`
	Realm       string            `river:"realm,attr,optional"`
	UseKeyTab   bool              `river:"use_keytab,attr,optional"`
	Username    string            `river:"username,attr"`
	Password    rivertypes.Secret `river:"password,attr,optional"`
	ConfigPath  string            `river:"config_file,attr,optional"`
	KeyTabPath  string            `river:"keytab_file,attr,optional"`
}

// Convert converts args into the upstream type.
func (args KerberosArguments) Convert() kafkaexporter.KerberosConfig {
	return kafkaexporter.KerberosConfig{
		ServiceName: args.ServiceName,
		Realm:       args.Realm,
		UseKeyTab:   args.UseKeyTab,
		Username:    args.Username,
		Password:    string(args.Password),
		ConfigPath:  args.ConfigPath,
		KeyTabPath:  args.KeyTabPath,
	}
}

// MetadataArguments configures how the otelcol.receiver.kafka component will
// retrieve metadata from the Kafka broker.
type MetadataArguments struct {
	IncludeAllTopics bool                   `river:"include_all_topics,attr,optional"`
	Retry            MetadataRetryArguments `river:"retry,block,optional"`
}

// Convert converts args into the upstream type.
func (args MetadataArguments) Convert() kafkaexporter.Metadata {
	return kafkaexporter.Metadata{
		Full:  args.IncludeAllTopics,
		Retry: args.Retry.Convert(),
	}
}

// MetadataRetryArguments configures how to retry retrieving metadata from the
// Kafka broker. Retrying is useful to avoid race conditions when the Kafka
// broker is starting at the same time as the otelcol.receiver.kafka component.
type MetadataRetryArguments struct {
	MaxRetries int           `river:"max_retries,attr,optional"`
	Backoff    time.Duration `river:"backoff,attr,optional"`
}

// Convert converts args into the upstream type.
func (args MetadataRetryArguments) Convert() kafkaexporter.MetadataRetry {
	return kafkaexporter.MetadataRetry{
		Max:     args.MaxRetries,
		Backoff: args.Backoff,
	}
}

// AutoCommitArguments configures how to automatically commit updated topic
// offsets back to the Kafka broker.
type AutoCommitArguments struct {
	Enable   bool          `river:"enable,attr,optional"`
	Interval time.Duration `river:"interval,attr,optional"`
}

// Convert converts args into the upstream type.
func (args AutoCommitArguments) Convert() kafkareceiver.AutoCommit {
	return kafkareceiver.AutoCommit{
		Enable:   args.Enable,
		Interval: args.Interval,
	}
}

// MessageMarkingArguments configures when Kafka messages are marked as read.
type MessageMarkingArguments struct {
	AfterExecution      bool `river:"after_execution,attr,optional"`
	IncludeUnsuccessful bool `river:"include_unsuccessful,attr,optional"`
}

// Convert converts args into the upstream type.
func (args MessageMarkingArguments) Convert() kafkareceiver.MessageMarking {
	return kafkareceiver.MessageMarking{
		After:   args.AfterExecution,
		OnError: args.IncludeUnsuccessful,
	}
}

// DebugMetricsConfig implements receiver.Arguments.
func (args Arguments) DebugMetricsConfig() otelcol.DebugMetricsArguments {
	return args.DebugMetrics
}
