---
aliases:
- ../../configuration/tempo-config/
- ../../configuration/traces-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/traces-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/traces-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/traces-config/
description: Learn about traces_config
title: traces_config
weight: 400
---

# traces_config

The `traces_config` block configures a set of Tempo instances, each of which
configures its own tracing pipeline. Having multiple configs allows you to
configure multiple distinct pipelines, each of which collects spans and sends
them to a different location.

{{< admonition type="note" >}}
If you are using multiple configs, you must manually set port numbers for
each receiver, otherwise they will all try to use the same port and fail to
start.
{{< /admonition >}}

```yaml
configs:
 [ - <traces_instance_config> ... ]
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
[ attributes: <attributes.config> ]

# This field allows to configure grouping spans into batches. Batching helps
# better compress the data and reduce the number of outgoing connections
# required transmit the data.
[ batch: <batch.config> ]

remote_write:
  # host:port to send traces to.
  # Here must be the port of gRPC receiver, not the Tempo default port.
  # Example for cloud instances: `tempo-us-central1.grafana.net:443`
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

    # Controls what format to use when exporting traces, in combination with protocol.
    # protocol/format supported combinations are grpc/otlp and http/otlp.
    # Only grpc/otlp is supported in Grafana Cloud.
    [ format: <string> | default = "otlp" | supported = "otlp" ]

    # Controls whether or not TLS is required. See https://godoc.org/google.golang.org/grpc#WithInsecure
    [ insecure: <boolean> | default = false ]

    # Deprecated in favor of tls_config
    # If both `insecure_skip_verify` and `tls_config.insecure_skip_verify` are used,
    # the latter take precedence.
    [ insecure_skip_verify: <bool> | default = false ]

    # Configures opentelemetry exporters to use the OpenTelemetry auth extension `oauth2clientauthextension`.
    # Can not be used in combination with `basic_auth`.
    # See https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/{{< param "OTEL_VERSION" >}}/extension/oauth2clientauthextension/README.md
    oauth2:
      # Configures the TLS settings specific to the oauth2 client
      # The client identifier issued to the oauth client
      [ client_id: <string> ]
      # The secret string associated with the oauth client
      [ client_secret: <string> ]
      # Additional parameters for requests to the token endpoint
      [ endpoint_params: <string> ]
      # The resource server's token endpoint URL
      [ token_url: <string> ]
      # Optional, requested permissions associated with the oauth client
      [ scopes: [<string>] ]
      # Optional, specifies the timeout fetching tokens from the token_url. Default: no timeout
      [ timeout: <duration> ]
      # TLS client configuration for the underneath client to authorization server.
      # https://github.com/open-telemetry/opentelemetry-collector/blob/{{< param "OTEL_VERSION" >}}/config/configtls/README.md
      tls:
        # Disable validation of the server certificate.
        [ insecure: <bool> | default = false ]
        # InsecureSkipVerify will enable TLS but not verify the certificate.
        [ insecure_skip_verify: <bool> | default = false ]
        # ServerName requested by client for virtual hosting.
        # This sets the ServerName in the TLSConfig. Please refer to
        # https://godoc.org/crypto/tls#Config for more information.
        [ server_name_override: <string> ]
        # Path to the CA cert. For a client this verifies the server certificate. If empty uses system root CA.
        [ ca_file: <string> ]
        # In memory PEM encoded cert.
        [ ca_pem: <string> ]
        # Path to the TLS cert to use for TLS required connections
        [ cert_file: <string> ]
        # In memory PEM encoded TLS cert to use for TLS required connections.
        [ cert_pem: <string> ]
        # Path to the TLS key to use for TLS required connections
        [ key_file: <string> ]
        # In memory PEM encoded TLS key to use for TLS required connections.
        [ key_pem: <string> ]
        # Minimum acceptable TLS version.
        [ min_version: <string> | default = "1.2" ]
        # Maximum acceptable TLS version.
        # If not set, it is handled by crypto/tls - currently it is "1.3".
        [ max_version: <string> | default = "" ]
        # ReloadInterval specifies the duration after which the certificate will be reloaded.
        # If not set, it will never be reloaded.
        [ reload_interval: <duration> ]

    # Controls TLS settings of the exporter's client:
    # https://prometheus.io/docs/prometheus/2.45/configuration/configuration/#tls_config
    # This should be used only if `insecure` is set to false
    tls_config:
      # Path to the CA cert. For a client this verifies the server certificate. If empty uses system root CA.
      [ ca_file: <string> ]
      # Path to the TLS cert to use for TLS required connections
      [ cert_file: <string> ]
      # Path to the TLS key to use for TLS required connections
      [ key_file: <string> ]
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
  [ backend: <string> | default = "stdout" | supported = "stdout", "logs_instance" ]
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
# The Agent uses OpenTelemetry {{< param "OTEL_VERSION" >}}. Refer to the corresponding receiver's config.
#
# Supported receivers: otlp, jaeger, kafka, opencensus and zipkin.
receivers: <receivers>

# A list of prometheus scrape configs.  Targets discovered through these scrape
# configs have their __address__ matched against the ip on incoming spans. If a
# match is found then relabeling rules are applied.
scrape_configs:
  [ - <scrape_config> ... ]
# Defines what method is used when adding k/v to spans.
# Options are `update`, `insert` and `upsert`.
# `update` only modifies an existing k/v and `insert` only appends if the k/v
# is not present. `upsert` does both.
[ prom_sd_operation_type: <string> | default = "upsert" ]
# Configures what methods to use to do association between spans and pods.
# PromSD processor matches the IP address of the metadata labels from the k8s API
# with the IP address obtained from the specified pod association method.
# If a match is found then the span is labeled.
#
# Options are `ip`, `net.host.ip`, `k8s.pod.ip`, `hostname` and `connection`.
#   - `ip`, `net.host.ip` and `k8s.pod.ip`, `hostname` match spans tags.
#   - `connection` inspects the context from the incoming requests (gRPC and HTTP).
#
# Tracing instrumentation is commonly the responsible for tagging spans
# with IP address to the labels mentioned above.
# If running on kubernetes, `k8s.pod.ip` can be automatically attached via the
# downward API. For example, if you're using OTel instrumentation libraries, set
# OTEL_RESOURCE_ATTRIBUTES=k8s.pod.ip=$(POD_IP) to inject spans with the sender
# pod's IP.
#
# By default, all methods are enabled, and evaluated in the order specified above.
# Order of evaluation is honored when multiple methods are enabled.
prom_sd_pod_associations:
  [ - <string> ... ]

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
    [ <string>: <string> ... ]
  # Metrics are namespaced to `traces_spanmetrics` by default.
  # They can be further namespaced, i.e. `{namespace}_traces_spanmetrics`
  [ namespace: <string> ]
  # metrics_instance is the metrics instance used to remote write metrics.
  [ metrics_instance: <string> ]
  # handler_endpoint defines the endpoint where the OTel prometheus exporter will be exposed.
  [ handler_endpoint: <string> ]
  # dimensions_cache_size defines the size of cache for storing Dimensions.
  [ dimensions_cache_size: <int> | default = 1000 ]
  # aggregation_temporality configures whether to reset the metrics after flushing.
  # It can be either AGGREGATION_TEMPORALITY_CUMULATIVE or AGGREGATION_TEMPORALITY_DELTA.
  [ aggregation_temporality: <string> | default = "AGGREGATION_TEMPORALITY_CUMULATIVE" ]
  # metrics_flush_interval configures how often to flush generated metrics.
  [ metrics_flush_interval: <duration> | default = 15s ]

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
    [ - <tailsamplingprocessor.policies> ... ]

  # Time that to wait before making a decision for a trace.
  # Longer wait times reduce the probability of sampling an incomplete trace at
  # the cost of higher memory usage.
  [ decision_wait: <duration> | default = 5s ]

  # Optional, number of traces kept in memory
  [ num_traces: <int> | default = 50000 ]

  # Optional, expected number of new traces (helps in allocating data structures)
  [ expected_new_traces_per_sec: <int> | default = 0 ]

# load_balancing configures load balancing of spans across multi agent deployments.
# It ensures that all spans of a trace are sampled in the same instance.
# It works by exporting spans based on their traceID via consistent hashing.
#
# Enabling this feature is required for "tail_sampling", "spanmetrics", and "service_graphs"
# to correctly work when spans are ingested by multiple agent instances.
#
# Load balancing works by layering two pipelines and consistently exporting
# spans belonging to a trace to the same agent instance.
# Agent instances need to be able to communicate with each other via gRPC.
#
# When load_balancing is enabled:
# 1. When an Agent receives spans from the configured "receivers".
# 2. If the "attributes" processor is configured, it will run through all the spans.
# 3. The spans will be exported using the "load_balancing" configuration to any of the Agent instances.
#    This may or may not be the same Agent which has already received the span.
# 4. The Agent which received the span from the loadbalancer will run these processors, 
#    in this order, if they are configured:
#    1. "spanmetrics"
#    2. "service_graphs"
#    3. "tail_sampling"
#    4. "automatic_logging"
#    5. "batch"
# 5. The spans are then remote written using the "remote_write" configuration.
# 
# Load balancing significantly increases CPU usage. This is because spans are
# exported an additional time between agents.
load_balancing:
  # resolver configures the resolution strategy for the involved backends
  # It can be either "static", "dns" or "kubernetes".
  resolver:
    static:
      # A fixed list of hostnames.
      hostnames:
        [ - <string> ... ]
    dns:
      # DNS hostname from which to resolve IP addresses.
      hostname: <string>
      # Port number to use with the resolved IP address when exporting spans.
      [ port: <int> | default = 4317 ]
      # Resolver interval
      [ interval: <duration> | default = 5s ]
      # Resolver timeout
      [ timeout: <duration> | default = 1s ]
    # The kubernetes resolver receives IP addresses of a Kubernetes service 
    # from the Kubernetes API. It does not require polling. The Kubernetes API
    # notifies the Agent when a new pod is available and when an old pod has exited.
    #
    # For the kubernetes resolver to work, Agent must be running under
    # a system account with "list", "watch" and "get" permissions.
    kubernetes:
      service: <string>
      [ ports: <int array> | default = 4317 ]

  # routing_key can be either "traceID" or "service":
  # * "service": exports spans based on their service name.
  # * "traceID": exports spans based on their traceID.
  [ routing_key: <string> | default = "traceID" ]

  # receiver_port is the port the instance will use to receive load balanced traces
  receiver_port: [ <int> | default = 4318 ]

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
  [ wait: <duration> | default = 10s ]

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
  [ workers: <integer> | default = 10 ]

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

# jaeger_remote_sampling configures one or more jaeger remote sampling extensions.
# For more details about the configuration please consult the OpenTelemetry documentation:
# https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/{{< param "OTEL_VERSION" >}}/extension/jaegerremotesampling
#
# Example config:
#
# jaeger_remote_sampling:
#   - source:
#       remote:
#         endpoint: jaeger-collector:14250
#         tls:
#           insecure: true
#   - source:
#       reload_interval: 1s
#       file: /etc/otelcol/sampling_strategies.json
#
jaeger_remote_sampling:
  [ - <jaeger_remote_sampling> ... ]
```

