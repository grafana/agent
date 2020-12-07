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
push_config:
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
push_config:
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
    retry_on_failure:
      max_elapsed_time: 60s
service:
  pipelines:
    traces:
      exporters: ["otlp"]
      processors: []
      receivers: ["jaeger"]
`,
		},
		{
			name: "push_config options",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
push_config:
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
    retry_on_failure:
      enabled: true
      max_elapsed_time: 60s
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
attributes:
  actions:
  - key: montgomery
    value: forever
    action: update
push_config:
  endpoint: example.com:12345
  batch:
    timeout: 5s
    send_batch_size: 100
  retry_on_failure:
    initial_interval: 10s
  sending_queue:
    num_consumers: 15	
`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp:
    endpoint: example.com:12345
    retry_on_failure:
      initial_interval: 10s
      max_elapsed_time: 60s
    sending_queue:
      num_consumers: 15
processors:
  attributes:
    actions:
    - key: montgomery
      value: forever
      action: update
  batch:
    timeout: 5s
    send_batch_size: 100
service:
  pipelines:
    traces:
      exporters: ["otlp"]
      processors: ["attributes", "batch"]
      receivers: ["jaeger"]
`,
		},
		{
			name: "push_config password in file",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
push_config:
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
    retry_on_failure:
      max_elapsed_time: 60s
service:
  pipelines:
    traces:
      exporters: ["otlp"]
      processors: []
      receivers: ["jaeger"]
`,
		},
		{
			name: "insecure skip verify",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
push_config:
  insecure_skip_verify: true
  endpoint: example.com:12345`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp:
    endpoint: example.com:12345
    insecure_skip_verify: true
    retry_on_failure:
      max_elapsed_time: 60s
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
