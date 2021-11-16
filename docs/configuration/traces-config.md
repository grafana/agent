+++
title = "traces_config"
weight = 400
aliases = ["/docs/agent/latest/configuration/tempo-config/"]
+++

# traces_config

The `traces_config` block configures a set of Tempo instances, each of which
configures its own tracing pipeline. Having multiple configs allows you to
configure multiple distinct pipelines, each of which collects spans and sends
them to a different location.

Note that if using multiple configs, you must manually set port numbers for
each receiver, otherwise they will all try to use the same port and fail to
start.

```yaml
configs:
 - [<traces_instance_config>]
 ```

## traces_instance_config

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
  # Here must be the port of gRPC receiver, not the Tempo default port.
  # Example for cloud instances:  `tempo-us-central1.grafana.net:443`
  # For local / on-premises instances: `localhost:55680` or `tempo.example.com:14250`
  # Note: for non-encrypted connections you must also set `insecure: true`
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

# This processor writes a well formatted log line to a logs instance for each span, root, or process
# that passes through the Agent. This allows for automatically building a mechanism for trace
# discovery and building metrics from traces using Loki. It should be considered experimental.
automatic_logging:
  # Indicates where the stream of log lines should go. Either supports writing
  # to a logs instance defined in this same config or to stdout.
  [ backend: <string> | default = "stdout" | supported "stdout", "logs_instance" ]
  # Indicates the logs instance to write logs to.
  # Required if backend is set to logs_instance.
  [ logs_instance_name: <string> ]
  # Log one line per span. Warning! possibly very high volume
  [ spans: <boolean> ]
  # Log one line for every root span of a trace.
  [ roots: <boolean> ]
  # Log one line for every process
  [ processes: <boolean> ]
  # Additional span attributes to log
  [ span_attributes: <string array> ]
  # Additional process attributes to log
  [ process_attributes: <string array> ]
  # Timeout on writing logs to Loki when backend is "logs_instance."
  [ timeout: <duration> | default = 1ms ]
  # Configures a set of key values that will be logged as labels
  # They need to be span or process attributes logged in the log line
  #
  # This feature only applies when `backend = logs_instance`
  #
  # Loki only accepts alphanumeric and "_" as valid characters for labels.
  # Labels are sanitized by replacing invalid characters with underscores.
  [ labels: <string array> ]
  overrides:
    [ logs_instance_tag: <string> | default = "traces" ]
    [ service_key: <string> | default = "svc" ]
    [ span_name_key: <string> | default = "span" ]
    [ status_key: <string> | default = "status" ]
    [ duration_key: <string> | default = "dur" ]
    [ trace_id_key: <string> | default = "tid" ]

# Receiver configurations are mapped directly into the OpenTelemetry receivers
# block. At least one receiver is required.
# The Agent uses OpenTelemetry v0.36.0. Refer to the corresponding receiver's config.
#
# Supported receivers: otlp, jaeger, kafka, opencensus and zipkin.
receivers: <receivers>

# A list of prometheus scrape configs.  Targets discovered through these scrape
# configs have their __address__ matched against the ip on incoming spans. If a
# match is found then relabeling rules are applied.
scrape_configs:
  - [<scrape_config>]
# Defines what method is used when adding k/v to spans.
# Options are `update`, `insert` and `upsert`.
# `update` only modifies an existing k/v and `insert` only appends if the k/v
# is not present. `upsert` does both.
[ prom_sd_operation_type: <string> | default = "upsert" ]

# spanmetrics supports aggregating Request, Error and Duration (R.E.D) metrics
# from span data.
#
# spanmetrics generates two metrics from spans and uses remote_write or
# OpenTelemetry Prometheus exporters to serve the metrics locally.
#
# In order to use the remote_write exporter, you have to configure a Prometheus
# instance in the Agent and pass its name to the `metrics_instance` field.
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

  # Metrics are namespaced to `traces_spanmetrics` by default.
  # They can be further namespaced, i.e. `{namespace}_traces_spanmetrics`
  [ namespace: <string> ]

  # metrics_instance is the metrics instance used to remote write metrics.
  [ metrics_instance: <string> ]
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
# Tail sampling also supports multi agent deployments, allowing to group all
# spans of a trace in the same agent by load balancing the spans by trace ID
# between the instances.
# * To make use of this feature, check load_balancing below *
tail_sampling:
  # policies define the rules by which traces will be sampled. Multiple policies
  # can be added to the same pipeline.
  policies:
    - [<tailsamplingprocessor.policies>]

  # Time that to wait before making a decision for a trace.
  # Longer wait times reduce the probability of sampling an incomplete trace at
  # the cost of higher memory usage.
  decision_wait: [ <duration> | default="5s" ]

