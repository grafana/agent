# Configuration Reference

The Grafana Agent is configured in a YAML file (usually called
`agent.yaml`) which contains information on the Grafana Agent and its
Prometheus instances.

- [server_config]({{< relref "./server-config.md" >}})
- [prometheus_config](#prometheus_config)
* [loki_config](#loki_config)
* [tempo_config](#tempo_config)
* [integrations_config](#integrations_config)

## Variable substitution

You can use environment variables in the configuration file to set values that
need to be configurable during deployment. To enable this functionality, you
must pass `-config.expand-env` as a command-line flag to the Agent.

To refer to an environment variable in the config file, use:

```
${VAR}
```

Where VAR is the name of the environment variable.

Each variable reference is replaced at startup by the value of the environment
variable. The replacement is case-sensitive and occurs before the YAML file is
parsed. References to undefined variables are replaced by empty strings unless
you specify a default value or custom error text.

To specify a default value, use:

```
${VAR:-default_value}
```

Where default_value is the value to use if the environment variable is
undefined. The full list of supported syntax can be found at Drone's
[envsubst repository](https://github.com/drone/envsubst).

## Reloading (beta)

The configuration file can be reloaded at runtime. Read the [API
documentation](./api.md#reload-configuration-file-beta) for more information.

This functionality is in beta, and may have issues. Please open GitHub issues
for any problems you encounter.

A reload-only HTTP server can be started to safely reload the system. To start
this, provide `--reload-addr` and `--reload-port` as command line flags.
`reload-port` must be set to a non-zero port to launch the reload server. The
reload server is currently HTTP-only and supports no other options; it does not
read any values from the `server` block in the config file.

While `/-/reload` is enabled on the primary HTTP server, it is not recommended
to use it, since changing the HTTP server configuration will cause it to
restart.

## File Format

To specify which configuration file to load, pass the `-config.file` flag at
the command line. The file is written in the [YAML
format](https://en.wikipedia.org/wiki/YAML), defined by the scheme below.
Brackets indicate that a parameter is optional. For non-list parameters the
value is set to the specified default.

Generic placeholders are defined as follows:

* `<boolean>`: a boolean that can take the values `true` or `false`
* `<int>`: any integer matching the regular expression `[1-9]+[0-9]*`
* `<duration>`: a duration matching the regular expression `[0-9]+(ns|us|µs|ms|[smh])`
* `<labelname>`: a string matching the regular expression `[a-zA-Z_][a-zA-Z0-9_]*`
* `<labelvalue>`: a string of unicode characters
* `<filename>`: a valid path relative to current working directory or an
    absolute path.
* `<host>`: a valid string consisting of a hostname or IP followed by an optional port number
* `<string>`: a regular string
* `<secret>`: a regular string that is a secret, such as a password

Support contents and default values of `agent.yaml`:

```yaml
# Configures the server of the Agent used to enable self-scraping.
[server: <server_config>]

# Configures Prometheus instances.
[prometheus: <prometheus_config>]

# Configures Loki log collection.
[loki: <loki_config>]

# Configures Tempo trace collection.
[tempo: <tempo_config>]

# Configures integrations for the Agent.
[integrations: <integrations_config>]
```

### tempo_config

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

### tempo_instance_config

```yaml
# Name configures the name of this Tempo instance. Names must be non-empty and
# unique across all Tempo instances. The value of the name here will appear in
# logs and as a label on metrics.
name: <string>

# Attributes options: https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/processor/attributesprocessor
#  This field allows for the general manipulation of tags on spans that pass through this agent.  A common use may be to add an environment or cluster variable.
attributes: [attributes.config]

# Batch options: https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/processor/batchprocessor
#  This field allows to configure grouping spans into batches.  Batching helps better compress the data and reduce the number of outgoing connections required to transmit the data.
batch: [batch.config]

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

    # sending_queue and retry_on_failure are the same as: https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/exporter/otlpexporter
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

# Receiver configurations are mapped directly into the OpenTelemetry receivers block.
#   At least one receiver is required. Supported receivers: otlp, jaeger, kafka, opencensus and zipkin.
#   Documentation for each receiver can be found at https://github.com/open-telemetry/opentelemetry-collector/blob/7d7ae2eb34b5d387627875c498d7f43619f37ee3/receiver/README.md
receivers:

# A list of prometheus scrape configs.  Targets discovered through these scrape configs have their __address__ matched against the ip on incoming spans.
# If a match is found then relabeling rules are applied.
scrape_configs:
  - [<scrape_config>]

# spanmetrics supports aggregating Request, Error and Duration (R.E.D) metrics from span data.
# https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.21.0/processor/spanmetricsprocessor/README.md.
# spanmetrics generates two metrics from spans and uses remote write or opentelemetry prometheus exporters to serve the metrics locally.
# In order to use the remote write exporter, you have to configure a tempo instance in the Agent and pass its name in `prom_instance`.
# If you want to use the opentelemetry prometheus exporter, you have to configure handler_endpoint and then scrape that endpoint.
# The first one is `calls` which is a counter to compute requests.
# The second one is `latency` which is a histogram to compute the operations' duration.
# If you want to rename them, you can configure the `namespace` option of prometheus exporter.
# This is an experimental feature of Opentelemetry collector and the behavior may change in the future.
spanmetrics:
  # latency_histogram_buckets and dimensions are the same as the configs in spanmetricsprocessor.
  # https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.21.0/processor/spanmetricsprocessor/config.go#L38-L47
  [ latency_histogram_buckets: <spanmetricsprocessor.latency_histogram_buckets> ]
  [ dimensions: <spanmetricsprocessor.dimensions> ]

  # const_labels are labels that will always get applied to the exported metrics.
  [ const_labels: <prometheus.labels> ]
  # Metrics are namespaced to `tempo_spanmetrics` by default.
  # They can be further namespaced, i.e. `{namespace}_tempo_spanmetrics`
  [ namespace: <string> ]
  # prom_instance is the prometheus used to remote write metrics.
  [ prom_instance: <string> ]
  # handler_endpoint defines the endpoint where the OTel prometheus exporter will be exposed.
  [ handler_endpoint: <string> ]

# tail_sampling supports tail-based sampling of traces in the agent.
# Policies can be defined that determine what traces are sampled and sent to the backends and what traces are dropped.
# In order to make a correct sampling decision it's important that the agent has a complete trace.
# This is achieved by waiting a given time for all the spans before evaluating the trace.
# Tail sampling also supports multiple agent deployments, allowing to group all spans of a trace
# in the same agent by load balancing the spans by trace ID between the instances.
tail_sampling:
  # policies define the rules by which traces will be sampled. Multiple policies can be added to the same pipeline
  # They are the same as the policies in Open-Telemetry's tailsamplingprocessor.
  # https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/tailsamplingprocessor
  policies:
    - [<tailsamplingprocessor.policies>]
  # decision_wait is the time that will be waited before making a decision for a trace.
  # Longer times reduce the probability of sampling an incomplete trace at the cost of higher memory usage.
  decision_wait: [ <string> | default="5s" ]
  # load_balancing configures load balancing of spans across multiple agents.
  # It ensures that all spans of a trace are sampled in the same instance.
  # It's not necessary when only one agent is receiving traces (e.g. single instance deployments).
  load_balancing:
    # resolver configures the resolution strategy for the involved backends
    # It can be static, with a fixed list of hostnames, or DNS, with a hostname (and port) that will resolve to all IP addresses.
    # It's the same as the config in loadbalancingexporter.
    # https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/loadbalancingexporter
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

      # Controls whether or not TLS is required.  See https://godoc.org/google.golang.org/grpc#WithInsecure
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

### integrations_config

The `integrations_config` block configures how the Agent runs integrations that
scrape and send metrics without needing to run specific Prometheus exporters or
manually write `scrape_configs`:

```yaml
# Controls the Agent integration
agent:
  # Enables the Agent integration, allowing the Agent to automatically
  # collect and send metrics about itself.
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the agent integration will be run but not scraped and thus not
  # remote_written. Metrics for the integration will be exposed at
  # /integrations/agent/metrics and can be scraped by an external process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timtout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

# Client TLS Configuration
# Client Cert/Key Values need to be defined if the server is requesting a certificate
#  (Client Auth Type = RequireAndVerifyClientCert || RequireAnyClientCert).
http_tls_config: <tls_config>

# Controls the node_exporter integration
node_exporter: <node_exporter_config>

# Controls the process_exporter integration
process_exporter: <process_exporter_config>

# Controls the mysqld_exporter integration
mysqld_exporter: <mysqld_exporter_config>

# Controls the redis_exporter integration
redis_exporter: <redis_exporter_config>

# Controls the dnsmasq_exporter integration
dnsmasq_exporter: <dnsmasq_exporter_config>

# Controls the elasticsearch_expoter integration
elasticsearch_expoter: <elasticsearch_expoter_config>

# Controls the memcached_exporter integration
memcached_exporter: <memcached_exporter_config>

# Controls the postgres_exporter integration
postgres_exporter: <postgres_exporter_config>

# Controls the statsd_exporter integration
statsd_exporter: <statsd_exporter_config>

# Controls the consul_exporter integration
consul_exporter: <consul_exporter_config>

# Controls the windows_exporter integration
windows_exporter: <windows_exporter_config>

# Automatically collect metrics from enabled integrations. If disabled,
# integrations will be run but not scraped and thus not remote_written. Metrics
# for integrations will be exposed at /integrations/<integration_key>/metrics
# and can be scraped by an external process.
[scrape_integrations: <boolean> | default = true]

# When true, replaces the instance label with the hostname of the machine,
# rather than 127.0.0.1:<server.http_listen_port>. Useful when running multiple
# Agents with the same integrations and uniquely identifying where metrics are
# coming from.
#
# The value for the instance label can be replaced by providing custom
# relabel_configs for an integration. The overwritten instance label will be
# available when relabel_configs run.
[replace_instance_label: <boolean> | default = true]

# When true, adds an agent_hostname label to all samples coming from
# integrations. The value of the agent_hostname label will be the
# value of $HOSTNAME (if available) or the machine's hostname.
#
# DEPRECATED. May be removed in a future version. Rely on
# replace_instance_label instead, since it has better compatability
# with existing dashboards.
[use_hostname_label: <boolean> | default = true]

# Extra labels to add to all samples coming from integrations.
labels:
  { <string>: <string> }

# The period to wait before restarting an integration that exits with an
# error.
[integration_restart_backoff: <duration> | default = "5s"]

# A list of remote_write targets. Defaults to global_config.remote_write.
# If provided, overrides the global defaults.
prometheus_remote_write:
  - [<remote_write>]
```

### node_exporter_config

The `node_exporter_config` block configures the `node_exporter` integration,
which is an embedded version of
[`node_exporter`](https://github.com/prometheus/node_exporter)
and allows for collecting metrics from the UNIX system that `node_exporter` is
running on. It provides a significant amount of collectors that are responsible
for monitoring various aspects of the host system.

Note that if running the Agent in a container, you will need to bind mount
folders from the host system so the integration can monitor them. You can use
the example below, making sure to replace `/path/to/config.yaml` with a path on
your host machine where an Agent configuration file is:

```
docker run \
  --net="host" \
  --pid="host" \
  --cap-add=SYS_TIME \
  -v "/:/host/root:ro,rslave" \
  -v "/sys:/host/sys:ro,rslave" \
  -v "/proc:/host/proc:ro,rslave" \
  -v /tmp/agent:/etc/agent \
  -v /path/to/config.yaml:/etc/agent-config/agent.yaml \
  grafana/agent:v0.15.0 \
  --config.file=/etc/agent-config/agent.yaml
```

Use this configuration file for testing out `node_exporter` support, replacing
the `remote_write` settings with settings appropriate for you:

```yaml
server:
  log_level: info
  http_listen_port: 12345

prometheus:
  wal_directory: /tmp/agent
  global:
    scrape_interval: 15s
    remote_write:
    - url: https://prometheus-us-central1.grafana.net/api/prom/push
      basic_auth:
        username: user-id
        password: api-token

integrations:
  node_exporter:
    enabled: true
    rootfs_path: /host/root
    sysfs_path: /host/sys
    procfs_path: /host/proc
```

For running on Kubernetes, ensure to set the equivalent mounts and capabilities
there as well:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: agent
spec:
  containers:
  - image: grafana/agent:v0.15.0
    name: agent
    args:
    - --config.file=/etc/agent-config/agent.yaml
    securityContext:
      capabilities:
        add: ["SYS_TIME"]
      priviliged: true
      runAsUser: 0
    volumeMounts:
    - name: rootfs
      mountPath: /host/root
      readOnly: true
    - name: sysfs
      mountPath: /host/sys
      readOnly: true
    - name: procfs
      mountPath: /host/proc
      readOnly: true
  hostPID: true
  hostNetwork: true
  dnsPolicy: ClusterFirstWithHostNet
  volumes:
  - name: rootfs
    hostPath:
      path: /
  - name: sysfs
    hostPath:
      path: /sys
  - name: procfs
    hostPath:
      path: /proc
```

The manifest and Tanka configs provided by this repository do not have the
mounts or capabilities required for running this integration.

Some collectors only work on specific operating systems, documented in the
table below. Enabling a collector that is not supported by the operating system
the Agent is running on is a no-op.

| Name             | Description | OS | Enabled by default |
| ---------------- | ----------- | -- | ------------------ |
| arp              | Exposes ARP statistics from /proc/net/arp. | Linux | yes |
| bcache           | Exposes bcache statistics from /sys/fs/bcache. | Linux | yes |
| bonding          | Exposes the number of configured and active slaves of Linux bonding interfaces. | Linux | yes |
| boottime         | Exposes system boot time derived from the kern.boottime sysctl. | Darwin, Dragonfly, FreeBSD, NetBSD, OpenBSD, Solaris | yes |
| btrfs            | Exposes statistics on btrfs. | Linux | yes |
| buddyinfo        | Exposes statistics of memory fragments as reported by /proc/buddyinfo. | Linux | no |
| conntrack        | Shows conntrack statistics (does nothing if no /proc/sys/net/netfilter/ present). | Linux | yes |
| cpu              | Exposes CPU statistics. | Darwin, Dragonfly, FreeBSD, Linux, Solaris | yes |
| cpufreq          | Exposes CPU frequency statistics. | Linux, Solaris | yes |
| devstat          | Exposes device statistics. | Dragonfly, FreeBSD | no |
| diskstats        | Exposes disk I/O statistics. | Darwin, Linux, OpenBSD | yes |
| drbd             | Exposes Distributed Replicated Block Device statistics (to version 8.4). | Linux | no |
| edac             | Exposes error detection and correction statistics. | Linux | yes |
| entropy          | Exposes available entropy. | Linux | yes |
| exec             | Exposes execution statistics. | Dragonfly, FreeBSD | yes |
| filefd           | Exposes file descriptor statistics from /proc/sys/fs/file-nr. | Linux | yes |
| filesystem       | Exposes filesystem statistics, such as disk space used. | Darwin, Dragonfly, FreeBSD, Linux, OpenBSD | yes |
| hwmon            | Exposes hardware monitoring and sensor data from /sys/class/hwmon. | Linux | yes |
| infiniband       | Exposes network statistics specific to InfiniBand and Intel OmniPath configurations. | Linux | yes |
| interrupts       | Exposes detailed interrupts statistics. | Linux, OpenBSD | no |
| ipvs             | Exposes IPVS status from /proc/net/ip_vs and stats from /proc/net/ip_vs_stats. | Linux | yes |
| ksmd             | Exposes kernel and system statistics from /sys/kernel/mm/ksm. | Linux | no |
| loadavg          | Exposes load average. | Darwin, Dragonfly, FreeBSD, Linux, NetBSD, OpenBSD, Solaris | yes |
| logind           | Exposes session counts from logind. | Linux | no |
| mdadm            | Exposes statistics about devices in /proc/mdstat (does nothing if no /proc/mdstat present). | Linux | yes |
| meminfo          | Exposes memory statistics. | Darwin, Dragonfly, FreeBSD, Linux, OpenBSD | yes |
| meminfo_numa     | Exposes memory statistics from /proc/meminfo_numa. | Linux | no |
| mountstats       | Exposes filesystem statistics from /proc/self/mountstats. Exposes detailed NFS client statistics. | Linux | no |
| netclass         | Exposes network interface info from /sys/class/net. | Linux | yes |
| netdev           | Exposes network interface statistics such as bytes transferred. | Darwin, Dragonfly, FreeBSD, Linux, OpenBSD | yes |
| netstat          | Exposes network statistics from /proc/net/netstat. This is the same information as netstat -s. | Linux | yes |
| nfs              | Exposes NFS client statistics from /proc/net/rpc/nfs. This is the same information as nfsstat -c. | Linux | yes |
| nfsd             | Exposes NFS kernel server statistics from /proc/net/rpc/nfsd. This is the same information as nfsstat -s. | Linux | yes |
| ntp              | Exposes local NTP daemon helath to check time. | any | no |
| perf             | Exposes perf based metrics (Warning: Metrics are dependent on kernel configuration and settings). | Linux | no |
| powersupplyclass | Collects information on power supplies. | any | yes |
| pressure         | Exposes pressure stall statistics from /proc/pressure/. | Linux (kernel 4.20+ and/or CONFIG_PSI) | yes |
| processes        | Exposes aggregate process statistics from /proc. | Linux | no |
| qdisc            | Exposes queuing discipline statistics. | Linux | no |
| rapl             | Exposes various statistics from /sys/class/powercap. | Linux | yes |
| runit            | Exposes service status from runit. | any | no |
| schedstat        | Exposes task scheduler statistics from /proc/schedstat. | Linux | yes |
| sockstat         | Exposes various statistics from /proc/net/sockstat. | Linux | yes |
| softnet          | Exposes statistics from /proc/net/softnet_stat. | Linux | yes |
| stat             | Exposes various statistics from /proc/stat. This includes boot time, forks and interrupts. | Linux | yes |
| supervisord      | Exposes service status from supervisord. | any | no |
| systemd          | Exposes service and system status from systemd. | Linux | no |
| tcpstat          | Exposes TCP connection status information from /proc/net/tcp and /proc/net/tcp6. (Warning: the current version has potential performance issues in high load situations). | Linux | no |
| textfile         | Collects metrics from files in a directory matching the filename pattern *.prom. The files must be using the text format defined here: https://prometheus.io/docs/instrumenting/exposition_formats/ | any | yes |
| thermal_zone     | Exposes thermal zone & cooling device statistics from /sys/class/thermal. | Linux | yes |
| time             | Exposes the current system time. | any | yes |
| timex            | Exposes selected adjtimex(2) system call stats. | Linux | yes |
| udp_queues       | Exposes UDP total lengths of the rx_queue and tx_queue from /proc/net/udp and /proc/net/udp6. | Linux | yes |
| uname            | Exposes system information as provided by the uname system call. | Darwin, FreeBSD, Linux, OpenBSD | yes |
| vmstat           | Exposes statistics from /proc/vmstat. | Linux | yes |
| wifi             | Exposes WiFi device and station statistics. | Linux | no |
| xfs              | Exposes XFS runtime statistics. | Linux (kernel 4.4+) | yes |
| zfs              | Exposes ZFS performance statistics. | Linux, Solaris | yes |


```yaml
  # Enables the node_exporter integration, allowing the Agent to automatically
  # collect system metrics from the host UNIX system.
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the node_exporter integration will be run but not scraped and thus not remote-written. Metrics for the
  # integration will be exposed at /integrations/node_exporter/metrics and can
  # be scraped by an external process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timtout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <boolean> | default = false]

  # Optionally defines the the list of enabled-by-default collectors.
  # Anything not provided in the list below will be disabled by default,
  # but requires at least one element to be treated as defined.
  #
  # This is useful if you have a very explicit set of collectors you wish
  # to run.
  set_collectors:
    - [<string>]

  # Additional collectors to enable on top of the default set of enabled
  # collectors or on top of the list provided by set_collectors.
  #
  # This is useful if you have a few collectors you wish to run that are
  # not enabled by default, but do not want to explicitly provide an entire
  # list through set_collectors.
  enable_collectors:
    - [<string>]

  # Additional collectors to disable on top of the default set of disabled
  # collectors. Takes precedence over enable_collectors.

  # Additional collectors to disable from the set of enabled collectors.
  # Takes precedence over enabled_collectors.
  #
  # This is useful if you have a few collectors you do not want to run that
  # are enabled by default, but do not want to explicitly provide an entire
  # list through set_collectors.
  disable_collectors:
    - [<string>]

  # procfs mountpoint.
  [procfs_path: <string> | default = "/proc"]

  # sysfs mountpoint.
  [sysfs_path: <string> | default = "/sys"]

  # rootfs mountpoint. If running in docker, the root filesystem of the host
  # machine should be mounted and this value should be changed to the mount
  # directory.
  [rootfs_path: <string> | default = "/"]

  # Enable the cpu_info metric for the cpu collector.
  [enable_cpu_info_metric: <boolean> | default = true]

  # Regexmp of devices to ignore for diskstats.
  [diskstats_ignored_devices: <string> | default = "^(ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$"]

  # Regexp of mount points to ignore for filesystem collector.
  [filesystem_ignored_mount_points: <string> | default = "^/(dev|proc|sys|var/lib/docker/.+)($|/)"]

  # Regexp of filesystem types to ignore for filesystem collector.
  [filesystem_ignored_fs_types: <string> | default = "^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$"]

  # NTP server to use for ntp collector
  [ntp_server: <string> | default = "127.0.0.1"]

  # NTP protocol version
  [ntp_protocol_version: <int> | default = 4]

  # Certify that the server address is not a public ntp server.
  [ntp_server_is_local: <boolean> | default = false]

  # IP TTL to use wile sending NTP query.
  [ntp_ip_ttl: <int> | default = 1]

  # Max accumulated distance to the root.
  [ntp_max_distance: <duration> | default = "3466080us"]

  # Offset between local clock and local ntpd time to tolerate.
  [ntp_local_offset_tolerance: <duration> | default = "1ms"]

  # Regexp of net devices to ignore for netclass collector.
  [netclass_ignored_devices: <string> | default = "^$"]

  # Regexp of net devices to blacklist (mutually exclusive with whitelist)
  [netdev_device_blacklist: <string> | default = ""]

  # Regexp of net devices to whitelist (mutually exclusive with blacklist)
  [netdev_device_whitelist: <string> | default = ""]

  # Regexp of fields to return for netstat collector.
  [netstat_fields: <string> | default = "^(.*_(InErrors|InErrs)|Ip_Forwarding|Ip(6|Ext)_(InOctets|OutOctets)|Icmp6?_(InMsgs|OutMsgs)|TcpExt_(Listen.*|Syncookies.*|TCPSynRetrans)|Tcp_(ActiveOpens|InSegs|OutSegs|PassiveOpens|RetransSegs|CurrEstab)|Udp6?_(InDatagrams|OutDatagrams|NoPorts|RcvbufErrors|SndbufErrors))$"]

  # List of CPUs from which perf metrics should be collected.
  [perf_cpus: <string> | default = ""]

  # Regexp of power supplies to ignore for the powersupplyclass collector.
  [powersupply_ignored_supplies: <string> | default = "^$"]

  # Path to runit service directory.
  [runit_service_dir: <string> | default = "/etc/service"]

  # XML RPC endpoint for the supervisord collector.
  [supervisord_url: <string> | default = "http://localhost:9001/RPC2"]

  # Regexp of systemd units to whitelist. Units must both match whitelist
  # and not match blacklist to be included.
  [systemd_unit_whitelist: <string> | default = ".+"]

  # Regexp of systemd units to blacklist. Units must both match whitelist
  # and not match blacklist to be included.
  [systemd_unit_blacklist: <string> | default = ".+\\.(automount|device|mount|scope|slice)"]

  # Enables service unit tasks metrics unit_tasks_current and unit_tasks_max
  [systemd_enable_task_metrics: <boolean> | default = false]

  # Enables service unit metric service_restart_total
  [systemd_enable_restarts_metrics: <boolean> | default = false]

  # Enables service unit metric unit_start_time_seconds
  [systemd_enable_start_time_metrics: <boolean> | default = false]

  # Directory to read *.prom files from for the textfile collector.
  [textfile_directory: <string> | default = ""]

  # Regexp of fields to return for the vmstat collector.
  [vmstat_fields: <string> | default = "^(oom_kill|pgpg|pswp|pg.*fault).*"]
```

### process_exporter_config

The `process_exporter_config` block configures the `process_exporter` integration,
which is an embedded version of
[`process-exporter`](https://github.com/ncabatoff/process-exporter)
and allows for collection metrics based on the /proc filesystem on Linux
systems. Note that on non-Linux systems, enabling this exporter is a no-op.

Note that if running the Agent in a container, you will need to bind mount
folders from the host system so the integration can monitor them:

```
docker run \
  -v "/proc:/proc:ro" \
  -v /tmp/agent:/etc/agent \
  -v /path/to/config.yaml:/etc/agent-config/agent.yaml \
  grafana/agent:v0.15.0 \
  --config.file=/etc/agent-config/agent.yaml
```

Replace `/path/to/config.yaml` with the appropriate path on your host system
where an Agent config file can be found.

For running on Kubernetes, ensure to set the equivalent mounts and capabilities
there as well:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: agent
spec:
  containers:
  - image: grafana/agent:v0.16.1
    name: agent
    args:
    - --config.file=/etc/agent-config/agent.yaml
    volumeMounts:
    - name: procfs
      mountPath: /proc
      readOnly: true
  volumes:
  - name: procfs
    hostPath:
      path: /proc
```

The manifest and Tanka configs provided by this repository do not have the
mounts or capabilities required for running this integration.

An example config for `process_exporter_config` that tracks all processes is the
following:

```
enabled: true
process_names:
- name: "{{.Comm}}"
  cmdline:
  - '.+'
```

Full reference of options:

```yaml
  # Enables the process_exporter integration, allowing the Agent to automatically
  # collect system metrics from the host UNIX system.
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the process_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/process_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # procfs mountpoint.
  [procfs_path: <string> | default = "/proc"]

  # If a proc is tracked, track with it any children that aren't a part of their
  # own group.
  [track_children: <boolean> | default = true]

  # Report on per-threadname metrics as well.
  [track_threads: <boolean> | default = true]

  # Gather metrics from smaps file, which contains proportional resident memory
  # size.
  [gather_smaps: <boolean> | default = true]

  # Recheck process names on each scrape.
  [recheck_on_scrape: <boolean> | default = false]

  # A collection of matching rules to use for deciding which processes to
  # monitor. Each config can match multiple processes to be tracked as a single
  # process "group."
  process_names:
    [- <process_matcher_config>]
```

#### process_matcher_config

```yaml
# The name to use for identifying the process group name in the metric. By
# default, it uses the base path of the executable.
#
# The following template variables are available:
#
# - {{.Comm}}:      Basename of the original executable from /proc/<pid>/stat
# - {{.ExeBase}}:   Basename of the executable from argv[0]
# - {{.ExeFull}}:   Fully qualified path of the executable
# - {{.Username}}:  Username of the effective user
# - {{.Matches}}:   Map containing all regex capture groups resulting from
#                   matching a process with the cmdline rule group.
# - {{.PID}}:       PID of the process. Note that the PID is copied from the
#                   first executable found.
# - {{.StartTime}}: The start time of the process. This is useful when combined
#                   with PID as PIDS get reused over time.
[name: <string> | default = "{{.ExeBase}}"]

# A list of strings that match the base executable name for a process, truncated
# at 15 characters. It is derived from reading the second field of
# /proc/<pid>/stat minus the parens.
#
# If any of the strings match, the process will be tracked.
comm:
  [- <string>]

# A list of strings that match argv[0] for a process. If there are no slashes,
# only the basename of argv[0] needs to match. Otherwise the name must be an
# exact match. For example, "postgres" may match any postgres binary but
# "/usr/local/bin/postgres" can only match a postgres at that path exactly.
#
# If any of the strings match, the process will be tracked.
exe:
  [- <string>]

# A list of regular expressions applied to the argv of the process. Each
# regex here must match the corresponding argv for the process to be tracked.
# The first element that is matched is argv[1].
#
# Regex Captures are added to the .Matches map for use in the name.
cmdline:
  [- <string>]
```

### mysqld_exporter_config

The `mysqld_exporter_config` block configures the `mysqld_exporter` integration,
which is an embedded version of
[`mysqld_exporter`](https://github.com/prometheus/mysqld_exporter)
and allows for collection metrics from MySQL servers.

Note that currently, an Agent can only collect metrics from a single MySQL
server. If you want to collect metrics from multiple servers, run multiple
Agents and add labels using `relabel_configs` to differentiate between the MySQL
servers:

```yaml
mysqld_exporter:
  enabled: true
  data_source_name: root@(server-a:3306)/
  relabel_configs:
  - source_labels: [__address__]
    target_label: instance
    replacement: server-a
```

Full reference of options:

```yaml
  # Enables the mysqld_exporter integration, allowing the Agent to collect
  # metrics from a MySQL server.
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the mysqld_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/mysqld_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Data Source Name specifies the MySQL server to connect to. This is REQUIRED
  # but may also be specified by the MYSQLD_EXPORTER_DATA_SOURCE_NAME
  # environment variable. If neither are set, the integration will fail to
  # start.
  #
  # The format of this is specified here: https://github.com/go-sql-driver/mysql#dsn-data-source-name
  #
  # A working example value for a server with no required password
  # authentication is: "root@(localhost:3306)/"
  data_source_name: <string>

  # A list of collector names to enable on top of the default set.
  enable_collectors:
    [ - <string> ]
  # A list of collector names to disable from the default set.
  disable_collectors:
    [ - <string> ]
  # A list of collectors to run. Fully overrides the default set.
  set_collectors:
    [ - <string> ]

  # Set a lock_wait_timeout on the connection to avoid long metadata locking.
  [lock_wait_timeout: <int> | default = 2]
  # Add a low_slow_filter to avoid slow query logging of scrapes. NOT supported
  # by Oracle MySQL.
  [log_slow_filter: <bool> | default = false]

  ## Collector-specific options

  # Minimum time a thread must be in each state to be counted.
  [info_schema_processlist_min_time: <int> | default = 0]
  # Enable collecting the number of processes by user.
  [info_schema_processlist_processes_by_user: <bool> | default = true]
  # Enable collecting the number of processes by host.
  [info_schema_processlist_processes_by_host: <bool> | default = true]
  # The list of databases to collect table stats for. * for all
  [info_schema_tables_databases: <string> | default = "*"]
  # Limit the number of events statements digests by response time.
  [perf_schema_eventsstatements_limit: <int> | default = 250]
  # Limit how old the 'last_seen' events statements can be, in seconds.
  [perf_schema_eventsstatements_time_limit: <int> | default = 86400]
  # Maximum length of the normalized statement text.
  [perf_schema_eventsstatements_digtext_text_limit: <int> | default = 120]
  # Regex file_name filter for performance_schema.file_summary_by_instance
  [perf_schema_file_instances_filter: <string> | default = ".*"]
  # Remove path prefix in performance_schema.file_summary_by_instance
  [perf_schema_file_instances_remove_prefix: <string> | default = "/var/lib/mysql"]
  # Database from where to collect heartbeat data.
  [heartbeat_database: <string> | default = "heartbeat"]
  # Table from where to collect heartbeat data.
  [heartbeat_table: <string> | default = "heartbeat"]
  # Use UTC for timestamps of the current server (`pt-heartbeat` is called with `--utc`)
  [heartbeat_utc: <bool> | default = false]
  # Enable collecting user privileges from mysql.user
  [mysql_user_privileges: <bool> | default = false]
```

The full list of collectors that are supported for `mysqld_exporter` is:

| Name                                             | Description | Enabled by default |
| ------------------------------------------------ | ----------- | ------------------ |
| auto_increment.columns                           | Collect auto_increment columns and max values from information_schema | no |
| binlog_size                                      | Collect the current size of all registered binlog files | no |
| engine_innodb_status                             | Collect from SHOW ENGINE INNODB STATUS | no |
| engine_tokudb_status                             | Collect from SHOW ENGINE TOKUDB STATUS | no |
| global_status                                    | Collect from SHOW GLOBAL STATUS | yes |
| global_variables                                 | Collect from SHOW GLOBAL VARIABLES | yes |
| heartbeat                                        | Collect from heartbeat | no |
| info_schema.clientstats                          | If running with userstat=1, enable to collect client statistics | no |
| info_schema.innodb_cmpmem                        | Collect metrics from information_schema.innodb_cmpmem | yes |
| info_schema.innodb_metrics                       | Collect metrics from information_schema.innodb_metrics | yes |
| info_schema.innodb_tablespaces                   | Collect metrics from information_schema.innodb_sys_tablespaces | no |
| info_schema.processlist                          | Collect current thread state counts from the information_schema.processlist | no |
| info_schema.query_response_time                  | Collect query response time distribution if query_response_time_stats is ON | yes |
| info_schema.replica_host                         | Collect metrics from information_schema.replica_host_status | no |
| info_schema.schemastats                          | If running with userstat=1, enable to collect schema statistics | no |
| info_schema.tables                               | Collect metrics from information_schema.tables | no |
| info_schema.tablestats                           | If running with userstat=1, enable to collect table statistics | no |
| info_schema.userstats                            | If running with userstat=1, enable to collect user statistics | no |
| mysql.user                                       | Collect data from mysql.user | no |
| perf_schema.eventsstatements                     | Collect metrics from performance_schema.events_statements_summary_by_digest | no |
| perf_schema.eventsstatementssum                  | Collect metrics of grand sums from performance_schema.events_statements_summary_by_digest | no |
| perf_schema.eventswaits                          | Collect metrics from performance_schema.events_waits_summary_global_by_event_name | no |
| perf_schema.file_events                          | Collect metrics from performance_schema.file_summary_by_event_name | no |
| perf_schema.file_instances                       | Collect metrics from performance_schema.file_summary_by_instance | no |
| perf_schema.indexiowaits                         | Collect metrics from performance_schema.table_io_waits_summary_by_index_usage | no |
| perf_schema.replication_applier_status_by_worker | Collect metrics from performance_schema.replication_applier_status_by_worker | no |
| perf_schema.replication_group_member_stats       | Collect metrics from performance_schema.replication_group_member_stats | no |
| perf_schema.replication_group_members            | Collect metrics from performance_schema.replication_group_members | no |
| perf_schema.tableiowaits                         | Collect metrics from performance_schema.table_io_waits_summary_by_table | no |
| perf_schema.tablelocks                           | Collect metrics from performance_schema.table_lock_waits_summary_by_table | no |
| slave_hosts                                      | Scrape information from 'SHOW SLAVE HOSTS' | no |
| slave_status                                     | Scrape information from SHOW SLAVE STATUS | yes |


### redis_exporter_config

The `redis_exporter_config` block configures the `redis_exporter` integration, which is an embedded version of [`redis_exporter`](https://github.com/oliver006/redis_exporter). This allows for the collection of metrics from Redis servers.

Note that currently, an Agent can only collect metrics from a single Redis server. If you want to collect metrics from multiple Redis servers, you can run multiple Agents and add labels using `relabel_configs` to differentiate between the Redis servers:

```yaml
redis_exporter:
  enabled: true
  redis_addr: "redis-2:6379"
  relabel_configs:
  - source_labels: [__address__]
    target_label: instance
    replacement: redis-2
```

Full reference of options:
```yaml
  # Enables the redis_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured redis address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the redis_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/redis_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  # exporter-specific configuration options

  # Address of the redis instance.
  redis_addr: <string>

  # User name to use for authentication (Redis ACL for Redis 6.0 and newer).
  [redis_user: <string>]

  # Password of the redis instance.
  [redis_password: <string>]

  # Path of a file containing a passord. If this is defined, it takes precedece
  # over redis_password.
  [redis_password_file: <string>]

  # Namespace for the metrics.
  [namespace: <string> | default = "redis"]

  # What to use for the CONFIG command.
  [config_command: <string> | default = "CONFIG"]

  # Comma separated list of key-patterns to export value and length/size, searched for with SCAN.
  [check_keys: <string>]

  # Comma separated list of LUA regex for grouping keys. When unset, no key
  # groups will be made.
  [check_key_groups: <string>]

  # Check key groups batch size hint for the underlying SCAN.
  [check_key_groups_batch_size: <int> | default = 10000]

  # The maximum number of distinct key groups with the most memory utilization
  # to present as distinct metrics per database. The leftover key groups will be
  # aggregated in the 'overflow' bucket.
  [max_distinct_key_groups: <int> | default = 100]

  # Comma separated list of single keys to export value and length/size.
  [check_single_keys: <string>]

  # Comma separated list of stream-patterns to export info about streams, groups and consumers, searched for with SCAN.
  [check_streams: <string>]

  # Comma separated list of single streams to export info about streams, groups and consumers.
  [check_single_streams: <string>]

  # Comma separated list of individual keys to export counts for.
  [count_keys: <string>]

  # Path to Lua Redis script for collecting extra metrics.
  [script_path: <string>]

  # Timeout for connection to Redis instance (in Golang duration format).
  [connection_timeout: <time.Duration> | default = "15s"]

  # Name of the client key file (including full path) if the server requires TLS client authentication.
  [tls_client_key_file: <string>]

  # Name of the client certificate file (including full path) if the server requires TLS client authentication.
  [tls_client_cert_file: <string>]

  # Name of the CA certificate file (including full path) if the server requires TLS client authentication.
  [tls_ca_cert_file: <string>]

  # Whether to set client name to redis_exporter.
  [set_client_name: <bool>]

  # Whether to scrape Tile38 specific metrics.
  [is_tile38: <bool>]

  # Whether to scrape Client List specific metrics.
  [export_client_list: <bool>]

  # Whether to include the client's port when exporting the client list. Note
  # that including this will increase the cardinality of all redis metrics.
  [export_client_port: <bool>]

  # Whether to also export go runtime metrics.
  [redis_metrics_only: <bool>]

  # Whether to ping the redis instance after connecting.
  [ping_on_connect: <bool>]

  # Whether to include system metrics like e.g. redis_total_system_memory_bytes.
  [incl_system_metrics: <bool>]

  # Whether to to skip TLS verification.
  [skip_tls_verification: <bool>]
```

### dnsmasq_exporter_config

The `dnsmasq_exporter_config` block configures the `dnsmasq_exporter` integration,
which is an embedded version of
[`dnsmasq_exporter`](https://github.com/google/dnsmasq_exporter). This allows for
the collection of metrics from dnsmasq servers.

Note that currently, an Agent can only collect metrics from a single dnsmasq
server. If you want to collect metrics from multiple servers, you can run
multiple Agents and add labels using `relabel_configs` to differentiate between
the servers:

```yaml
dnsmasq_exporter:
  enabled: true
  dnsmasq_address: dnsmasq-a:53
  relabel_configs:
  - source_labels: [__address__]
    target_label: instance
    replacement: dnsmasq-a
```

Full reference of options:

```yaml
  # Enables the dnsmasq_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured dnsmasq server address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the dnsmasq_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/dnsmasq_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  #
  # Exporter-specific configuration options
  #

  # Address of the dnsmasq server in host:port form.
  [dnsmasq_address: <string> | default = "localhost:53"]

  # Path to the dnsmasq leases file. If this file doesn't exist, scraping
  # dnsmasq # will fail with an warning log message.
  [leases_path: <string> | default = "/var/lib/misc/dnsmasq.leases"]
```

### elasticsearch_exporter_config

The `elasticsearch_exporter_config` block configures the `elasticsearch_exporter` integration,
which is an embedded version of
[`elasticsearch_exporter`](https://github.com/justwatchcom/elasticsearch_exporter). This allows for
the collection of metrics from ElasticSearch servers.

Note that currently, an Agent can only collect metrics from a single ElasticSearch server.
However, the exporter is able to collect the metrics from all nodes through that server configured.

Full reference of options:

```yaml
  # Enables the elasticsearch_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured ElasticSearch server address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the elasticsearch_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/elasticsearch_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  #
  # Exporter-specific configuration options
  #

  # HTTP API address of an Elasticsearch node.
  [ address : <string> | default = "http://localhost:9200" ]

  # Timeout for trying to get stats from Elasticsearch.
  [ timeout: <duration> | default = "5s" ]

  # Export stats for all nodes in the cluster. If used, this flag will override the flag `node`.
  [ all: <boolean> ]

  # Node's name of which metrics should be exposed.
  [ node: <boolean> ]

  # Export stats for indices in the cluster.
  [ indices: <boolean> ]

  # Export stats for settings of all indices of the cluster.
  [ indices_settings: <boolean> ]

  # Export stats for cluster settings.
  [ cluster_settings: <boolean> ]

  # Export stats for shards in the cluster (implies indices).
  [ shards: <boolean> ]

  # Export stats for the cluster snapshots.
  [ snapshots: <boolean> ]

  # Cluster info update interval for the cluster label.
  [ clusterinfo_interval: <duration> | default = "5m" ]

  # Path to PEM file that contains trusted Certificate Authorities for the Elasticsearch connection.
  [ ca: <string> ]

  # Path to PEM file that contains the private key for client auth when connecting to Elasticsearch.
  [ client_private_key: <string> ]

  # Path to PEM file that contains the corresponding cert for the private key to connect to Elasticsearch.
  [ client_cert: <string> ]

  # Skip SSL verification when connecting to Elasticsearch.
  [ ssl_skip_verify: <boolean> ]
```

### memcached_exporter_config

The `memcached_exporter_config` block configures the `memcached_exporter`
integration, which is an embedded version of
[`memcached_exporter`](https://github.com/prometheus/memcached_exporter). This
allows for the collection of metrics from memcached servers.

Note that currently, an Agent can only collect metrics from a single memcached
server. If you want to collect metrics from multiple servers, you can run
multiple Agents and add labels using `relabel_configs` to differentiate between
the servers:

```yaml
memcached_exporter:
  enabled: true
  memcached_address: memcached-a:53
  relabel_configs:
  - source_labels: [__address__]
    target_label: instance
    replacement: memcached-a
```

Full reference of options:

```yaml
  # Enables the memcached_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured memcached server address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the memcached_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/memcached_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  #
  # Exporter-specific configuration options
  #

  # Address of the memcached server in host:port form.
  [memcached_address: <string> | default = "localhost:53"]

  # Timeout for connecting to memcached.
  [timeout: <duration> | default = "1s"]
```

### postgres_exporter_config

The `postgres_exporter_config` block configures the `postgres_exporter`
integration, which is an embedded version of
[`postgres_exporter`](https://github.com/wrouesnel/postgres_exporter). This
allows for the collection of metrics from Postgres servers.

Full reference of options:

```yaml
  # Enables the postgres_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured postgres server address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the postgres_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/postgres_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  #
  # Exporter-specific configuration options
  #

  # Data Source Names specifies the Postgres server(s) to connect to. This is
  # REQUIRED but may also be specified by the POSTGRES_EXPORTER_DATA_SOURCE_NAME
  # environment variable, where DSNs the environment variable are separated by
  # commas. If neither are set, the integration will fail to start.
  #
  # The format of this is specified here: https://pkg.go.dev/github.com/lib/pq#ParseURL
  #
  # A working example value for a server with a password is:
  # "postgresql://username:passwword@localhost:5432/database?sslmode=disable"
  #
  # Multiple DSNs may be provided here, allowing for scraping from multiple
  # servers.
  data_source_names:
  - <string>

  # Disables collection of metrics from pg_settings.
  [disable_settings_metrics: <boolean> | default = false]

  # Autodiscover databases to collect metrics from. If false, only collects
  # metrics from databases collected from data_source_names.
  [autodiscover_databases: <boolean> | default = false]

  # Excludes specific databases from being collected when autodiscover_databases
  # is true.
  exclude_databases:
  [ - <string> ]

  # Path to a YAML file containing custom queries to run. Check out
  # postgres_exporter's queries.yaml for examples of the format:
  # https://github.com/prometheus-community/postgres_exporter/blob/master/queries.yaml
  [query_path: <string> | default = ""]

  # When true, only exposes metrics supplied from query_path.
  [disable_default_metrics: <boolean> | default = false]
```

### statsd_exporter_config

The `statsd_exporter_config` block configures the `statsd_exporter`
integration, which is an embedded version of
[`statsd_exporter`](https://github.com/prometheus/statsd_exporter). This allows
for the collection of statsd metrics and exposing them as Prometheus metrics.

Full reference of options:

```yaml
  # Enables the statsd_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured statsd server address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the statsd_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/statsd_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  #
  # Exporter-specific configuration options
  #

  # The UDP address on which to receive statsd metric lines. An empty string
  # will disable UDP collection.
  [listen_udp: <string> | default = ":9125"]

  # The TCP address on which to receive statsd metric lines. An empty string
  # will disable TCP collection.
  [listen_tcp: <string> | default = ":9125"]

  # The Unixgram socket path to receive statsd metric lines. An empty string
  # will disable unixgram collection.
  [listen_unixgram: <string> | default = ""]

  # The permission mode of the unixgram socket, when enabled.
  [unix_socket_mode: <string> | default = "755"]

  # An optional mapping config that can translate dot-separated StatsD metrics
  # into labeled Prometheus metrics. For full instructions on how to write this
  # object, see the official documentation from the statsd_exporter:
  #
  # https://github.com/prometheus/statsd_exporter#metric-mapping-and-configuration
  #
  # Note that a SIGHUP will not reload this config.
  [mapping_config: <statsd_exporter.mapping_config>]

  # Size (in bytes) of the operating system's transmit read buffer associated
  # with the UDP or unixgram connection. Please make sure the kernel parameters
  # net.core.rmem_max is set to a value greater than the value specified.
  [read_buffer: <int> | default = 0]

  # Maximum size of your metric mapping cache. Relies on least recently used
  # replacement policy if max size is reached.
  [cache_size: <int> | default = 1000]

  # Metric mapping cache type. Valid values are "lru" and "random".
  [cache_type: <string> | default = "lru"]

  # Size of internal queue for processing events.
  [event_queue_size: <int> | default = 10000]

  # Number of events to hold in queue before flushing.
  [event_flush_threshold: <int> | default = 1000]

  # Number of events to hold in queue before flushing.
  [event_flush_interval: <duration> | default = "200ms"]

  # Parse DogStatsd style tags.
  [parse_dogstatsd_tags: <bool> | default = true]

  # Parse InfluxDB style tags.
  [parse_influxdb_tags: <bool> | default = true]

  # Parse Librato style tags.
  [parse_librato_tags: <bool> | default = true]

  # Parse SignalFX style tags.
  [parse_signalfx_tags: <bool> | default = true]
```

### consul_exporter_config

The `consul_exporter_config` block configures the `consul_exporter`
integration, which is an embedded version of
[`consul_exporter`](https://github.com/prometheus/consul_exporter). This allows
for the collection of consul metrics and exposing them as Prometheus metrics.

Full reference of options:

```yaml
  # Enables the consul_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured consul server address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the consul_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/consul_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  #
  # Exporter-specific configuration options
  #

  # Prefix from which to expose key/value pairs.
  [kv_prefix: <string> | default = ""]

  # Regex that determines which keys to expose.
  [kv_filter: <string> | default = ".*"]

  # Generate a health summary for each service instance. Needs n+1 queries to
  # collect all information.
  [generate_health_summary: <bool> | default = true]

  # HTTP API address of a Consul server or agent. Prefix with https:// to
  # connect using HTTPS.
  [server: <string> | default = "http://localhost:8500"]

  # Disable TLS host verification.
  [insecure_skip_verify: <bool> | default = false]

  # File path to a PEM-encoded certificate authority used to validate the
  # authenticity of a server certificate.
  [ca_file: <string> | default = ""]

  # File path to a PEM-encoded certificate used with the private key to verify
  # the exporter's authenticity.
  [cert_file: <string> | default = ""]

  # File path to a PEM-encoded private key used with the certificate to verify
  # the exporter's authenticity.
  [key_file: <string> | default = ""]

  # When provided, this overrides the hostname for the TLS certificate. It can
  # be used to ensure that the certificate name matches the hostname we declare.
  [server_name: <string> | default = ""]

  # Timeout on HTTP requests to the Consul API.
  [timeout: <duration> | default = "500ms"]

  # Limit the maximum number of concurrent requests to consul. 0 means no limit.
  [concurrent_request_limit: <int> | default = 0]

  # Allows any Consul server (non-leader) to service a read.
  [allow_stale: <bool> | default = true]

  # Forces the read to be fully consistent.
  [require_consistent: <bool> | default = false]
```


### windows_exporter_config

The `windows_exporter_config` block configures the `windows_exporter`
integration, which is an embedded version of
[`windows_exporter`](https://github.com/grafana/windows_exporter). This allows
for the collection of Windows metrics and exposing them as Prometheus metrics.

Full reference of options:

```yaml
  # Enables the windows_exporter integration, allowing the Agent to automatically
  # collect system metrics from the local windows instance
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the consul_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/windows_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  #
  # Exporter-specific configuration options
  #

  # List of collectors to enable
  [enabled_collectors: <string> | default = "cpu,cs,logical_disk,net,os,service,system,textfile"]

  # The following settings are only used if they are enabled by specifying them in enabled_collectors

  # Configuration for Exchange Mail Server
  exchange:
    # Comma-separated List of collectors to use. Defaults to all, if not specified.
    # Maps to collectors.exchange.enabled in windows_exporter
    [enabled_list: <string>]

  # Configuration for the IIS web server
  iis:
    # Regexp of sites to whitelist. Site name must both match whitelist and not match blacklist to be included.
    # Maps to collector.iis.site-whitelist in windows_exporter
    [site_whitelist: <string> | default = ".+"]

    # Regexp of sites to blacklist. Site name must both match whitelist and not match blacklist to be included.
    # Maps to collector.iis.site-blacklist in windows_exporter
    [site_blacklist: <string> | default = ""]

    # Regexp of apps to whitelist. App name must both match whitelist and not match blacklist to be included.
    # Maps to collector.iis.app-whitelist in windows_exporter
    [app_whitelist: <string> | default=".+"]

    # Regexp of apps to blacklist. App name must both match whitelist and not match blacklist to be included.
    # Maps to collector.iis.app-blacklist in windows_exporter
    [app_blacklist: <string> | default=".+"]

  # Configuration for reading metrics from a text files in a directory
  text_file:
    # Directory to read text files with metrics from.
    # Maps to collector.textfile.directory in windows_exporter
    [text_file_directory: <string> | default="C:\Program Files\windows_exporter\textfile_inputs"]

  # Configuration for SMTP metrics
  smtp:
    # Regexp of virtual servers to whitelist. Server name must both match whitelist and not match blacklist to be included.
    # Maps to collector.smtp.server-whitelist in windows_exporter
    [whitelist: <string> | default=".+"]

    # Regexp of virtual servers to blacklist. Server name must both match whitelist and not match blacklist to be included.
    # Maps to collector.smtp.server-blacklist in windows_exporter
    [blacklist: <string> | default=""]

  # Configuration for Windows Services
  service:
    # "WQL 'where' clause to use in WMI metrics query. Limits the response to the services you specify and reduces the size of the response.
    # Maps to collector.service.services-where in windows_exporter
    [where_clause: <string> | default=""]

  # Configuration for Windows Processes
  process:
    # Regexp of processes to include. Process name must both match whitelist and not match blacklist to be included.
    # Maps to collector.process.whitelist in windows_exporter
    [whitelist: <string> | default=".+"]

    # Regexp of processes to exclude. Process name must both match whitelist and not match blacklist to be included.
    # Maps to collector.process.blacklist in windows_exporter
    [blacklist: <string> | default=""]

  # Configuration for NICs
  network:
    # Regexp of NIC's to whitelist. NIC name must both match whitelist and not match blacklist to be included.
    # Maps to collector.net.nic-whitelist in windows_exporter
    [whitelist: <string> | default=".+"]

    # Regexp of NIC's to blacklist. NIC name must both match whitelist and not match blacklist to be included.
    # Maps to collector.net.nic-blacklist in windows_exporter
    [blacklist: <string> | default=""]

  # Configuration for Microsoft SQL Server
  mssql:
    # Comma-separated list of mssql WMI classes to use.
    # Maps to collectors.mssql.classes-enabled in windows_exporter
    [enabled_classes: <string> | default="accessmethods,availreplica,bufman,databases,dbreplica,genstats,locks,memmgr,sqlstats,sqlerrors,transactions"]

  # Configuration for Microsoft Queue
  msqm:
    # WQL 'where' clause to use in WMI metrics query. Limits the response to the msmqs you specify and reduces the size of the response.
    # Maps to collector.msmq.msmq-where in windows_exporter
    [where_clause: <string> | default=""]

  # Configuration for disk information
  logical_disk:
    # Regexp of volumes to whitelist. Volume name must both match whitelist and not match blacklist to be included.
    # Maps to collector.logical_disk.volume-whitelist in windows_exporter
    [whitelist: <string> | default=".+"]

    # Regexp of volumes to blacklist. Volume name must both match whitelist and not match blacklist to be included.
    # Maps to collector.logical_disk.volume-blacklist in windows_exporter
    [blacklist: <string> | default=".+"]
```
