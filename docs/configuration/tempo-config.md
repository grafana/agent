+++
title = "tempo_config"
weight = 400
+++

# tempo_config

The `tempo_config` block configures a set of Tempo instances, each of which
configures its own tracing pipeline. Having multiple configs allows you to
configure multiple distinct pipelines, each of which collects spans and sends
them to a different location.

Note that if using multiple configs, you must manually set port numbers for
each receiver, otherwise they will all try to use the same port and fail to
start.

```yaml
configs:
 - [<tempo_instance_config>]
 ```

## tempo_instance_config

```yaml
# Name configures the name of this Tempo instance. Names must be non-empty and
# unique across all Tempo instances. The value of the name here will appear in
# logs and as a label on metrics.
name: <string>

# This field allows for the general manipulation of tags on spans that pass
# through this agent. A common use may be to add an environment or cluster
# variable.
[attributes: <attributes.config>]

# This field allows to configure grouping spans into batches. Batching helps
# better compress the data and reduce the number of outgoing connections
# required transmit the data.
[batch: <batch.config>]

remote_write:
  # host:port to send traces to
  - endpoint: <string>

    # Custom HTTP headers to be sent along with each remote write request.
    # Be aware that 'authorization' header will be overwritten in presence
    # of basic_auth.
    headers:
      [ <string>: <string> ... ]

    # Controls whether compression is enabled.
    [ compression: <string> | default = "gzip" | supported = "none", "gzip"]

    # Controls what protocol to use when exporting traces.
    # Only "grpc" is supported in Grafana Cloud.
    [ protocol: <string> | default = "grpc" | supported = "grpc", "http" ]

    # Controls whether or not TLS is required.  See https://godoc.org/google.golang.org/grpc#WithInsecure
    [ insecure: <boolean> | default = false ]

    # Deprecated in favor of tls_config
    # If both `insecure_skip_verify` and `tls_config.insecure_skip_verify` are used,
    # the latter take precedence.
    [ insecure_skip_verify: <bool> | default = false ]

    # Controls TLS settings of the exporter's client. See https://github.com/open-telemetry/opentelemetry-collector/blob/v0.21.0/config/configtls/README.md
    # This should be used only if `insecure` is set to false
    tls_config:
      # Path to the CA cert. For a client this verifies the server certificate. If empty uses system root CA.
      [ca_file: <string>]
      # Path to the TLS cert to use for TLS required connections
      [cert_file: <string>]
      # Path to the TLS key to use for TLS required connections
      [key_file: <string>]
      # Disable validation of the server certificate.
      [ insecure_skip_verify: <bool> | default = false ]

    # Sets the `Authorization` header on every trace push with the
    # configured username and password.
    # password and password_file are mutually exclusive.
    basic_auth:
      [ username: <string> ]
      [ password: <secret> ]
      [ password_file: <string> ]

    [ sending_queue: <otlpexporter.sending_queue> ]
    [ retry_on_failure: <otlpexporter.retry_on_failure> ]

# This processor writes a well formatted log line to a Loki instance for each span, root, or process
# that passes through the Agent. This allows for automatically building a mechanism for trace
# discovery and building metrics from traces using Loki. It should be considered experimental.
automatic_logging:
  # indicates where the stream of log lines should go. Either supports writing to a loki instance defined in this same config or to stdout.
  [ backend: <string> | default = "stdout" | supported "stdout", "loki" ]
  # indicates the Loki instance to write logs to. Required if backend is set to loki.
  [ loki_name: <string> ]
  # log one line per span. Warning! possibly very high volume
  [ spans: <boolean> ]
  # log one line for every root span of a trace.
  [ roots: <boolean> ]
  # log one line for every process
  [ processes: <boolean> ]
  # additional span attributes to log
  [ span_attributes: <string array> ]
  # additional process attributes to log
  [ process_attributes: <string array> ]
  # timeout on sending logs to Loki
  [ timeout: <duration> | default = 1ms ]
  overrides:
    [ loki_tag: <string> | default = "tempo" ]
    [ service_key: <string> | default = "svc" ]
    [ span_name_key: <string> | default = "span" ]
    [ status_key: <string> | default = "status" ]
    [ duration_key: <string> | default = "dur" ]
    [ trace_id_key: <string> | default = "tid" ]

# Receiver configurations are mapped directly into the OpenTelemetry receivers
# block. At least one receiver is required.
#
# Supported receivers: otlp, jaeger, kafka, opencensus and zipkin.
receivers: <receivers>

# A list of prometheus scrape configs.  Targets discovered through these scrape
# configs have their __address__ matched against the ip on incoming spans. If a
# match is found then relabeling rules are applied.
scrape_configs:
  - [<scrape_config>]

# spanmetrics supports aggregating Request, Error and Duration (R.E.D) metrics
# from span data.
#
# spanmetrics generates two metrics from spans and uses remote_write or
# OpenTelemetry Prometheus exporters to serve the metrics locally.
#
# In order to use the remote_write exporter, you have to configure a Prometheus
# instance in the Agent and pass its name to the `prom_instance` field.
#
# If you want to use the OpenTelemetry Prometheus exporter, you have to
# configure handler_endpoint and then scrape that endpoint.
#
# The first generated metric is `calls`, a counter to compute requests.
# The second generated metric is `latency`, a histogram to compute the
# operation's duration.
#
# If you want to rename the generated metrics, you can configure the `namespace`
# option of prometheus exporter.
#
# This is an experimental feature of Opentelemetry-Collector and the behavior
# may change in the future.
spanmetrics:
  # latency_histogram_buckets and dimensions are the same as the configs in
  # spanmetricsprocessor.
  [ latency_histogram_buckets: <spanmetricsprocessor.latency_histogram_buckets> ]
  [ dimensions: <spanmetricsprocessor.dimensions> ]

  # const_labels are labels that will always get applied to the exported
  # metrics.
  const_labels:
    [ <string>: <string>... ]

  # Metrics are namespaced to `tempo_spanmetrics` by default.
  # They can be further namespaced, i.e. `{namespace}_tempo_spanmetrics`
  [ namespace: <string> ]

  # prom_instance is the prometheus used to remote write metrics.
  [ prom_instance: <string> ]
  # handler_endpoint defines the endpoint where the OTel prometheus exporter will be exposed.
  [ handler_endpoint: <string> ]

# tail_sampling supports tail-based sampling of traces in the agent.
#
# Policies can be defined that determine what traces are sampled and sent to the
# backends and what traces are dropped.
#
# In order to make a correct sampling decision it's important that the agent has
# a complete trace. This is achieved by waiting a given time for all the spans
# before evaluating the trace.
#
# Tail sampling also supports multiple agent deployments, allowing to group all
# spans of a trace in the same agent by load balancing the spans by trace ID
# between the instances.
tail_sampling:
  # policies define the rules by which traces will be sampled. Multiple policies
  # can be added to the same pipeline.
  policies:
    - [<tailsamplingprocessor.policies>]

  # Time that to wait before making a decision for a trace.
  # Longer wait times reduce the probability of sampling an incomplete trace at
  # the cost of higher memory usage.
  decision_wait: [ <duration> | default="5s" ]

  # load_balancing configures load balancing of spans across multiple agents.
  # It ensures that all spans of a trace are sampled in the same instance.
  # Only necessary if more than one agent process is receiving traces.
  load_balancing:
    # resolver configures the resolution strategy for the involved backends
    # It can be static, with a fixed list of hostnames, or DNS, with a hostname
    # (and port) that will resolve to all IP addresses.
    resolver:
      static:
        hostnames:
          [ - <string> ... ]
      dns:
        hostname: <string>
        [ port: <int> ]

    # Load balancing is done via an otlp exporter.
    # The remaining configuration is common with the remote_write block.
    exporter:
      # Controls whether compression is enabled.
      [ compression: <string> | default = "gzip" | supported = "none", "gzip"]

      # Controls whether or not TLS is required.
      [ insecure: <boolean> | default = false ]

      # Disable validation of the server certificate. Only used when insecure is set
      # to false.
      [ insecure_skip_verify: <bool> | default = false ]

      # Sets the `Authorization` header on every trace push with the
      # configured username and password.
      # password and password_file are mutually exclusive.
      basic_auth:
        [ username: <string> ]
        [ password: <secret> ]
        [ password_file: <string> ]
```

> **Note:** More information on the following types can be found on the
> documentation for their respective projects:
>
> * [`attributes.config`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/processor/attributesprocessor)
> * [`batch.config`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/processor/batchprocessor)
> * [`otlpexporter.sending_queue`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/exporter/otlpexporter)
> * [`otlpexporter.retry_on_failure`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/exporter/otlpexporter)
> * [`receivers`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/receiver/README.md)
> * [`scrape_config`: Prometheus](https://prometheus.io/docs/prometheus/2.27/configuration/configuration/#scrape_config)
> * [`spanmetricsprocessor.latency_histogram_buckets`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.21.0/processor/spanmetricsprocessor/config.go#L38-L47)
> * [`spanmetricsprocessor.dimensions`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.21.0/processor/spanmetricsprocessor/config.go#L38-L47)
> * [`tailsamplingprocessor.policies`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/tailsamplingprocessor)