# load_balancing configures load balancing of spans across multi agent deployments.
# It ensures that all spans of a trace are sampled in the same instance.
# It works by exporting spans based on their traceID via consistent hashing.
#
# Enabling this feature is required for tail_sampling to correctly work when
# different agent instances can receive spans for the same trace.
#
# Load balancing works by layering two pipelines and consistently exporting
# spans belonging to a trace to the same agent instance.
# Agent instances need to be able to communicate with each other via gRPC.
#
# Load balancing significantly increases CPU usage. This is because spans are
# exported an additional time between agents.
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

# service_graphs configures processing of traces for building service graphs in
# the form of prometheus metrics. The generated metrics represent edges between
# nodes in the graph. Nodes are represented by `client` and `server` labels.
#
#  e.g. tempo_service_graph_request_total{client="app", server="db"} 20
#
# Service graphs works by inspecting spans and looking for the tag `span.kind`.
# If it finds the span kind to be client or server, it stores the request in a
# local in-memory store.
#
# That request waits until its corresponding client or server pair span is
# processed or until the maximum waiting time has passed.
# When either of those conditions is reached, the request is processed and
# removed from the local store. If the request is complete by that time, it'll
# be recorded as an edge in the graph.
#
# Service graphs supports multi-agent deployments, allowing to group all spans
# of a trace in the same agent by load balancing the spans by trace ID between
# the instances.
# * To make use of this feature, check load_balancing above *
service_graphs:
  [ enabled: <bool> | default = false ]

  # configures the time the processor will wait since a span is consumed until
  # it's considered expired if its paired has not been processed.
  #
  # increasing the waiting time will increase the percentage of paired spans.
  # retaining unpaired spans for longer will make reaching max_items more likely.
  [ wait: <duration> | default = "10s"]

  # configures the max amount of edges that will be stored in memory.
  #
  # spans that arrive to the processor that do not pair with an already
  # processed span are dropped.
  #
  # a higher max number of items increases the max throughput of processed spans
  # with a higher memory consumption.
  [ max_items: <integer> | default = 10_000 ]
  
  # configures the number of workers that will process completed edges concurrently.
  # as edges are completed, they get queued to be collected as metrics for the graph.
  [ workers: <integer> | default = 10]

  # configures what status codes are considered as successful (e.g. HTTP 404).
  #
  # by default, a request is considered failed in the following cases:
  #   1. HTTP status is not 2XX
  #   1. gRPC status code is not OK
  #   1. span status is Error
  success_codes:
    # http status codes not to be considered as failure
    http:
      [ - <int> ... ]
    # grpc status codes not to be considered as failure
    grpc:
      [ - <int> ... ]
```

> **Note:** More information on the following types can be found on the
> documentation for their respective projects:
>
* [`attributes.config`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/b2327211df976e0a57ef0425493448988772a16b/processor/attributesprocessor)
* [`batch.config`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/tree/1f5dd9f9a566a937ec15093ca3bc377fba86f5f9/processor/batchprocessor)
* [`otlpexporter.sending_queue`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/tree/1f5dd9f9a566a937ec15093ca3bc377fba86f5f9/exporter/otlpexporter)
* [`otlpexporter.retry_on_failure`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/tree/1f5dd9f9a566a937ec15093ca3bc377fba86f5f9/exporter/otlpexporter)
* `receivers`:
  * [`jaegerreceiver`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/b2327211df976e0a57ef0425493448988772a16b/receiver/jaegerreceiver)
  * [`kafkareceiver`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/b2327211df976e0a57ef0425493448988772a16b/receiver/kafkareceiver)
  * [`otlpreceiver`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/tree/1f5dd9f9a566a937ec15093ca3bc377fba86f5f9/receiver/otlpreceiver)
  * [`opencensusreceiver`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/b2327211df976e0a57ef0425493448988772a16b/receiver/opencensusreceiver)
  * [`zipkinreceiver`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/b2327211df976e0a57ef0425493448988772a16b/receiver/zipkinreceiver)
* [`scrape_config`: Prometheus](https://prometheus.io/docs/prometheus/2.27/configuration/configuration/#scrape_config)
* [`spanmetricsprocessor.latency_histogram_buckets`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/b2327211df976e0a57ef0425493448988772a16b/processor/spanmetricsprocessor/config.go#L38-L47)
* [`spanmetricsprocessor.dimensions`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/b2327211df976e0a57ef0425493448988772a16b/processor/spanmetricsprocessor/config.go#L38-L47)
* [`tailsamplingprocessor.policies`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/b2327211df976e0a57ef0425493448988772a16b/processor/tailsamplingprocessor)
