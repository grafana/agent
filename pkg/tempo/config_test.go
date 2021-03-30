package tempo

import (
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
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
    compression: gzip
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
    compression: gzip
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
    compression: gzip
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
    compression: gzip
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
    compression: gzip
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
		{
			name: "no compression",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
push_config:
  insecure_skip_verify: true
  endpoint: example.com:12345
  compression: none`,
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
		{
			name: "push_config and remote_write",
			cfg: `
receivers:
  jaeger:
push_config:
  endpoint: example:12345
remote_write:
  - endpoint: anotherexample.com:12345
`,
			expectedError: true,
		},
		{
			name: "push_config.batch and batch",
			cfg: `
receivers:
  jaeger:
push_config:
  endpoint: example:12345
  batch:
    timeout: 5s
    send_batch_size: 100
batch:
  timeout: 5s
  send_batch_size: 100
remote_write:
  - endpoint: anotherexample.com:12345
`,
			expectedError: true,
		},
		{
			name: "one backend with remote_write",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp/0:
    endpoint: example.com:12345
    compression: gzip
    retry_on_failure:
      max_elapsed_time: 60s
service:
  pipelines:
    traces:
      exporters: ["otlp/0"]
      processors: []
      receivers: ["jaeger"]
`,
		},
		{
			name: "two backends in a remote_write block",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
    basic_auth:
      username: test
      password: blerg
  - endpoint: anotherexample.com:12345
    compression: none
    insecure: false
    insecure_skip_verify: true
    basic_auth:
      username: test
      password_file: ` + tmpfile.Name() + `
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
  otlp/0:
    endpoint: example.com:12345
    compression: gzip
    headers:
      authorization: Basic dGVzdDpibGVyZw==
    retry_on_failure:
      max_elapsed_time: 60s
  otlp/1:
    endpoint: anotherexample.com:12345
    insecure: false
    insecure_skip_verify: true
    headers:
      authorization: Basic dGVzdDpwYXNzd29yZF9pbl9maWxl
    retry_on_failure:
      initial_interval: 10s
      max_elapsed_time: 60s
    sending_queue:
      num_consumers: 15
service:
  pipelines:
    traces:
      exporters: ["otlp/1", "otlp/0"]
      processors: []
      receivers: ["jaeger"]
`,
		},
		{
			name: "batch block",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
batch:
  timeout: 5s
  send_batch_size: 100
`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp/0:
    endpoint: example.com:12345
    compression: gzip
    retry_on_failure:
      max_elapsed_time: 60s
processors:
  batch:
    timeout: 5s
    send_batch_size: 100
service:
  pipelines:
    traces:
      exporters: ["otlp/0"]
      processors: ["batch"]
      receivers: ["jaeger"]
`,
		},
		{
			name: "span metrics prometheus exporter",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
spanmetrics:
  latency_histogram_buckets: [2ms, 6ms, 10ms, 100ms, 250ms]
  dimensions:
    - name: http.method
      default: GET
    - name: http.status_code
  metrics_exporter:
    name: prometheus
    config:
      endpoint: "0.0.0.0:8889"
      namespace: promexample
`,
			expectedConfig: `
receivers:
  dummy:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp/0:
    endpoint: example.com:12345
    compression: gzip
    retry_on_failure:
      max_elapsed_time: 60s
  prometheus:
    endpoint: "0.0.0.0:8889"
    namespace: promexample    
processors:
  spanmetrics:
    metrics_exporter: prometheus
    latency_histogram_buckets: [2ms, 6ms, 10ms, 100ms, 250ms]
    dimensions:
      - name: http.method
        default: GET
      - name: http.status_code
service:
  pipelines:
    traces:
      exporters: ["otlp/0"]
      processors: ["spanmetrics"]
      receivers: ["jaeger"]
    metrics/spanmetrics:
      exporters: ["prometheus"]
      receivers: ["dummy"]
`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var cfg InstanceConfig
			err := yaml.Unmarshal([]byte(tc.cfg), &cfg)
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

			// Exporters in the config's pipelines need to be in the same order for them to be asserted as equal
			sortPipelinesExporters(actualConfig)
			sortPipelinesExporters(expectedConfig)

			assert.Equal(t, expectedConfig, actualConfig)
		})
	}
}

// sortPipelinesExporters is a helper function to lexicographically sort a pipeline's exporters
func sortPipelinesExporters(cfg *configmodels.Config) {
	for _, p := range cfg.Pipelines {
		sort.Strings(p.Exporters)
	}
}
