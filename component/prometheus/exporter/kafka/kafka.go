package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/kafka_exporter"
	config_util "github.com/prometheus/common/config"
)

var DefaultArguments = Arguments{
	UseSASLHandshake:        true,
	KafkaVersion:            sarama.V2_0_0_0.String(),
	MetadataRefreshInterval: "1m",
	AllowConcurrent:         true,
	MaxOffsets:              1000,
	PruneIntervalSeconds:    30,
	TopicsFilter:            ".*",
	GroupFilter:             ".*",
}

type Arguments struct {
	Instance string `river:"instance,attr,optional"`
	// Address array (host:port) of Kafka server
	KafkaURIs []string `river:"kafka_uris,attr,optional"`

	// Connect using SASL/PLAIN
	UseSASL bool `river:"use_sasl,attr,optional"`

	// Only set this to false if using a non-Kafka SASL proxy
	UseSASLHandshake bool `river:"use_sasl_handshake,attr,optional"`

	// SASL user name
	SASLUsername string `river:"sasl_username,attr,optional"`

	// SASL user password
	SASLPassword config_util.Secret `river:"sasl_password,attr,optional"`

	// The SASL SCRAM SHA algorithm sha256 or sha512 as mechanism
	SASLMechanism string `river:"sasl_mechanism,attr,optional"`

	// Connect using TLS
	UseTLS bool `river:"use_tls,attr,optional"`

	// The optional certificate authority file for TLS client authentication
	CAFile string `river:"ca_file,attr,optional"`

	// The optional certificate file for TLS client authentication
	CertFile string `river:"cert_file,attr,optional"`

	// The optional key file for TLS client authentication
	KeyFile string `river:"key_file,attr,optional"`

	// If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
	InsecureSkipVerify bool `river:"insecure_skip_verify,attr,optional"`

	// Kafka broker version
	KafkaVersion string `river:"kafka_version,attr,optional"`

	// if you need to use a group from zookeeper
	UseZooKeeperLag bool `river:"use_zookeeper_lag,attr,optional"`

	// Address array (hosts) of zookeeper server.
	ZookeeperURIs []string `river:"zookeeper_uris,attr,optional"`

	// Kafka cluster name
	ClusterName string `river:"kafka_cluster_name,attr,optional"`

	// Metadata refresh interval
	MetadataRefreshInterval string `river:"metadata_refresh_interval,attr,optional"`

	// If true, all scrapes will trigger kafka operations otherwise, they will share results. WARN: This should be disabled on large clusters
	AllowConcurrent bool `river:"allow_concurrency,attr,optional"`

	// Maximum number of offsets to store in the interpolation table for a partition
	MaxOffsets int `river:"max_offsets,attr,optional"`

	// How frequently should the interpolation table be pruned, in seconds
	PruneIntervalSeconds int `river:"prune_interval_seconds,attr,optional"`

	// Regex filter for topics to be monitored
	TopicsFilter string `river:"topics_filter_regex,attr,optional"`

	// Regex filter for consumer groups to be monitored
	GroupFilter string `river:"groups_filter_regex,attr,optional"`
}

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.kafka",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.NewWithTargetBuilder(createIntegration, "kafka", customizeTarget),
	})
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

// Validate implements river.Validator.
func (a *Arguments) Validate() error {
	if a.Instance == "" && len(a.KafkaURIs) > 1 {
		return fmt.Errorf("an automatic value for `instance` cannot be determined from %d kafka servers, manually provide one for this component", len(a.KafkaURIs))
	}
	return nil
}

func customizeTarget(baseTarget discovery.Target, args component.Arguments) []discovery.Target {
	a := args.(Arguments)
	target := baseTarget
	if len(a.KafkaURIs) > 1 {
		target["instance"] = a.Instance
	} else {
		target["instance"] = a.KafkaURIs[0]
	}
	return []discovery.Target{target}
}

func createIntegration(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

func (a *Arguments) Convert() *kafka_exporter.Config {
	return &kafka_exporter.Config{
		KafkaURIs:               a.KafkaURIs,
		UseSASL:                 a.UseSASL,
		UseSASLHandshake:        a.UseSASLHandshake,
		SASLUsername:            a.SASLUsername,
		SASLPassword:            a.SASLPassword,
		SASLMechanism:           a.SASLMechanism,
		UseTLS:                  a.UseTLS,
		CAFile:                  a.CAFile,
		CertFile:                a.CertFile,
		KeyFile:                 a.KeyFile,
		InsecureSkipVerify:      a.InsecureSkipVerify,
		KafkaVersion:            a.KafkaVersion,
		UseZooKeeperLag:         a.UseZooKeeperLag,
		ZookeeperURIs:           a.ZookeeperURIs,
		ClusterName:             a.ClusterName,
		MetadataRefreshInterval: a.MetadataRefreshInterval,
		AllowConcurrent:         a.AllowConcurrent,
		MaxOffsets:              a.MaxOffsets,
		PruneIntervalSeconds:    a.PruneIntervalSeconds,
		TopicsFilter:            a.TopicsFilter,
		GroupFilter:             a.GroupFilter,
	}
}
