package tempo

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/grafana/agent/pkg/tempo/noopreceiver"
	"github.com/grafana/agent/pkg/tempo/promsdprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanmetricsprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"
	prom_config "github.com/prometheus/common/config"
	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/prometheusexporter"
	"go.opentelemetry.io/collector/processor/attributesprocessor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/receiver/jaegerreceiver"
	"go.opentelemetry.io/collector/receiver/kafkareceiver"
	"go.opentelemetry.io/collector/receiver/opencensusreceiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/zipkinreceiver"
)

const (
	spanMetricsPipelineName    = "metrics/spanmetrics"
	defaultSpanMetricsExporter = "prometheus"

	// defaultDecisionWait is the default time to wait for a trace before making a sampling decision
	defaultDecisionWait = time.Second * 5

	// defaultLoadBalancingPort is the default port the agent uses for internal load balancing
	defaultLoadBalancingPort = "9999"
	// agent's load balancing options
	dnsTagName    = "dns"
	staticTagName = "static"

	// sampling policies
	alwaysSamplePolicy     = "always_sample"
	stringAttributePolicy  = "string_attribute"
	numericAttributePolicy = "numeric_attribute"
	rateLimitingPolicy     = "rate_limiting"
)

// Config controls the configuration of Tempo trace pipelines.
type Config struct {
	Configs []InstanceConfig `yaml:"configs,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	return c.Validate()
}

// Validate ensures that the Config is valid.
func (c *Config) Validate() error {
	names := make(map[string]struct{}, len(c.Configs))
	for idx, c := range c.Configs {
		if c.Name == "" {
			return fmt.Errorf("tempo config at index %d is missing a name", idx)
		}
		if _, exist := names[c.Name]; exist {
			return fmt.Errorf("found multiple tempo configs with name %s", c.Name)
		}
		names[c.Name] = struct{}{}
	}

	return nil
}

// InstanceConfig configures an individual Tempo trace pipeline.
type InstanceConfig struct {
	Name string `yaml:"name"`

	// Deprecated in favor of RemoteWrite and Batch.
	PushConfig PushConfig `yaml:"push_config,omitempty"`

	// RemoteWrite defines one or multiple backends that can receive the pipeline's traffic.
	RemoteWrite []RemoteWriteConfig `yaml:"remote_write,omitempty"`

	// Receivers: https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/receiver/README.md
	Receivers map[string]interface{} `yaml:"receivers,omitempty"`

	// Batch: https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/processor/batchprocessor/config.go#L24
	Batch map[string]interface{} `yaml:"batch,omitempty"`

	// Attributes: https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/processor/attributesprocessor/config.go#L30
	Attributes map[string]interface{} `yaml:"attributes,omitempty"`

	// prom service discovery
	ScrapeConfigs []interface{} `yaml:"scrape_configs,omitempty"`

	// SpanMetricsProcessor: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/spanmetricsprocessor/README.md
	SpanMetrics *SpanMetricsConfig `yaml:"spanmetrics,omitempty"`

	// TailSampling defines a sampling strategy for the pipeline
	TailSampling *TailSamplingConfig `yaml:"tail_sampling"`
}

const (
	compressionNone = "none"
	compressionGzip = "gzip"
)

// DefaultPushConfig holds the default settings for a PushConfig.
var DefaultPushConfig = PushConfig{
	Compression: compressionGzip,
}

// PushConfig controls the configuration of exporting to Grafana Cloud
type PushConfig struct {
	Endpoint           string                 `yaml:"endpoint,omitempty"`
	Compression        string                 `yaml:"compression,omitempty"`
	Insecure           bool                   `yaml:"insecure,omitempty"`
	InsecureSkipVerify bool                   `yaml:"insecure_skip_verify,omitempty"`
	BasicAuth          *prom_config.BasicAuth `yaml:"basic_auth,omitempty,omitempty"`
	Batch              map[string]interface{} `yaml:"batch,omitempty"`            // https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/processor/batchprocessor/config.go#L24
	SendingQueue       map[string]interface{} `yaml:"sending_queue,omitempty"`    // https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/exporter/exporterhelper/queued_retry.go#L30
	RetryOnFailure     map[string]interface{} `yaml:"retry_on_failure,omitempty"` // https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/exporter/exporterhelper/queued_retry.go#L54
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *PushConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultPushConfig

	type plain PushConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if c.Compression != compressionGzip && c.Compression != compressionNone {
		return fmt.Errorf("unsupported compression '%s', expected 'gzip' or 'none'", c.Compression)
	}
	return nil
}

// DefaultRemoteWriteConfig holds the default settings for a PushConfig.
var DefaultRemoteWriteConfig = RemoteWriteConfig{
	Compression: compressionGzip,
}

// RemoteWriteConfig controls the configuration of an exporter
type RemoteWriteConfig struct {
	Endpoint           string                 `yaml:"endpoint,omitempty"`
	Compression        string                 `yaml:"compression,omitempty"`
	Insecure           bool                   `yaml:"insecure,omitempty"`
	InsecureSkipVerify bool                   `yaml:"insecure_skip_verify,omitempty"`
	BasicAuth          *prom_config.BasicAuth `yaml:"basic_auth,omitempty"`
	Headers            map[string]string      `yaml:"headers,omitempty"`
	SendingQueue       map[string]interface{} `yaml:"sending_queue,omitempty"`    // https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/exporter/exporterhelper/queued_retry.go#L30
	RetryOnFailure     map[string]interface{} `yaml:"retry_on_failure,omitempty"` // https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/exporter/exporterhelper/queued_retry.go#L54
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *RemoteWriteConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultRemoteWriteConfig

	type plain RemoteWriteConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if c.Compression != compressionGzip && c.Compression != compressionNone {
		return fmt.Errorf("unsupported compression '%s', expected 'gzip' or 'none'", c.Compression)
	}
	return nil
}

// SpanMetricsConfig controls the configuration of spanmetricsprocessor and the related metrics exporter.
type SpanMetricsConfig struct {
	LatencyHistogramBuckets []time.Duration                  `yaml:"latency_histogram_buckets,omitempty"`
	Dimensions              []spanmetricsprocessor.Dimension `yaml:"dimensions,omitempty"`

	// MetricsExporter is a Prometheus metrics exporter
	MetricsExporter metricsExporterConfig `yaml:"metrics_exporter,omitempty"`
}

// Configuration for Prometheus exporter: https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34/exporter/prometheusexporter/README.md.
type metricsExporterConfig struct {
	// The address on which the Prometheus scrape handler will be run on.
	Endpoint string `yaml:"endpoint"`
	// Namespace if set, exports metrics under the provided value.
	Namespace string `yaml:"namespace"`
	// ConstLabels are values that are applied for every exported metric.
	ConstLabels map[string]interface{} `yaml:"const_labels"`
	// SendTimestamps will send the underlying scrape timestamp with the export
	SendTimestamps bool `yaml:"send_timestamps"`
}

// Configuration for tail-based sampling
type TailSamplingConfig struct {
	// Policies are the strategies used for sampling. Multiple policies can be used in the same pipeline.
	// For more information, refer to https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/tailsamplingprocessor
	Policies []map[string]interface{} `yaml:"policies"`
	// DecisionWait defines the time to wait for a complete trace before making a decision
	DecisionWait time.Duration `yaml:"decision_wait,omitempty"`
	// LoadBalancing is used to distribute spans of the same trace to the same agent instance
	LoadBalancing loadBalancingConfig `yaml:"load_balancing"`
}

// loadBalancingConfig defines the configuration for load balancing spans between agent instances
// loadBalancingConfig is an OTel exporter's config with extra resolver config
type loadBalancingConfig struct {
	Compression        string                 `yaml:"compression,omitempty"`
	Insecure           bool                   `yaml:"insecure,omitempty"`
	InsecureSkipVerify bool                   `yaml:"insecure_skip_verify,omitempty"`
	BasicAuth          *prom_config.BasicAuth `yaml:"basic_auth,omitempty"`
	Resolver           map[string]interface{} `yaml:"resolver"`
}

// exporter builds an OTel exporter from RemoteWriteConfig
func exporter(remoteWriteConfig RemoteWriteConfig) (map[string]interface{}, error) {
	if len(remoteWriteConfig.Endpoint) == 0 {
		return nil, errors.New("must have a configured a backend endpoint")
	}

	headers := map[string]string{}
	if remoteWriteConfig.Headers != nil {
		headers = remoteWriteConfig.Headers
	}

	if remoteWriteConfig.BasicAuth != nil {
		password := string(remoteWriteConfig.BasicAuth.Password)

		if len(remoteWriteConfig.BasicAuth.PasswordFile) > 0 {
			buff, err := ioutil.ReadFile(remoteWriteConfig.BasicAuth.PasswordFile)
			if err != nil {
				return nil, fmt.Errorf("unable to load password file %s: %w", remoteWriteConfig.BasicAuth.PasswordFile, err)
			}
			password = string(buff)
		}

		encodedAuth := base64.StdEncoding.EncodeToString([]byte(remoteWriteConfig.BasicAuth.Username + ":" + password))
		headers["authorization"] = "Basic " + encodedAuth
	}

	compression := remoteWriteConfig.Compression
	if compression == compressionNone {
		compression = ""
	}

	otlpExporter := map[string]interface{}{
		"endpoint":             remoteWriteConfig.Endpoint,
		"compression":          compression,
		"headers":              headers,
		"insecure":             remoteWriteConfig.Insecure,
		"insecure_skip_verify": remoteWriteConfig.InsecureSkipVerify,
		"sending_queue":        remoteWriteConfig.SendingQueue,
		"retry_on_failure":     remoteWriteConfig.RetryOnFailure,
	}

	// Apply some sane defaults to the exporter. The
	// sending_queue.retry_on_failure default is 300s which prevents any
	// sending-related errors to not be logged for 5 minutes. We'll lower that
	// to 60s.
	if retryConfig := otlpExporter["retry_on_failure"].(map[string]interface{}); retryConfig == nil {
		otlpExporter["retry_on_failure"] = map[string]interface{}{
			"max_elapsed_time": "60s",
		}
	} else if retryConfig["max_elapsed_time"] == nil {
		retryConfig["max_elapsed_time"] = "60s"
	}

	return otlpExporter, nil
}

// exporters builds one or multiple exporters from a remote_write block.
// It also supports building an exporter from push_config.
func (c *InstanceConfig) exporters() (map[string]interface{}, error) {
	if len(c.RemoteWrite) == 0 {
		otlpExporter, err := exporter(RemoteWriteConfig{
			Endpoint:           c.PushConfig.Endpoint,
			Compression:        c.PushConfig.Compression,
			Insecure:           c.PushConfig.Insecure,
			InsecureSkipVerify: c.PushConfig.InsecureSkipVerify,
			BasicAuth:          c.PushConfig.BasicAuth,
			SendingQueue:       c.PushConfig.SendingQueue,
			RetryOnFailure:     c.PushConfig.RetryOnFailure,
		})
		return map[string]interface{}{
			"otlp": otlpExporter,
		}, err
	}

	exporters := map[string]interface{}{}
	for i, remoteWriteConfig := range c.RemoteWrite {
		exporter, err := exporter(remoteWriteConfig)
		if err != nil {
			return nil, err
		}
		exporterName := fmt.Sprintf("otlp/%d", i)
		exporters[exporterName] = exporter
	}
	return exporters, nil
}

func resolver(config map[string]interface{}) (map[string]interface{}, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("must configure one resolver config (dns or static)")
	}
	resolverCfg := make(map[string]interface{})
	for typ, cfg := range config {
		switch typ {
		case dnsTagName, staticTagName:
			resolverCfg[typ] = cfg
		default:
			return nil, fmt.Errorf("unsupported resolver config type: %s", typ)
		}
	}
	return resolverCfg, nil
}

func (c *InstanceConfig) loadBalancingExporter() (map[string]interface{}, error) {
	exporter, err := exporter(RemoteWriteConfig{
		// Endpoint is omitted in OTel load balancing exporter
		Endpoint:           "noop",
		Compression:        c.TailSampling.LoadBalancing.Compression,
		Insecure:           c.TailSampling.LoadBalancing.Insecure,
		InsecureSkipVerify: c.TailSampling.LoadBalancing.InsecureSkipVerify,
		BasicAuth:          c.TailSampling.LoadBalancing.BasicAuth,
	})
	if err != nil {
		return nil, err
	}
	resolverCfg, err := resolver(c.TailSampling.LoadBalancing.Resolver)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"protocol": map[string]interface{}{
			"otlp": exporter,
		},
		"resolver": resolverCfg,
	}, nil
}

// formatPolicies creates sampling policies (i.e. rules) compatible with OTel's tail sampling processor
// https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.21.0/processor/tailsamplingprocessor
func formatPolicies(cfg []map[string]interface{}) ([]map[string]interface{}, error) {
	policies := make([]map[string]interface{}, 0, len(cfg))
	for i, policy := range cfg {
		if len(policy) != 1 {
			return nil, errors.New("malformed sampling policy")
		}
		for typ, rules := range policy {
			switch typ {
			case alwaysSamplePolicy:
				policies = append(policies, map[string]interface{}{
					"name": fmt.Sprintf("%s/%d", typ, i),
					"type": typ,
				})
			case stringAttributePolicy, rateLimitingPolicy, numericAttributePolicy:
				policies = append(policies, map[string]interface{}{
					"name": fmt.Sprintf("%s/%d", typ, i),
					"type": typ,
					typ:    rules,
				})
			default:
				return nil, fmt.Errorf("unsupported policy type %s", typ)
			}
		}
	}
	return policies, nil
}

func (c *InstanceConfig) otelConfig() (*configmodels.Config, error) {
	otelMapStructure := map[string]interface{}{}

	if len(c.Receivers) == 0 {
		return nil, errors.New("must have at least one configured receiver")
	}

	if len(c.RemoteWrite) != 0 && len(c.PushConfig.Endpoint) != 0 {
		return nil, errors.New("must not configure push_config and remote_write. push_config is deprecated in favor of remote_write")
	}

	if c.Batch != nil && c.PushConfig.Batch != nil {
		return nil, errors.New("must not configure push_config.batch and batch. push_config.batch is deprecated in favor of batch")
	}

	exporters, err := c.exporters()
	if err != nil {
		return nil, err
	}
	exportersNames := make([]string, 0, len(exporters))
	for name := range exporters {
		exportersNames = append(exportersNames, name)
	}

	// processors
	processors := map[string]interface{}{}
	processorNames := []string{}
	if c.ScrapeConfigs != nil {
		processorNames = append(processorNames, promsdprocessor.TypeStr)
		processors[promsdprocessor.TypeStr] = map[string]interface{}{
			"scrape_configs": c.ScrapeConfigs,
		}
	}

	if c.Attributes != nil {
		processors["attributes"] = c.Attributes
		processorNames = append(processorNames, "attributes")
	}

	if c.Batch != nil {
		processors["batch"] = c.Batch
		processorNames = append(processorNames, "batch")
	} else if c.PushConfig.Batch != nil {
		processors["batch"] = c.PushConfig.Batch
		processorNames = append(processorNames, "batch")
	}

	if c.SpanMetrics != nil {
		// Configure the metrics exporter.
		namespace := "tempo_spanmetrics"
		if len(c.SpanMetrics.MetricsExporter.Namespace) != 0 {
			namespace = fmt.Sprintf("%s_%s", c.SpanMetrics.MetricsExporter.Namespace, namespace)
		}

		exporters[defaultSpanMetricsExporter] = map[string]interface{}{
			"endpoint":        c.SpanMetrics.MetricsExporter.Endpoint,
			"namespace":       namespace,
			"const_labels":    c.SpanMetrics.MetricsExporter.ConstLabels,
			"send_timestamps": c.SpanMetrics.MetricsExporter.SendTimestamps,
		}

		processorNames = append(processorNames, "spanmetrics")
		processors["spanmetrics"] = map[string]interface{}{
			"metrics_exporter":          defaultSpanMetricsExporter,
			"latency_histogram_buckets": c.SpanMetrics.LatencyHistogramBuckets,
			"dimensions":                c.SpanMetrics.Dimensions,
		}
	}

	// receivers
	receiverNames := []string{}
	for name := range c.Receivers {
		receiverNames = append(receiverNames, name)
	}

	if c.TailSampling != nil {
		wait := defaultDecisionWait
		if c.TailSampling.DecisionWait != 0 {
			wait = c.TailSampling.DecisionWait
		}

		policies, err := formatPolicies(c.TailSampling.Policies)
		if err != nil {
			return nil, err
		}

		// tail_sampling should be executed before the batch processor
		// TODO(mario.rodriguez): put attributes processor before tail_sampling. Maybe we want to sample on mutated spans
		processorNames = append([]string{"tail_sampling"}, processorNames...)
		processors["tail_sampling"] = map[string]interface{}{
			"policies":      policies,
			"decision_wait": wait,
		}

		internalExporter, err := c.loadBalancingExporter()
		if err != nil {
			return nil, err
		}
		exporters["loadbalancing"] = internalExporter

		c.Receivers["otlp/lb"] = map[string]interface{}{
			"protocols": map[string]interface{}{
				"grpc": map[string]interface{}{
					"endpoint": net.JoinHostPort("0.0.0.0", defaultLoadBalancingPort),
				},
			},
		}
	}

	pipelines := make(map[string]interface{})
	if c.TailSampling != nil {
		// load balancing pipeline
		pipelines["traces/0"] = map[string]interface{}{
			"receivers": receiverNames,
			"exporters": []string{"loadbalancing"},
		}
		// processing pipeline
		pipelines["traces/1"] = map[string]interface{}{
			"exporters":  exportersNames,
			"processors": processorNames,
			"receivers":  []string{"otlp/lb"},
		}
	} else {
		pipelines["traces"] = map[string]interface{}{
			"exporters":  exportersNames,
			"processors": processorNames,
			"receivers":  receiverNames,
		}
	}

	if c.SpanMetrics != nil {
		// Insert a noop receiver in the metrics pipeline.
		// Added to pass validation requiring at least one receiver in a pipeline.
		c.Receivers[noopreceiver.TypeStr] = nil

		pipelines[spanMetricsPipelineName] = map[string]interface{}{
			"receivers": []string{noopreceiver.TypeStr},
			"exporters": []string{defaultSpanMetricsExporter},
		}
	}

	otelMapStructure["exporters"] = exporters
	otelMapStructure["processors"] = processors
	otelMapStructure["receivers"] = c.Receivers

	// pipelines
	otelMapStructure["service"] = map[string]interface{}{
		"pipelines": pipelines,
	}

	// now build the otel configmodel from the mapstructure
	v := viper.New()
	err = v.MergeConfigMap(otelMapStructure)
	if err != nil {
		return nil, fmt.Errorf("failed to merge in mapstructure config: %w", err)
	}

	factories, err := tracingFactories()
	if err != nil {
		return nil, fmt.Errorf("failed to create factories: %w", err)
	}

	otelCfg, err := config.Load(v, factories)
	if err != nil {
		return nil, fmt.Errorf("failed to load OTel config: %w", err)
	}

	return otelCfg, nil
}

// tracingFactories() only creates the needed factories.  if we decide to add support for a new
// processor, exporter, receiver we need to add it here
func tracingFactories() (component.Factories, error) {
	extensions, err := component.MakeExtensionFactoryMap()
	if err != nil {
		return component.Factories{}, err
	}

	receivers, err := component.MakeReceiverFactoryMap(
		jaegerreceiver.NewFactory(),
		zipkinreceiver.NewFactory(),
		otlpreceiver.NewFactory(),
		opencensusreceiver.NewFactory(),
		kafkareceiver.NewFactory(),
		noopreceiver.NewFactory(),
	)
	if err != nil {
		return component.Factories{}, err
	}

	exporters, err := component.MakeExporterFactoryMap(
		otlpexporter.NewFactory(),
		prometheusexporter.NewFactory(),
		loadbalancingexporter.NewFactory(),
	)
	if err != nil {
		return component.Factories{}, err
	}

	processors, err := component.MakeProcessorFactoryMap(
		batchprocessor.NewFactory(),
		attributesprocessor.NewFactory(),
		promsdprocessor.NewFactory(),
		spanmetricsprocessor.NewFactory(),
		tailsamplingprocessor.NewFactory(),
	)
	if err != nil {
		return component.Factories{}, err
	}

	return component.Factories{
		Extensions: extensions,
		Receivers:  receivers,
		Processors: processors,
		Exporters:  exporters,
	}, nil
}
