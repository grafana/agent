package bijection

import (
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/loki/source/kafka"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	common "github.com/prometheus/common/config"
	"testing"
)

var promtailKafka = scrapeconfig.KafkaTargetConfig{
	Labels:               nil,
	UseIncomingTimestamp: false,
	Brokers:              nil,
	GroupID:              "",
	Topics:               nil,
	Version:              "",
	Assignor:             "",
	Authentication: scrapeconfig.KafkaAuthentication{
		Type: "",
		TLSConfig: common.TLSConfig{
			CA:                 "",
			Cert:               "",
			Key:                "",
			CAFile:             "",
			CertFile:           "",
			KeyFile:            "",
			ServerName:         "",
			InsecureSkipVerify: false,
			MinVersion:         0,
			MaxVersion:         0,
		},
		SASLConfig: scrapeconfig.KafkaSASLConfig{
			Mechanism: "",
			User:      "",
			Password:  flagext.SecretWithValue(""),
			UseTLS:    false,
			TLSConfig: common.TLSConfig{
				CA:                 "",
				Cert:               "",
				Key:                "",
				CAFile:             "",
				CertFile:           "",
				KeyFile:            "",
				ServerName:         "",
				InsecureSkipVerify: false,
				MinVersion:         0,
				MaxVersion:         0,
			},
		},
	},
}

var riverKafka = kafka.Arguments{
	Brokers:  nil,
	Topics:   nil,
	GroupID:  "",
	Assignor: "",
	Version:  "",
	Authentication: kafka.KafkaAuthentication{
		Type: "",
		TLSConfig: config.TLSConfig{
			CA:                 "",
			CAFile:             "",
			Cert:               "",
			CertFile:           "",
			Key:                "",
			KeyFile:            "",
			ServerName:         "",
			InsecureSkipVerify: false,
			MinVersion:         0,
		},
		SASLConfig: kafka.KafkaSASLConfig{
			Mechanism: "",
			User:      "",
			Password:  "",
			UseTLS:    false,
			TLSConfig: config.TLSConfig{
				CA:                 "",
				CAFile:             "",
				Cert:               "",
				CertFile:           "",
				Key:                "",
				KeyFile:            "",
				ServerName:         "",
				InsecureSkipVerify: false,
				MinVersion:         0,
			},
			OAuthConfig: kafka.OAuthConfigConfig{
				TokenProvider: "",
				Scopes:        nil,
			},
		},
	},
	UseIncomingTimestamp: false,
	Labels:               nil,
	ForwardTo:            nil,
	RelabelRules:         nil,
}

func TestKafkaBijection(t *testing.T) {
	tlsBj := tlsBijection()

	from := config.TLSConfig{
		CA:                 "ca",
		CAFile:             "cert-file",
		Cert:               "cert",
		CertFile:           "cert-file",
		Key:                "key",
		KeyFile:            "key-file",
		ServerName:         "server-name",
		InsecureSkipVerify: true,
		MinVersion:         123,
	}
	expectedTo := common.TLSConfig{
		CA:                 "ca",
		Cert:               "cert",
		Key:                "key",
		CAFile:             "cert-file",
		CertFile:           "cert-file",
		KeyFile:            "key-file",
		ServerName:         "server-name",
		InsecureSkipVerify: true,
		MinVersion:         123,
		MaxVersion:         0,
	}

	testTwoWayConversion(t, tlsBj, from, expectedTo)
}

func tlsBijection() *StructBijection[config.TLSConfig, common.TLSConfig] {
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
	return tlsBj
}

func saslBijection() *StructBijection[kafka.KafkaSASLConfig, scrapeconfig.KafkaSASLConfig] {
	type A = kafka.KafkaSASLConfig
	type B = scrapeconfig.KafkaSASLConfig

	tlsBj := &StructBijection[A, B]{}
	//TODO(thampiotr): finish
	return tlsBj
}
