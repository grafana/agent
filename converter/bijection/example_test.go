package bijection

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/loki/source/kafka"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	common "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

func TestKafkaBijection(t *testing.T) {
	var fromRiverKafka = kafka.Arguments{
		Brokers:  []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		Topics:   []string{"topic1", "topic2", "topic3"},
		GroupID:  "group1",
		Assignor: "assignor1",
		Version:  "version1",
		Authentication: kafka.KafkaAuthentication{
			Type: "type1",
			TLSConfig: config.TLSConfig{
				CA:                 "ca1",
				CAFile:             "cafile1",
				Cert:               "cert1",
				CertFile:           "certfile1",
				Key:                "key1",
				KeyFile:            "keyfile1",
				ServerName:         "servername1",
				InsecureSkipVerify: true,
				MinVersion:         10,
			},
			SASLConfig: kafka.KafkaSASLConfig{
				Mechanism: "mechanism1",
				User:      "user1",
				Password:  "password1",
				UseTLS:    true,
				TLSConfig: config.TLSConfig{
					CA:                 "ca2",
					CAFile:             "cafile2",
					Cert:               "cert2",
					CertFile:           "certfile2",
					Key:                "key2",
					KeyFile:            "keyfile2",
					ServerName:         "servername2",
					InsecureSkipVerify: false,
					MinVersion:         0,
				},
				OAuthConfig: kafka.OAuthConfigConfig{
					TokenProvider: "",
					Scopes:        nil,
				},
			},
		},
		UseIncomingTimestamp: true,
		Labels:               map[string]string{"label1": "value1", "label2": "value2"},
		ForwardTo:            nil,
		RelabelRules:         nil,
	}

	var expectedPromtailKafka = scrapeconfig.KafkaTargetConfig{
		Labels:               model.LabelSet{"label1": "value1", "label2": "value2"},
		UseIncomingTimestamp: true,
		Brokers:              []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		GroupID:              "group1",
		Topics:               []string{"topic1", "topic2", "topic3"},
		Version:              "version1",
		Assignor:             "assignor1",
		Authentication: scrapeconfig.KafkaAuthentication{
			Type: "type1",
			TLSConfig: common.TLSConfig{
				CA:                 "ca1",
				Cert:               "cert1",
				Key:                "key1",
				CAFile:             "cafile1",
				CertFile:           "certfile1",
				KeyFile:            "keyfile1",
				ServerName:         "servername1",
				InsecureSkipVerify: true,
				MinVersion:         10,
				MaxVersion:         0,
			},
			SASLConfig: scrapeconfig.KafkaSASLConfig{
				Mechanism: "mechanism1",
				User:      "user1",
				Password:  flagext.SecretWithValue("password1"),
				UseTLS:    true,
				TLSConfig: common.TLSConfig{
					CA:                 "ca2",
					Cert:               "cert2",
					Key:                "key2",
					CAFile:             "cafile2",
					CertFile:           "certfile2",
					KeyFile:            "keyfile2",
					ServerName:         "servername2",
					InsecureSkipVerify: false,
					MinVersion:         0,
					MaxVersion:         0,
				},
			},
		},
	}

	bj := kafkaBijection()
	testTwoWayConversion(t, bj, fromRiverKafka, expectedPromtailKafka)
}

func kafkaBijection() Bijection[kafka.Arguments, scrapeconfig.KafkaTargetConfig] {
	type A = kafka.Arguments
	type B = scrapeconfig.KafkaTargetConfig

	kafkaBj := &StructBijection[A, B]{}
	BindMatchingField[A, B, []string](kafkaBj, "Brokers")
	BindMatchingField[A, B, []string](kafkaBj, "Topics")
	BindMatchingField[A, B, string](kafkaBj, "GroupID")
	BindMatchingField[A, B, string](kafkaBj, "Assignor")
	BindMatchingField[A, B, string](kafkaBj, "Version")
	BindField(kafkaBj, MatchingNames("Authentication"), kafkaAuthentication())
	BindMatchingField[A, B, bool](kafkaBj, "UseIncomingTimestamp")
	BindField(kafkaBj, MatchingNames("Labels"), labelsBijection())
	//TODO(thampiotr): add relabel rules support
	return kafkaBj
}

