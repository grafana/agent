package kafkatarget

import (
	"github.com/grafana/agent/component/common/config"
)

type KafkaTargetConfig struct {
	Brokers              []string            `river:"brokers,attr"`
	Topics               []string            `river:"topics,attr"`
	GroupID              string              `river:"group_id,attr,optional"`
	Assignor             string              `river:"assignor,attr,optional"`
	Version              string              `river:"version,attr,optional"`
	Authentication       KafkaAuthentication `river:"authentication,block,optional"`
	UseIncomingTimestamp bool                `river:"use_incoming_timestamp,attr,optional"`
	Labels               map[string]string   `river:"labels,attr,optional"`
}

// KafkaAuthentication describe the configuration for authentication with Kafka brokers
type KafkaAuthentication struct {
	Type       string           `river:"type,attr,optional"`
	TLSConfig  config.TLSConfig `river:"tls_config,block,optional"`
	SASLConfig KafkaSASLConfig  `river:"sasl_config,block,optional"`
}

const (
	// KafkaAuthenticationTypeNone represents using no authentication
	KafkaAuthenticationTypeNone = "none"
	// KafkaAuthenticationTypeSSL represents using SSL/TLS to authenticate
	KafkaAuthenticationTypeSSL = "ssl"
	// KafkaAuthenticationTypeSASL represents using SASL to authenticate
	KafkaAuthenticationTypeSASL = "sasl"
)

// KafkaSASLConfig describe the SASL configuration for authentication with Kafka brokers
type KafkaSASLConfig struct {
	Mechanism string           `river:"mechanism,attr,optional"`
	User      string           `river:"user,attr"`
	Password  string           `river:"password,attr"`
	UseTLS    bool             `river:"use_tls,attr,optional"`
	TLSConfig config.TLSConfig `river:"tls_config,block,optional"`
}
