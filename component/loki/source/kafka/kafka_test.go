package kafka

import (
	"testing"

	"github.com/grafana/river"
	"github.com/stretchr/testify/require"
)

func TestRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	brokers                = ["localhost:9092","localhost:23456"]
	topics                 = ["quickstart-events"]
	labels                 = {component = "loki.source.kafka"}
	forward_to             = []
	use_incoming_timestamp = true
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}

func TestTLSRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	brokers                = ["localhost:9092","localhost:23456"]
	topics                 = ["quickstart-events"]
	authentication {
		type = "ssl"
		tls_config {
			cert_file = "/fake/file.cert"
            key_file  = "/fake/file.key"
		}
	}
	labels                 = {component = "loki.source.kafka"}
	forward_to             = []
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}

func TestSASLRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	brokers                = ["localhost:9092","localhost:23456"]
	topics                 = ["quickstart-events"]
	authentication {
		type = "sasl"
		sasl_config {
			user     = "user"
            password = "password"
		}
	}
	labels                 = {component = "loki.source.kafka"}
	forward_to             = []
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	// Ensures that the rivertype.Secret is working properly and there's no error
	password, err_ := ConvertSecretToString(args.Authentication.SASLConfig.Password)
	require.NoError(t, err_)
	require.Equal(t, "password", password)
}

func TestSASLOAuthRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	brokers = ["localhost:9092", "localhost:23456"]
	topics  = ["quickstart-events"]

	authentication {
		type = "sasl"
		sasl_config {
			mechanism = "OAUTHBEARER"
			oauth_config {
				token_provider = "azure"
				scopes         = ["my-scope"]
			}
		}
	}
	labels     = {component = "loki.source.kafka"}
	forward_to = []
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}
