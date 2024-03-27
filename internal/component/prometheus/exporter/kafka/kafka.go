package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/component/prometheus/exporter"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/grafana/agent/internal/static/integrations"
	"github.com/grafana/agent/internal/static/integrations/kafka_exporter"
	"github.com/grafana/river/rivertypes"
	"github.com/prometheus/common/config"
)

var DefaultArguments = Arguments{
	UseSASLHandshake:        true,
	KafkaVersion:            sarama.V2_0_0_0.String(),
	MetadataRefreshInterval: "1m",
	AllowConcurrent:         true,
	MaxOffsets:              1000,
	OffsetShowAll:           true,
	TopicWorkers:            100,
	TopicsFilter:            ".*",
	TopicsExclude:           "^$",
	GroupFilter:             ".*",
	GroupExclude:            "^$",
}

type Arguments struct {
	Instance                string            `river:"instance,attr,optional"`
	KafkaURIs               []string          `river:"kafka_uris,attr,optional"`
	UseSASL                 bool              `river:"use_sasl,attr,optional"`
	UseSASLHandshake        bool              `river:"use_sasl_handshake,attr,optional"`
	SASLUsername            string            `river:"sasl_username,attr,optional"`
	SASLPassword            rivertypes.Secret `river:"sasl_password,attr,optional"`
	SASLMechanism           string            `river:"sasl_mechanism,attr,optional"`
	SASLDisablePAFXFast     bool              `river:"sasl_disable_pafx_fast,attr,optional"`
	UseTLS                  bool              `river:"use_tls,attr,optional"`
	TlsServerName           string            `river:"tls_server_name,attr,optional"`
	CAFile                  string            `river:"ca_file,attr,optional"`
	CertFile                string            `river:"cert_file,attr,optional"`
	KeyFile                 string            `river:"key_file,attr,optional"`
	InsecureSkipVerify      bool              `river:"insecure_skip_verify,attr,optional"`
	KafkaVersion            string            `river:"kafka_version,attr,optional"`
	UseZooKeeperLag         bool              `river:"use_zookeeper_lag,attr,optional"`
	ZookeeperURIs           []string          `river:"zookeeper_uris,attr,optional"`
	ClusterName             string            `river:"kafka_cluster_name,attr,optional"`
	MetadataRefreshInterval string            `river:"metadata_refresh_interval,attr,optional"`
	ServiceName             string            `river:"gssapi_service_name,attr,optional"`
	KerberosConfigPath      string            `river:"gssapi_kerberos_config_path,attr,optional"`
	Realm                   string            `river:"gssapi_realm,attr,optional"`
	KeyTabPath              string            `river:"gssapi_key_tab_path,attr,optional"`
	KerberosAuthType        string            `river:"gssapi_kerberos_auth_type,attr,optional"`
	OffsetShowAll           bool              `river:"offset_show_all,attr,optional"`
	TopicWorkers            int               `river:"topic_workers,attr,optional"`
	AllowConcurrent         bool              `river:"allow_concurrency,attr,optional"`
	AllowAutoTopicCreation  bool              `river:"allow_auto_topic_creation,attr,optional"`
	MaxOffsets              int               `river:"max_offsets,attr,optional"`
	TopicsFilter            string            `river:"topics_filter_regex,attr,optional"`
	TopicsExclude           string            `river:"topics_exclude_regex,attr,optional"`
	GroupFilter             string            `river:"groups_filter_regex,attr,optional"`
	GroupExclude            string            `river:"groups_exclude_regex,attr,optional"`
}

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.exporter.kafka",
		Stability: featuregate.StabilityStable,
		Args:      Arguments{},
		Exports:   exporter.Exports{},

		Build: exporter.NewWithTargetBuilder(createExporter, "kafka", customizeTarget),
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

func createExporter(opts component.Options, args component.Arguments, defaultInstanceKey string) (integrations.Integration, string, error) {
	a := args.(Arguments)
	return integrations.NewIntegrationWithInstanceKey(opts.Logger, a.Convert(), defaultInstanceKey)
}

func (a *Arguments) Convert() *kafka_exporter.Config {
	return &kafka_exporter.Config{
		Instance:                a.Instance,
		KafkaURIs:               a.KafkaURIs,
		UseSASL:                 a.UseSASL,
		UseSASLHandshake:        a.UseSASLHandshake,
		SASLUsername:            a.SASLUsername,
		SASLPassword:            config.Secret(a.SASLPassword),
		SASLMechanism:           a.SASLMechanism,
		SASLDisablePAFXFast:     a.SASLDisablePAFXFast,
		UseTLS:                  a.UseTLS,
		TlsServerName:           a.TlsServerName,
		CAFile:                  a.CAFile,
		CertFile:                a.CertFile,
		KeyFile:                 a.KeyFile,
		InsecureSkipVerify:      a.InsecureSkipVerify,
		KafkaVersion:            a.KafkaVersion,
		UseZooKeeperLag:         a.UseZooKeeperLag,
		ZookeeperURIs:           a.ZookeeperURIs,
		ClusterName:             a.ClusterName,
		MetadataRefreshInterval: a.MetadataRefreshInterval,
		ServiceName:             a.ServiceName,
		KerberosConfigPath:      a.KerberosConfigPath,
		Realm:                   a.Realm,
		KeyTabPath:              a.KeyTabPath,
		KerberosAuthType:        a.KerberosAuthType,
		OffsetShowAll:           a.OffsetShowAll,
		TopicWorkers:            a.TopicWorkers,
		AllowConcurrent:         a.AllowConcurrent,
		AllowAutoTopicCreation:  a.AllowAutoTopicCreation,
		MaxOffsets:              a.MaxOffsets,
		TopicsFilter:            a.TopicsFilter,
		TopicsExclude:           a.TopicsExclude,
		GroupFilter:             a.GroupFilter,
		GroupExclude:            a.GroupExclude,
	}
}
