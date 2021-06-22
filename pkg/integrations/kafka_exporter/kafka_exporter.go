package kafka_exporter //nolint:golint

import (
	"fmt"

	"github.com/Shopify/sarama"
	kafka_exporter "github.com/davidmparrott/kafka_exporter/v2/exporter"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
)

// DefaultConfig holds the default settings for the kafka_lag_exporter
// integration.
var DefaultConfig = Config{
	UseSASLHandshake:        true,
	KafkaVersion:            sarama.V2_0_0_0.String(),
	MetadataRefreshInterval: "1m",
	AllowConcurrent:         true,
	MaxOffsets:              1000,
	PruneIntervalSeconds:    30,
	TopicsFilter:            ".*",
	GroupFilter:             ".*",
}

// Config controls kafka_exporter
type Config struct {
	Common config.Common `yaml:",inline"`

	// Address array (host:port) of Kafka server
	KafkaUri []string `yaml:"kafka_uris,omitempty"`

	// Connect using SASL/PLAIN
	UseSASL bool `yaml:"use_sasl,omitempty"`

	// Only set this to false if using a non-Kafka SASL proxy
	UseSASLHandshake bool `yaml:"use_sasl_handshake,omitempty"`

	// SASL user name
	SASLUsername string `yaml:"sasl_username,omitempty"`

	// SASL user password
	SASLPassword string `yaml:"sasl_password,omitempty"`

	// The SASL SCRAM SHA algorithm sha256 or sha512 as mechanism
	SASLMechanism string `yaml:"sasl_mechanism,omitempty"`

	// Connect using TLS
	UseTLS bool `yaml:"use_tls,omitempty"`

	// The optional certificate authority file for TLS client authentication
	TLSCAFile string `yaml:"tls_cafile,omitempty"`

	// The optional certificate file for TLS client authentication
	TLSCertFile string `yam:"tls_certfile,omitempty"`

	// The optional key file for TLS client authentication
	TLSKeyFile string `yaml:"tls_keyfile,omitempty"`

	// If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
	TLSInsecureSkipTLSVerify bool `yaml:"tls_insecure_skip_tlsverify,omitempty"`

	// Kafka broker version
	KafkaVersion string `yaml:"kafka_version,omitempty"`

	// if you need to use a group from zookeeper
	UseZooKeeperLag bool `yaml:"use_zookeeper_lag,omitempty"`

	// Address array (hosts) of zookeeper server.
	UriZookeeper []string `yaml:"zookeeper_uris,omitempty"`

	// Kafka cluster name
	Labels string `yaml:"kafka_cluster_name,omitempty"`

	// Metadata refresh interval
	MetadataRefreshInterval string `yaml:"metadata_refresh_interval,omitempty"`

	// If true, all scrapes will trigger kafka operations otherwise, they will share results. WARN: This should be disabled on large clusters
	AllowConcurrent bool `yaml:"allow_concurrency,omitempty"`

	// Maximum number of offsets to store in the interpolation table for a partition
	MaxOffsets int `yaml:"max_offsets,omitempty"`

	// How frequently should the interpolation table be pruned, in seconds
	PruneIntervalSeconds int `yaml:"prune_interval_seconds,omitempty"`

	// Regex filter for topics to be monitored
	TopicsFilter string `yaml:"topics_filter_regex,omitempty"`

	// Regex filter for consumer groups to be monitored
	GroupFilter string `yaml:"groups_filter_regex,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "kafka_exporter"
}

// CommonConfig returns the common settings shared across all configs for
// integrations.
func (c *Config) CommonConfig() config.Common {
	return c.Common
}

// NewIntegration creates a new elasticsearch_exporter
func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	return New(logger, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// New creates a new kafka_exporter integration.
func New(logger log.Logger, c *Config) (integrations.Integration, error) {
	if len(c.KafkaUri) == 0 || c.KafkaUri[0] == "" {
		return nil, fmt.Errorf("empty kafka_uris provided")
	}
	if c.UseTLS && (c.TLSCertFile == "" || c.TLSKeyFile == "") {
		return nil, fmt.Errorf("tls is enabled but key pair was not provided")
	}
	if c.UseSASL && (c.SASLPassword == "" || c.SASLUsername == "") {
		return nil, fmt.Errorf("SASL is enabled but username or password was not provided")
	}
	if c.UseZooKeeperLag && (len(c.UriZookeeper) == 0 || c.UriZookeeper[0] == "") {
		return nil, fmt.Errorf("zookeeper lag is enabled but no zookeeper uri was provided")
	}

	var options kafka_exporter.Options
	options.Uri = c.KafkaUri
	options.UseSASL = c.UseSASL
	options.UseSASLHandshake = c.UseSASLHandshake
	options.SaslUsername = c.SASLUsername
	options.SaslPassword = c.SASLPassword
	options.SaslMechanism = c.SASLMechanism
	options.UseTLS = c.UseTLS
	options.TlsCAFile = c.TLSCAFile
	options.TlsCertFile = c.TLSCertFile
	options.TlsKeyFile = c.TLSKeyFile
	options.TlsInsecureSkipTLSVerify = c.TLSInsecureSkipTLSVerify
	options.KafkaVersion = c.KafkaVersion
	options.UseZooKeeperLag = c.UseZooKeeperLag
	options.UriZookeeper = c.UriZookeeper
	options.Labels = c.Labels
	options.MetadataRefreshInterval = c.MetadataRefreshInterval
	options.AllowConcurrent = c.AllowConcurrent
	options.MaxOffsets = c.MaxOffsets
	options.PruneIntervalSeconds = c.PruneIntervalSeconds

	newExporter, err := kafka_exporter.New(logger, options, c.TopicsFilter, c.GroupFilter)
	if err != nil {
		return nil, fmt.Errorf("could not instanciate kafka lag exporter: %w", err)
	}

	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(
			newExporter,
		),
	), nil

}
