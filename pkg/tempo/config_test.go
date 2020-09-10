package tempo

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"gopkg.in/yaml.v2"
)

func TestOTelConfig(t *testing.T) {
	// create a password file to test the password file logic
	password := "password_in_file"
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(password))
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	// tests!
	tt := []struct {
		name           string
		cfg            string
		expectedError  bool
		expectedConfig string
	}{
		{
			name:          "disabled",
			cfg:           "",
			expectedError: true,
		},
		{
			name: "no receivers",
			cfg: `
receivers:
`,
			expectedError: true,
		},
		{
			name: "no rw endpoint",
			cfg: `
receivers:
  jaeger:
`,
			expectedError: true,
		},
		{
			name: "empty receiver config",
			cfg: `
receivers:
  jaeger:
remote_write:
  endpoint: example.com:12345
`,
			expectedError: true,
		},
		{
			name: "basic config",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  endpoint: example.com:12345
`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp:
    endpoint: example.com:12345
service:
  pipelines:
    traces:
      exporters: ["otlp"]
      processors: []
      receivers: ["jaeger"]
`,
		},
		{
			name: "remote_write options",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  insecure: true
  endpoint: example.com:12345
  basic_auth:
    username: test
    password: blerg
`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp:
    endpoint: example.com:12345
    insecure: true
    headers:
      authorization: Basic dGVzdDpibGVyZw==
service:
  pipelines:
    traces:
      exporters: ["otlp"]
      processors: []
      receivers: ["jaeger"]
`,
		},
		{
			name: "processor config",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  endpoint: example.com:12345
  batch:
    timeout: 5s
    send_batch_size: 100
  queue:
    backoff_delay: 10s
    num_workers: 15	
`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp:
    endpoint: example.com:12345
processors:
  batch:
    timeout: 5s
    send_batch_size: 100
  queued_retry:
    backoff_delay: 10s
    num_workers: 15	
service:
  pipelines:
    traces:
      exporters: ["otlp"]
      processors: ["batch", "queued_retry"]
      receivers: ["jaeger"]
`,
		},
		{
			name: "remote_write password in file",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  insecure: true
  endpoint: example.com:12345
  basic_auth:
    username: test
    password_file: ` + tmpfile.Name(),
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp:
    endpoint: example.com:12345
    insecure: true
    headers:
      authorization: Basic dGVzdDpwYXNzd29yZF9pbl9maWxl
service:
  pipelines:
    traces:
      exporters: ["otlp"]
      processors: []
      receivers: ["jaeger"]
`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{}
			err := yaml.Unmarshal([]byte(tc.cfg), cfg)
			require.NoError(t, err)

			// check error
			actualConfig, err := cfg.otelConfig()
			if tc.expectedError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// convert actual config to otel config
			otelMapStructure := map[string]interface{}{}
			err = yaml.Unmarshal([]byte(tc.expectedConfig), otelMapStructure)
			require.NoError(t, err)

			v := viper.New()
			err = v.MergeConfigMap(otelMapStructure)
			require.NoError(t, err)

			factories, err := tracingFactories()
			require.NoError(t, err)

			expectedConfig, err := config.Load(v, factories)
			require.NoError(t, err)

			assert.Equal(t, expectedConfig, actualConfig)
		})
	}
}
