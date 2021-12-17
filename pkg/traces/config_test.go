package traces

import (
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configparser"
	"go.opentelemetry.io/collector/config/configunmarshaler"
	"gopkg.in/yaml.v2"
)

func tmpFile(t *testing.T, content string) (*os.File, func()) {
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)

	_, err = f.Write([]byte(content))
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

	return f, func() {
		os.Remove(f.Name())
	}
}

func TestOTelConfig(t *testing.T) {
	// create a password file to test the password file logic
	password := "password_in_file"
	passwordFile, teardown := tmpFile(t, password)
	defer teardown()

	// Extra linefeed in password_file. Spaces, tabs line feeds should be
	// stripped when reading it
	passwordFileExtraNewline, teardown := tmpFile(t, password+"\n")
	defer teardown()

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
  - endpoint: example.com:12345
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
batch:
  timeout: 5s
  send_batch_size: 100
remote_write:
  - endpoint: example.com:12345
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
      exporters: ["otlp/0"]
      processors: ["attributes", "batch"]
      receivers: ["jaeger"]
`,
		},
		{
			name: "password in file",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - insecure: true
    endpoint: example.com:12345
    basic_auth:
      username: test
      password_file: ` + passwordFile.Name(),
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp/0:
    endpoint: example.com:12345
    compression: gzip
    tls:
      insecure: true
    headers:
      authorization: Basic dGVzdDpwYXNzd29yZF9pbl9maWxl
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
			name: "password in file with extra newline",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - insecure: true
    endpoint: example.com:12345
    basic_auth:
      username: test
      password_file: ` + passwordFileExtraNewline.Name(),
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp/0:
    endpoint: example.com:12345
    compression: gzip
    tls:
      insecure: true
    headers:
      authorization: Basic dGVzdDpwYXNzd29yZF9pbl9maWxl
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
			name: "insecure skip verify",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - insecure_skip_verify: true
    endpoint: example.com:12345`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp/0:
    endpoint: example.com:12345
    compression: gzip
    tls:
      insecure_skip_verify: true
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
			name: "no compression",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - insecure_skip_verify: true
    endpoint: example.com:12345
    compression: none`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp/0:
    endpoint: example.com:12345
    tls:
      insecure_skip_verify: true
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
			name: "jaeger receiver remote_sampling TLS config",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
    remote_sampling:
      strategy_file: file_path
      tls:
        insecure: true
        insecure_skip_verify: true
        server_name_override: hostname
remote_write:
  - endpoint: example.com:12345
`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
    remote_sampling:
      strategy_file: file_path
      tls:
        insecure: true
        insecure_skip_verify: true
        server_name_override: hostname
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
    headers:
      x-some-header: Some value!
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
      x-some-header: Some value!
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
      password_file: ` + passwordFile.Name() + `
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
    tls:
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
			name: "span metrics remote write exporter",
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
  metrics_instance: traces
`,
			expectedConfig: `
receivers:
  noop:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp/0:
    endpoint: example.com:12345
    compression: gzip
    retry_on_failure:
      max_elapsed_time: 60s
  remote_write:
    namespace: traces_spanmetrics
    metrics_instance: traces
processors:
  spanmetrics:
    metrics_exporter: remote_write
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
      exporters: ["remote_write"]
      receivers: ["noop"]
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
  handler_endpoint: "0.0.0.0:8889"
`,
			expectedConfig: `
receivers:
  noop:
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
    namespace: traces_spanmetrics
processors:
  spanmetrics:
    metrics_exporter: prometheus
service:
  pipelines:
    traces:
      exporters: ["otlp/0"]
      processors: ["spanmetrics"]
      receivers: ["jaeger"]
    metrics/spanmetrics:
      exporters: ["prometheus"]
      receivers: ["noop"]
`,
		},
		{
			name: "span metrics prometheus and remote write exporters fail",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
spanmetrics:
  handler_endpoint: "0.0.0.0:8889"
  metrics_instance: traces
`,
			expectedError: true,
		},
		{
			name: "tail sampling config",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
tail_sampling:
  policies:
    - always_sample:
    - latency:
        threshold_ms: 5000
    - numeric_attribute:
        key: key1
        min_value: 50
        max_value: 100
    - probabilistic:
        sampling_percentage: 10
    - status_code:
        status_codes:
          - ERROR
          - UNSET
    - string_attribute:
        key: key
        values:
          - value1
          - value2
    - rate_limiting:
        spans_per_second: 35
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
  tail_sampling:
    decision_wait: 5s
    policies:
      - name: always_sample/0
        type: always_sample
      - name: latency/1
        type: latency
        latency:
          threshold_ms: 5000
      - name: numeric_attribute/2
        type: numeric_attribute
        numeric_attribute:
          key: key1
          min_value: 50
          max_value: 100
      - name: probabilistic/3
        type: probabilistic
        probabilistic:
          sampling_percentage: 10
      - name: status_code/4
        type: status_code
        status_code:
          status_codes:
            - ERROR
            - UNSET
      - name: string_attribute/5
        type: string_attribute
        string_attribute:
          key: key
          values:
            - value1
            - value2
      - name: rate_limiting/6
        type: rate_limiting
        rate_limiting:
          spans_per_second: 35
service:
  pipelines:
    traces:
      exporters: ["otlp/0"]
      processors: ["tail_sampling"]
      receivers: ["jaeger"]
`,
		},
		{
			name: "tail sampling config with load balancing",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
tail_sampling:
  policies:
    - always_sample:
    - string_attribute:
        key: key
        values:
          - value1
          - value2
load_balancing:
  receiver_port: 8080
  exporter:
    insecure: true
  resolver:
    dns:
      hostname: agent
      port: 8080
`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
  otlp/lb:
    protocols:
      grpc:
        endpoint: "0.0.0.0:8080"
exporters:
  otlp/0:
    endpoint: example.com:12345
    compression: gzip
    retry_on_failure:
      max_elapsed_time: 60s
  loadbalancing:
    protocol:
      otlp:
        tls:
          insecure: true
        endpoint: noop
        retry_on_failure:
          max_elapsed_time: 60s
    resolver:
      dns:
        hostname: agent
        port: 8080
processors:
  tail_sampling:
    decision_wait: 5s
    policies:
      - name: always_sample/0
        type: always_sample
      - name: string_attribute/1
        type: string_attribute
        string_attribute:
          key: key
          values:
            - value1
            - value2
service:
  pipelines:
    traces/0:
      exporters: ["loadbalancing"]
      processors: []
      receivers: ["jaeger"]
    traces/1:
      exporters: ["otlp/0"]
      processors: ["tail_sampling"]
      receivers: ["otlp/lb"]
`,
		},
		{
			name: "automatic logging : default",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
automatic_logging:
  spans: true
`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
processors:
  automatic_logging:
    automatic_logging:
      spans: true
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
      processors: ["automatic_logging"]
      receivers: ["jaeger"]
      `,
		},
		{
			name: "tls config",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - insecure: false
    tls_config:
      ca_file: server.crt
      cert_file: client.crt
      key_file: client.key
    endpoint: example.com:12345
`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlp/0:
    endpoint: example.com:12345
    tls:
      insecure: false
      ca_file: server.crt
      cert_file: client.crt
      key_file: client.key
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
			name: "otlp http & grpc exporters",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
    protocol: http
  - endpoint: example.com:12345
    protocol: grpc
`,
			expectedConfig: `
receivers:
  jaeger:
    protocols:
      grpc:
exporters:
  otlphttp/0:
    endpoint: example.com:12345
    compression: gzip
    retry_on_failure:
      max_elapsed_time: 60s
  otlp/1:
    endpoint: example.com:12345
    compression: gzip
    retry_on_failure:
      max_elapsed_time: 60s
service:
  pipelines:
    traces:
      exporters: ["otlphttp/0", "otlp/1"]
      processors: []
      receivers: ["jaeger"]
`,
		},
		{
			name: "prom SD config",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
    protocol: grpc
scrape_configs:
  - im_a_scrape_config
prom_sd_operation_type: update
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
  prom_sd_processor:
    scrape_configs:
      - im_a_scrape_config
    operation_type: update
service:
  pipelines:
    traces:
      exporters: ["otlp/0"]
      processors: ["prom_sd_processor"]
      receivers: ["jaeger"]
`,
		},
		{
			name: "service graphs",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
service_graphs:
  enabled: true
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
  service_graphs:
service:
  pipelines:
    traces:
      exporters: ["otlp/0"]
      processors: ["service_graphs"]
      receivers: ["jaeger"]
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

			factories, err := tracingFactories()
			require.NoError(t, err)

			configMap := configparser.NewConfigMapFromStringMap(otelMapStructure)
			cfgUnmarshaler := configunmarshaler.NewDefault()
			expectedConfig, err := cfgUnmarshaler.Unmarshal(configMap, factories)
			require.NoError(t, err)

			// Exporters and receivers in the config's pipelines need to be in the same order for them to be asserted as equal
			sortPipelines(actualConfig)
			sortPipelines(expectedConfig)

			assert.Equal(t, expectedConfig, actualConfig)
		})
	}
}