func kafkaAuthentication() Bijection[kafka.KafkaAuthentication, scrapeconfig.KafkaAuthentication] {
	type A = kafka.KafkaAuthentication
	type B = scrapeconfig.KafkaAuthentication

	kafkaAuthBj := &StructBijection[A, B]{}
	BindField(kafkaAuthBj, MatchingNames("Type"), Cast[string, scrapeconfig.KafkaAuthenticationType]())
	BindField(kafkaAuthBj, MatchingNames("TLSConfig"), tlsBijection())
	BindField(kafkaAuthBj, MatchingNames("SASLConfig"), saslBijection())
	return kafkaAuthBj
}

func tlsBijection() Bijection[config.TLSConfig, common.TLSConfig] {
	type A = config.TLSConfig
	type B = common.TLSConfig

	tlsBj := &StructBijection[A, B]{}
	BindMatchingField[A, B, string](tlsBj, "CA")
	BindMatchingField[A, B, string](tlsBj, "CAFile")
	BindMatchingField[A, B, string](tlsBj, "Cert")
	BindMatchingField[A, B, string](tlsBj, "CertFile")
	BindField(tlsBj, MatchingNames("Key"), Compose(Cast[rivertypes.Secret, string](), Cast[string, common.Secret]()))
	BindMatchingField[A, B, string](tlsBj, "KeyFile")
	BindMatchingField[A, B, string](tlsBj, "ServerName")
	BindMatchingField[A, B, bool](tlsBj, "InsecureSkipVerify")
	BindField(tlsBj, MatchingNames("MinVersion"), Compose(Cast[config.TLSVersion, uint16](), Cast[uint16, common.TLSVersion]()))
	// TODO(thampiotr): MaxVersion is not supported by prometheus/common/config -
	//                  would need to support some warnings to keep a safe two-way conversion?
	//                  Also, technically this will no longer be a bijection, so a rename would be in order.
	return tlsBj
}

func saslBijection() Bijection[kafka.KafkaSASLConfig, scrapeconfig.KafkaSASLConfig] {
	type A = kafka.KafkaSASLConfig
	type B = scrapeconfig.KafkaSASLConfig
	bi := &StructBijection[A, B]{}
	BindField(bi, MatchingNames("Mechanism"), Cast[string, sarama.SASLMechanism]())
	BindMatchingField[A, B, string](bi, "User")
	BindField(bi, MatchingNames("Password"), flagextSecretBijection())
	BindMatchingField[A, B, bool](bi, "UseTLS")
	BindField(bi, MatchingNames("TLSConfig"), tlsBijection())
	//TODO(thampiotr): OAuthConfig is not supported by prometheus/common/config
	return bi
}

func labelsBijection() Bijection[map[string]string, model.LabelSet] {
	type A = map[string]string
	type B = model.LabelSet
	return &FnBijection[A, B]{
		AtoB: func(a *A, b *B) error {
			*b = model.LabelSet{}
			for k, v := range *a {
				(*b)[model.LabelName(k)] = model.LabelValue(v)
			}
			return nil
		},
		BtoA: func(b *B, a *A) error {
			*a = map[string]string{}
			for k, v := range *b {
				(*a)[string(k)] = string(v)
			}
			return nil
		},
	}
}

func flagextSecretBijection() Bijection[string, flagext.Secret] {
	type A = string
	type B = flagext.Secret
	return &FnBijection[A, B]{
		AtoB: func(a *A, b *B) error {
			*b = flagext.SecretWithValue(*a)
			return nil
		},
		BtoA: func(b *B, a *A) error {
			*a = b.String()
			return nil
		},
	}
}