More information on the following types can be found on the documentation for their respective projects:

* [`attributes.config`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/{{< param "OTEL_VERSION" >}}/processor/attributesprocessor)
* [`batch.config`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/tree/{{< param "OTEL_VERSION" >}}/processor/batchprocessor)
* [`otlpexporter.sending_queue`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/tree/{{< param "OTEL_VERSION" >}}/exporter/otlpexporter)
* [`otlpexporter.retry_on_failure`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/tree/{{< param "OTEL_VERSION" >}}/exporter/otlpexporter)
* `receivers`:
  * [`jaegerreceiver`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/{{< param "OTEL_VERSION" >}}/receiver/jaegerreceiver)
  * [`kafkareceiver`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/{{< param "OTEL_VERSION" >}}/receiver/kafkareceiver)
  * [`otlpreceiver`: OpenTelemetry-Collector](https://github.com/open-telemetry/opentelemetry-collector/tree/{{< param "OTEL_VERSION" >}}/receiver/otlpreceiver)
  * [`opencensusreceiver`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/{{< param "OTEL_VERSION" >}}/receiver/opencensusreceiver)
  * [`zipkinreceiver`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/{{< param "OTEL_VERSION" >}}/receiver/zipkinreceiver)
* [`scrape_config`: Prometheus](https://prometheus.io/docs/prometheus/2.45/configuration/configuration/#scrape_config)
* [`spanmetricsprocessor.latency_histogram_buckets`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/{{< param "OTEL_VERSION" >}}/processor/spanmetricsprocessor/config.go#L37-L39)
* [`spanmetricsprocessor.dimensions`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/{{< param "OTEL_VERSION" >}}/processor/spanmetricsprocessor/config.go#L41-L48)
* [`tailsamplingprocessor.policies`: OpenTelemetry-Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/{{< param "OTEL_VERSION" >}}/processor/tailsamplingprocessor)