func TestProcessorOrder(t *testing.T) {
	// tests!
	tt := []struct {
		name               string
		cfg                string
		expectedProcessors map[string][]config.ComponentID
	}{
		{
			name: "no processors",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
    headers:
      x-some-header: Some value!
`,
			expectedProcessors: map[string][]config.ComponentID{
				"traces": nil,
			},
		},
		{
			name: "all processors w/o load balancing",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
    headers:
      x-some-header: Some value!
attributes:
  actions:
  - key: montgomery
    value: forever
    action: update
spanmetrics:
  latency_histogram_buckets: [2ms, 6ms, 10ms, 100ms, 250ms]
  dimensions:
    - name: http.method
      default: GET
    - name: http.status_code
  metrics_instance: traces
automatic_logging:
  spans: true
batch:
  timeout: 5s
  send_batch_size: 100
tail_sampling:
  policies:
    - always_sample:
    - string_attribute:
        key: key
        values:
          - value1
          - value2
service_graphs:
  enabled: true
`,
			expectedProcessors: map[string][]config.ComponentID{
				"traces": {
					config.NewID("attributes"),
					config.NewID("spanmetrics"),
					config.NewID("service_graphs"),
					config.NewID("tail_sampling"),
					config.NewID("automatic_logging"),
					config.NewID("batch"),
				},
				"metrics/spanmetrics": nil,
			},
		},
		{
			name: "all processors with load balancing",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
    headers:
      x-some-header: Some value!
attributes:
  actions:
  - key: montgomery
    value: forever
    action: update
spanmetrics:
  latency_histogram_buckets: [2ms, 6ms, 10ms, 100ms, 250ms]
  dimensions:
    - name: http.method
      default: GET
    - name: http.status_code
  metrics_instance: traces
automatic_logging:
  spans: true
batch:
  timeout: 5s
  send_batch_size: 100
tail_sampling:
  policies:
    - always_sample:
    - string_attribute:
        key: key
        values:
          - value1
          - value2
load_balancing:
  exporter:
    tls:
      insecure: true
  resolver:
    dns:
      hostname: agent
      port: 4318
service_graphs:
  enabled: true
`,
			expectedProcessors: map[string][]config.ComponentID{
				"traces/0": {
					config.NewID("attributes"),
					config.NewID("spanmetrics"),
				},
				"traces/1": {
					config.NewID("service_graphs"),
					config.NewID("tail_sampling"),
					config.NewID("automatic_logging"),
					config.NewID("batch"),
				},
				"metrics/spanmetrics": nil,
			},
		},
		{
			name: "load balancing without tail sampling",
			cfg: `
receivers:
  jaeger:
    protocols:
      grpc:
remote_write:
  - endpoint: example.com:12345
    headers:
      x-some-header: Some value!
attributes:
  actions:
  - key: montgomery
    value: forever
    action: update
spanmetrics:
  latency_histogram_buckets: [2ms, 6ms, 10ms, 100ms, 250ms]
  dimensions:
    - name: http.method
      default: GET
    - name: http.status_code
  metrics_instance: traces
automatic_logging:
  spans: true
batch:
  timeout: 5s
  send_batch_size: 100
load_balancing:
  exporter:
    tls:
      insecure: true
  resolver:
    dns:
      hostname: agent
      port: 4318
`,
			expectedProcessors: map[string][]config.ComponentID{
				"traces/0": {
					config.NewID("attributes"),
					config.NewID("spanmetrics"),
				},
				"traces/1": {
					config.NewID("automatic_logging"),
					config.NewID("batch"),
				},
				"metrics/spanmetrics": nil,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var cfg InstanceConfig
			err := yaml.Unmarshal([]byte(tc.cfg), &cfg)
			require.NoError(t, err)

			// check error
			actualConfig, err := cfg.otelConfig()
			require.NoError(t, err)

			require.Equal(t, len(tc.expectedProcessors), len(actualConfig.Pipelines))
			for k := range tc.expectedProcessors {
				assert.Equal(t, tc.expectedProcessors[k], actualConfig.Pipelines[k].Processors)
			}
		})
	}
}

func TestOrderProcessors(t *testing.T) {
	tests := []struct {
		processors     []string
		splitPipelines bool
		expected       [][]string
	}{
		{
			expected: [][]string{
				nil,
			},
		},
		{
			processors: []string{
				"tail_sampling",
			},
			expected: [][]string{
				{"tail_sampling"},
			},
		},
		{
			processors: []string{
				"batch",
				"tail_sampling",
				"automatic_logging",
			},
			expected: [][]string{
				{
					"tail_sampling",
					"automatic_logging",
					"batch",
				},
			},
		},
		{
			processors: []string{
				"spanmetrics",
				"batch",
				"tail_sampling",
				"attributes",
				"automatic_logging",
			},
			expected: [][]string{
				{
					"attributes",
					"spanmetrics",
					"tail_sampling",
					"automatic_logging",
					"batch",
				},
			},
		},
		{
			splitPipelines: true,
			expected: [][]string{
				nil,
				nil,
			},
		},
		{
			processors: []string{
				"spanmetrics",
				"batch",
				"tail_sampling",
				"attributes",
				"automatic_logging",
			},
			splitPipelines: true,
			expected: [][]string{
				{
					"attributes",
					"spanmetrics",
				},
				{
					"tail_sampling",
					"automatic_logging",
					"batch",
				},
			},
		},
		{
			processors: []string{
				"batch",
				"tail_sampling",
				"automatic_logging",
			},
			splitPipelines: true,
			expected: [][]string{
				{},
				{
					"tail_sampling",
					"automatic_logging",
					"batch",
				},
			},
		},
		{
			processors: []string{
				"spanmetrics",
				"attributes",
			},
			splitPipelines: true,
			expected: [][]string{
				{
					"attributes",
					"spanmetrics",
				},
				{},
			},
		},
	}

	for _, tc := range tests {
		actual := orderProcessors(tc.processors, tc.splitPipelines)
		assert.Equal(t, tc.expected, actual)
	}
}

// sortPipelines is a helper function to lexicographically sort a pipeline's exporters
func sortPipelines(cfg *config.Config) {
	tracePipeline := cfg.Pipelines[string(config.TracesDataType)]
	if tracePipeline == nil {
		return
	}
	var (
		exp  = tracePipeline.Exporters
		recv = tracePipeline.Receivers
	)
	sort.Slice(exp, func(i, j int) bool { return exp[i].String() > exp[j].String() })
	sort.Slice(recv, func(i, j int) bool { return recv[i].String() > recv[j].String() })
}
