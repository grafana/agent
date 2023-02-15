package scrape

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/river"
	prometheus_client "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/require"
)

func TestRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	targets         = [{ "target1" = "target1" }]
	forward_to      = []
	scrape_interval = "10s"
	job_name        = "local-flow"

	bearer_token = "token"
	bearer_token_file = "/path/to/file.token"
	proxy_url = "http://0.0.0.0:11111"
	follow_redirects = true
	enable_http2 = true

	basic_auth {
		username = "user"
		password = "password"
		password_file = "/path/to/file.password"
	}

	authorization {
		type = "Bearer"
		credentials = "credential"
		credentials_file = "/path/to/file.credentials"
	}

	oauth2 {
		client_id = "client_id"
		client_secret = "client_secret"
		client_secret_file = "/path/to/file.oath2"
		scopes = ["scope1", "scope2"]
		token_url = "token_url"
		endpoint_params = {"param1" = "value1", "param2" = "value2"}
		proxy_url = "http://0.0.0.0:11111"
		tls_config = {
			ca_file = "/path/to/file.ca",
			cert_file = "/path/to/file.cert",
			key_file = "/path/to/file.key",
			server_name = "server_name",
			insecure_skip_verify = false,
			min_version = "TLS13",
		}
	}

	tls_config {
		ca_file = "/path/to/file.ca"
		cert_file = "/path/to/file.cert"
		key_file = "/path/to/file.key"
		server_name = "server_name"
		insecure_skip_verify = false
		min_version = "TLS13"
	}
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}

func TestForwardingToAppendable(t *testing.T) {
	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)
	opts := component.Options{
		Logger:     l,
		Registerer: prometheus_client.NewRegistry(),
	}

	nilReceivers := []storage.Appendable{nil, nil}

	args := DefaultArguments
	args.ForwardTo = nilReceivers

	s, err := New(opts, args)
	require.NoError(t, err)

	// Forwarding samples to the nil receivers shouldn't fail.
	appender := s.appendable.Appender(context.Background())
	_, err = appender.Append(0, labels.FromStrings("foo", "bar"), 0, 0)
	require.NoError(t, err)

	err = appender.Commit()
	require.NoError(t, err)

	// Update the component with a mock receiver; it should be passed along to the Appendable.
	var receivedTs int64
	var receivedSamples labels.Labels
	fanout := prometheus.NewInterceptor(nil, prometheus.WithAppendHook(func(ref storage.SeriesRef, l labels.Labels, t int64, _ float64, _ storage.Appender) (storage.SeriesRef, error) {
		receivedTs = t
		receivedSamples = l
		return ref, nil
	}))
	require.NoError(t, err)
	args.ForwardTo = []storage.Appendable{fanout}
	err = s.Update(args)
	require.NoError(t, err)

	// Forwarding a sample to the mock receiver should succeed.
	appender = s.appendable.Appender(context.Background())
	timestamp := time.Now().Unix()
	sample := labels.FromStrings("foo", "bar")
	_, err = appender.Append(0, sample, timestamp, 42.0)
	require.NoError(t, err)

	err = appender.Commit()
	require.NoError(t, err)

	require.Equal(t, receivedTs, timestamp)
	require.Len(t, receivedSamples, 1)
	require.Equal(t, receivedSamples, sample)
}
