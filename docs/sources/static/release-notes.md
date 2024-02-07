---
aliases:
- ../upgrade-guide/
- ./upgrade-guide/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/release-notes/
- /docs/grafana-cloud/send-data/agent/static/release-notes/
canonical: https://grafana.com/docs/agent/latest/static/release-notes/
description: Release notes for Grafana Agent static mode
menuTitle: Release notes
title: Release notes
weight: 999
---

# Release notes

The release notes provide information about deprecations and breaking changes in Grafana Agent static mode.

For a complete list of changes to Grafana Agent, with links to pull requests and related issues when available, refer to the [Changelog](https://github.com/grafana/agent/blob/main/CHANGELOG.md).

> **Note:** These release notes are specific to Grafana Agent static mode.
> Other release notes for the different Grafana Agent variants are contained on separate pages:
>
> * [Static mode Kubernetes operator release notes][release-notes-operator]
> * [Flow mode release notes][release-notes-flow]

{{% docs/reference %}}
[release-notes-operator]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/operator/release-notes"
[release-notes-operator]: "/docs/grafana-cloud/ -> ../operator/release-notes"

[release-notes-flow]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/release-notes"
[release-notes-flow]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow/release-notes"

[Modules]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/concepts/modules"
[Modules]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow/concepts/modules"
{{% /docs/reference %}}

## v0.38

### Breaking change: support for exporting Jaeger traces removed

The deprecated support for exporting Jaeger-formatted traces has been removed.
To send traces to Jaeger, export OTLP-formatted data to a version of Jaeger
that supports OTLP.

## v0.37

### Breaking change: The default value of `retry_on_http_429` is overriden to `true` for the `queue_config` in `remote_write` in `metrics` config.

{{< admonition type="note" >}}
The default set by Grafana Agent Static Mode is different than the default set by Prometheus.
{{< /admonition >}}

The Prometheus default value for `retry_on_http_429` is set to `true` for the `queue_config` in `remote_write`.
This changed default setting allows the agent to retry sending data when it receives an HTTP 429 error and helps avoid losing data in metric pipelines.

* If you explicitly set the `retry_on_http_429`, no action is required.
* If you do not explicitly set `retry_on_http_429` and you do *not* want to retry on HTTP 429, make sure you set it to `false` when you upgrade to this new version.

### Breaking change: Renamed `non_indexed_labels` Loki processing stage to `structured_metadata`.

If you use the Loki processing stage in your Agent configuration, you must rename the `non_indexed_labels` pipeline stage definition to `structured_metadata`.


Old configuration example:

```yaml
  pipeline_stages:
    - logfmt:
        mapping:
          app:
    - non_indexed_labels:
        app:
```

New configuration example:

```yaml
  pipeline_stages:
    - logfmt:
        mapping:
          app:
    - structured_metadata:
        app:
```

## v0.35

### Breaking change: Jaeger remote sampling no longer configurable using the Jaeger receiver

Jaeger remote sampling used to be configured using the Jaeger receiver configuration. This receiver was updated to a new version, where support for remote sampling in the receiver was removed.

Jaeger remote sampling is available as a separate configuration field starting in v0.35.3.

Old configuration example:

```yaml
receivers:
  jaeger:
    protocols:
      grpc:
    remote_sampling:
      strategy_file: /etc/agent/strategies.json
      strategy_file_reload_interval: 1s
```

New configuration example:

```yaml
jaeger_remote_sampling:
  - source:
      file: /etc/agent/strategies.json
      reload_interval: 1s
```

### Breaking change: `auth` and `version` attributes from `walk_params` block of SNMP integration have been removed

The SNMP integrations (both v1 and v2) wrap a new version of SNMP exporter which introduces a new configuration file format.
This new format separates the walk and metric mappings from the connection and authentication settings. This allows for easier configuration of different
auth params without having to duplicate the full walk and metric mapping.

Old configuration example:

```yaml
    snmp_targets:
      - name: network_switch_1
        address: 192.168.1.2
        module: if_mib
        walk_params: public
        auth: public
    walk_params:
      public:
        retries: 2
        version: 2
        auth:
          community: public
```

New configuration example:

```yaml
    snmp_targets:
      - name: network_switch_1
        address: 192.168.1.2
        module: if_mib
        walk_params: public
        auth: public
    walk_params:
      public:
        retries: 2
```

See [Module and Auth Split Migration](https://github.com/prometheus/snmp_exporter/blob/main/auth-split-migration.md) for more details.

### Removal of Dynamic Configuration

The experimental feature Dynamic Configuration has been removed. The use case of dynamic configuration will be replaced
with [Modules][] in Grafana Agent Flow.

### Breaking change: Removed and renamed tracing metrics

In the traces subsystem for Static mode some metrics are removed and others are renamed.
The reason for the removal is a bug which caused the metrics to be incorrect if more than one instance of a traces configuration is specified.

Removed metrics:
- "blackbox_exporter_config_last_reload_success_timestamp_seconds" (gauge)
- "blackbox_exporter_config_last_reload_successful" (gauge)
- "blackbox_module_unknown_total" (counter)
- "traces_processor_tail_sampling_count_traces_sampled" (counter)
- "traces_processor_tail_sampling_new_trace_id_received" (counter)
- "traces_processor_tail_sampling_sampling_decision_latency" (histogram)
- "traces_processor_tail_sampling_sampling_decision_timer_latency" (histogram)
- "traces_processor_tail_sampling_sampling_policy_evaluation_error" (counter)
- "traces_processor_tail_sampling_sampling_trace_dropped_too_early" (counter)
- "traces_processor_tail_sampling_sampling_traces_on_memory" (gauge)
- "traces_receiver_accepted_spans" (counter)
- "traces_receiver_refused_spans" (counter)
- "traces_exporter_enqueue_failed_log_records" (counter)
- "traces_exporter_enqueue_failed_metric_points" (counter)
- "traces_exporter_enqueue_failed_spans" (counter)
- "traces_exporter_queue_capacity" (gauge)
- "traces_exporter_queue_size" (gauge)

Renamed metrics:
- "traces_receiver_refused_spans" is renamed to "traces_receiver_refused_spans_total"
- "traces_receiver_accepted_spans" is renamed to "traces_receiver_refused_spans_total"
- "traces_exporter_sent_metric_points" is renamed to "traces_exporter_sent_metric_points_total"

## v0.33

### Symbolic links in Docker containers removed

We've removed the deprecated symbolic links to `/bin/agent*` in Docker
containers, as planned in v0.31. In case you're setting a custom entrypoint,
use the new binaries that are prefixed with `/bin/grafana*`.

### Deprecation of Dynamic Configuration

[Dynamic Configuration](/docs/agent/v0.33/cookbook/dynamic-configuration/) will be removed in v0.34.
The use case of dynamic configuration will be replaced with Modules in Grafana Agent Flow.

## v0.32

### Breaking change: `node_exporter` configuration options changed

With the update of the `node_exporter` integration to use v1.5.0, configuration
options for the `diskstats` collector have changed names:

- `diskstats_ignored_devices` is now `diskstats_device_exclude` in the static
  mode configuration.
- `ignored_devices` is now `device_exclude` in the Flow component
  configuration.

## v0.31.1

### Breaking change: all Windows executables are now zipped

All release Windows `.exe` files are now zipped. Prior to v0.31, only
`grafana-agent-installer.exe` was unzipped.

This fixes an issue from v0.31.0 where all `.exe` files were accidentally left
unzipped.

## v0.31

### Breaking change: binary names are now prefixed with `grafana-`

As first announced in v0.29, release binary names are now prefixed with
`grafana-`:

- `agent` is now `grafana-agent`.
- `agentctl` is now `grafana-agentctl`.

For the `grafana/agent` Docker container, the entrypoint is now
`/bin/grafana-agent`. A symbolic link from `/bin/agent` to the new binary has
been added.

For the `grafana/agentctl` Docker container, the entrypoint is now
`/bin/grafana-agentctl`. A symbolic link from `/bin/agentctl` to the new binary
has been added.

These symbolic links will be removed in v0.33. Custom entrypoints must be
updated prior to v0.33 to use the new binaries before the symbolic links get
removed.

## v0.30

### Breaking change: `ebpf_exporter` integration removed

The `ebpf_exporter` version bundled in the Agent used [bcc][] to compile eBPF
programs at runtime. This made it hard to run successfully, as the
dynamic linking approach required a compiler, the correct kernel headers, as
well as an exact match of the libbpf toolchain on the host system.  For these
reasons, we've decided to remove the `ebpf_exporter` integration.

Running the `ebpf_exporter` integration is now deprecated and will result in
configuration errors. To continue using the same configuration file, remove the
`ebpf` block.

[bcc]: https://github.com/iovisor/bcc

## v0.29

### Breaking change: JSON-encoded traces from OTLP versions below 0.16.0 are no longer supported

Grafana Agent's OpenTelemetry Collector dependency has been updated from
v0.55.0 to v0.61.0. OpenTelemetry Collector v0.58.0 [no longer
translates][translation-removal] from InstrumentationLibrary to Scope.

This means that JSON-encoded traces that still use InstrumentationLibrary will
be dropped. To work around this issue, either send traces using protobuf or
update your OTLP protocol version to v0.16.0 or newer.

[translation-removal]: https://github.com/open-telemetry/opentelemetry-collector/pull/5819

### Deprecation: binary names will be prefixed with `grafana-` in v0.31.0

The binary names `agent` and `agentctl` have been deprecated
and will be renamed to `grafana-agent` and `grafana-agentctl` respectively in
the v0.31.0 release.

As part of this change, the Docker containers for the v0.31.0 release will
include symbolic links from the old binary names to the new binary names.

There is no action to take at this time.

## v0.26

### Breaking change: Deprecated YAML fields in `server` block removed

The YAML fields which were first [deprecated in the v0.24.0
release](#deprecation-on-yaml-fields-in-server-block-that-have-flags) have now
been removed, replaced by equivalent command line flags. Please refer to the
original deprecation notice for instructions for how to migrate to the command
line flags.

### Breaking change: Reconcile sampling policies between Agent and OTel

Configuring sampling policies in the `tail_sampling` block of the `traces`
block has been changed to be equal with the upstream configuration of the OTel
processor. It now requires that the policy `type` is specified.

Old configuration:

```yaml
traces:
  configs:
    - name: default
    ...
    tail_sampling:
      policies:
      - latency:
          threshold_ms: 100
```

New configuration:

```yaml
traces:
  configs:
    - name: default
    ...
    tail_sampling:
      policies:
      - type: latency
        latency:
          threshold_ms: 100
```

## v0.24

### Breaking change: Integrations renamed when `integrations-next` feature flag is used

This change only applies to users utilizing the `integrations-next` feature
flag. Nothing is changed for configuring integrations when the feature flag is
not used.

Most `integrations-next` integrations have been renamed to describe what
telemetry data they generate instead of the projects they are powered by.

* `consul_exporter` is now `consul`
* `dnsmasq_exporter` is now `dnsmasq`
* `elasticsearch_exporter` is now `elasticsearch`
* `github_exporter` is now `github`
* `kafka_exporter` is now `kafka`
* `memcached_exporter` is now `memcached`
* `mongodb_exporter` is now `mongodb`
* `mysqld_exporter` is now `mysql`
  * Note that it is `mysql` and _not_ `mysqld`
* `postgres_exporter` is now `postgres`
* `process_exporter` is now `process`
* `redis_exporter` is now `redis`
* `statsd_exporter` is now `statsd`
* `windows_exporter` is now `windows`

Keys in the `integrations` config block have changed to match the above:

* `integrations.consul_exporter_configs` is now `integrations.consul_configs`
* `integrations.dnsmasq_exporter_configs` is now `integrations.dnsmasq_configs`
* `integrations.elasticsearch_exporter_configs` is now `integrations.elasticsearch_configs`
* `integrations.github_exporter_configs` is now `integrations.github_configs`
* `integrations.kafka_exporter_configs` is now `integrations.kafka_configs`
* `integrations.memcached_exporter_configs` is now `integrations.memcached_configs`
* `integrations.mongodb_exporter_configs` is now `integrations.mongodb_configs`
* `integrations.mysqld_exporter_configs` is now `integrations.mysql_configs`
* `integrations.postgres_exporter_configs` is now `integrations.postgres_configs`
* `integrations.process_exporter` is now `integrations.process`
* `integrations.redis_exporter_configs` is now `integrations.redis_configs`
* `integrations.statsd_exporter` is now `integrations.statsd`
* `integrations.windows_exporter` is now `integrations.windows`

Integrations not listed here have not changed; `node_exporter` still has the
same name.

This change propagates to the label values generated by these integrations. For
example, `job="integrations/redis_exporter` will now be `job="redis"`.

### Change: Separating YAML and command line flags

As of this release, we are starting to separate what can be configured within
the YAML file, and what can be configured by command line flag. Previously,
there was a lot of overlap: many things could be set by both command line flag
and configuration file, with command line flags taking precedence.

The configuration file will be used for settings that can be updated at runtime
using the `/-/reload` endpoint or sending SIGHUP. Meanwhile, command line flags
will be used for settings that must remain consistent throughout the process
lifetime, such as the HTTP listen port.

This conceptual change will require some number of breaking changes. This
release focuses on the `server` block of the YAML, which has historically
caused the most issues with the `/-/reload` endpoint working correctly.

There may be more breaking changes in the future as we identify more settings
that must be static and moved to flags. These changes will either be moving a
YAML field to a flag or moving a flag to a YAML field. After we are done with
this migration, there will be no overlap between flags and the YAML file.

### Deprecation on YAML fields in `server` block that have flags

The `server` block is the most impacted by the separation of flags/fields.
Instead of making a breaking change immediately, we are deprecating these
fields.

> **NOTE**: These deprecated fields will be removed in the v0.26.0 release. We
> will communicate when other deprecated features will be removed when a
> timeline is established.

The following fields are now deprecated in favor of command line flags:

* `server.register_instrumentation`
* `server.graceful_shutdown_timeout`
* `server.log_source_ips_enabled`
* `server.log_source_ips_header`
* `server.log_source_ips_regex`
* `server.http_listen_network`
* `server.http_listen_address`
* `server.http_listen_port`
* `server.http_listen_conn_limit`
* `server.http_server_read_timeout`
* `server.http_server_write_timout`
* `server.http_server_idle_timeout`
* `server.grpc_listen_network`
* `server.grpc_listen_address`
* `server.grpc_listen_port`
* `server.grpc_listen_conn_limit`
* `server.grpc_server_max_recv_msg_size`
* `server.grpc_server_max_send_msg_size`
* `server.grpc_server_max_concurrent_streams`
* `server.grpc_server_max_connection_idle`
* `server.grpc_server_max_connection_age`
* `server.grpc_server_max_connection_age_grace`
* `server.grpc_server_keepalive_time`
* `server.grpc_server_keepalive_timeout`
* `server.grpc_server_min_time_between_pings`
* `server.grpc_server_ping_without_stream_allowed`

This is most of the fields; the remaining non-deprecated fields are
`server.log_level`, `server.log_format`, `server.http_tls_config`, and
`server.grpc_tls_config`, which support dynamic updating.

### Breaking change: Removing support for dynamically updating deprecated server fields

`/-/reload` will now fail if any of the deprecated server block fields have
changed. It is still valid to change a non-deprecated field (i.e., changing the
log level).

### Breaking change: Server-specific command line flags have changed

The following flags are _new_:

* `-server.http.enable-tls`
* `-server.grpc.enable-tls`
* `-server.http.address`
* `-server.grpc.address`

The following flags have been _removed_:

* `-log.level` (replacement: use YAML field `server.log_level`)
* `-log.format` (replacement: use YAML field `server.log_format`)
* `-server.http-tls-cert-path` (replacement: use YAML field `server.http_tls_config`)
* `-server.http-tls-key-path` (replacement: use YAML field `server.http_tls_config`)
* `-server.http-tls-client-auth` (replacement: use YAML field `server.http_tls_config`)
* `-server.http-tls-ca-path` (replacement: use YAML field `server.http_tls_config`)
* `-server.grpc-tls-cert-path` (replacement: use YAML field `server.grpc_tls_config`)
* `-server.grpc-tls-key-path` (replacement: use YAML field `server.grpc_tls_config`)
* `-server.grpc-tls-client-auth` (replacement: use YAML field `server.grpc_tls_config`)
* `-server.grpc-tls-ca-path` (replacement: use YAML field `server.grpc_tls_config`)
* `-server.http-listen-address` (replacement: use the new `-server.http.address` flag, which combines host and port)
* `-server.http-listen-port` (replacement: use the new  `-server.http.address` flag, which combines host and port)
* `-server.grpc-listen-address` (replacement: use the new `-server.grpc.address` flag, which combines host and port)
* `-server.grpc-listen-port` (replacement: use the new `-server.grpc.address` flag, which combines host and port)
* `-server.path-prefix` (no replacement; this flag was unsupported and caused undefined behavior when set)

The following flags have been _renamed_:

* `-server.log-source-ips-enabled` has been renamed to `-server.log.source-ips.enabled`
* `-server.log-source-ips-header` has been renamed to `-server.log.source-ips.header`
* `-server.log-source-ips-regex` has been renamed to `-server.log.source-ips.regex`
* `-server.http-listen-network` has been renamed to `-server.http.network`
* `-server.http-conn-limit` has been renamed to `-server.http.conn-limit`
* `-server.http-read-timeout` has been renamed to `-server.http.read-timeout`
* `-server.http-write-timeout` has been renamed to `-server.http.write-timeout`
* `-server.http-idle-timeout` has been renamed to `-server.http.idle-timeout`
* `-server.grpc-listen-network` has been renamed to `-server.grpc.network`
* `-server.grpc-conn-limit` has been renamed to `-server.grpc.conn-limit`
* `-server.grpc-max-recv-msg-size-bytes` has been renamed to `-server.grpc.max-recv-msg-size-bytes`
* `-server.grpc-max-send-msg-size-bytes` has been renamed to `-server.grpc.max-send-msg-size-bytes`
* `-server.grpc-max-concurrent-streams` has been renamed to `-server.grpc.max-concurrent-streams`

### Breaking change: New TLS flags required for enabling TLS

The two new flags, `-server.http.enable-tls` and `-server.grpc.enable-tls` now
must be provided for TLS support to be enabled.

This is a change over the previous behavior where TLS was automatically enabled
when a certificate pair was provided.

### Breaking change: Default HTTP/gRPC address changes

The HTTP and gRPC listen addresses now default to `127.0.0.1:12345` and
`127.0.0.1:12346` respectively.

If running inside of a container, you must change these to `0.0.0.0` to
externally communicate with the agent's HTTP server.

The listen addresses may be changed via `-server.http.address` and
`-server.grpc.address` respectively.

### Breaking change: Removal of `-reload-addr` and `-reload-port` flags

The `-reload-addr` and `-reload-port` flags have been removed. They were
initially added to workaround an issue where reloading a changed server block
would cause the primary HTTP server to restart. As the HTTP server settings are
now static, this can no longer happen, and as such the flags have been removed.

### Change: In-memory autoscrape for integrations-next

This change is only relevant to those using the `integrations-next` feature flag.

In-memory connections will now be used for autoscraping-enabled integrations.
This is a change over the previous behavior where autoscraping integrations
would connect to themselves over the network. As a result of this change, the
`integrations.client_config` field is no longer necessary and has been removed.

## v0.22

### `node_exporter` integration deprecated field names

The following field names for the `node_exporter` integration are now deprecated:

* `netdev_device_whitelist` is deprecated in favor of `netdev_device_include`.
* `netdev_device_blacklist` is deprecated in favor of `netdev_device_exclude`.
* `systemd_unit_whitelist` is deprecated in favor of `systemd_unit_include`.
* `systemd_unit_blacklist` is deprecated in favor of `systemd_unit_exclude`.
* `filesystem_ignored_mount_points` is deprecated in favor of
  `filesystem_mount_points_exclude`.
* `filesystem_ignored_fs_types` is deprecated in favor of
  `filesystem_fs_types_exclude`.

This change aligns with the equivalent flag names also being deprecated in the
upstream node_exporter.

Support for the old field names will be removed in a future version. A warning
will be logged if using the old field names when the integration is enabled.

## v0.21.2, v0.20.1

### Disabling of config retrieval endpoints

These two patch releases, as part of a fix for
[CVE-2021-41090](https://github.com/grafana/agent/security/advisories/GHSA-9c4x-5hgq-q3wh),
disable the `/-/config` and `/agent/api/v1/configs/{name}` endpoints by
default. Pass the `--config.enable-read-api` flag at the command line to
re-enable them.

## v0.21

### Integrations: Change in how instance labels are handled (Breaking change)

Integrations will now use a SUO-specific `instance` label value. Integrations
that apply to a whole machine or agent will continue to use `<agent machine
hostname>:<agent listen port>`, but integrations that connect to an external
system will now infer an appropriate value based on the config for that specific
integration. Please refer to the documentation for each integration for which
defaults are used.

*Note:* In some cases, a default value for `instance` cannot be inferred. This
is the case for mongodb_exporter and postgres_exporter if more than one SUO is
being connected to. In these cases, the instance value can be manually set by
configuring the `instance` field on the integration. This can also be useful if
two agents infer the same value for instance for the same integration.

As part of this change, the `agent_hostname` label is permanently affixed to
self-scraped integrations and cannot be disabled. This disambiguates multiple
agents using the same instance label for an integration, and allows users to
identify which agents need to be updated with an override for `instance`.

Both `use_hostname_label` and `replace_instance_label` are now both deprecated
and ignored from the YAML file, permanently treated as true. A future release
will remove these fields, causing YAML errors on load instead of being silently
ignored.

## v0.20

### Traces: Changes to receiver's TLS config (Breaking change).

Upgrading to OpenTelemetry v0.36.0 contains a change in the receivers TLS config.
TLS params have been changed from being squashed to being in its own block.
This affect the jaeger receiver's `remote_sampling` config.

Example old config:

```yaml
receivers:
  jaeger:
    protocols:
      grpc: null,
    remote_sampling:
      strategy_file: <file_path>
      insecure: true
```

Example new config:

```yaml
receivers:
  jaeger:
    protocols:
      grpc: null,
    remote_sampling:
      strategy_file: <file_path>
      tls:
        insecure: true
```

### Traces: push_config is no longer supported (Breaking change)

`push_config` was deprecated in favor of `remote_write` in v0.14.0, while
maintaining backwards compatibility.
Refer to the [deprecation announcement](#tempo-push_config-deprecation) for how to upgrade.

### Traces: legacy OTLP gRPC port no longer default port

OTLP gRPC receivers listen at port `4317` by default, instead of at port `55680`.
This goes in line with OTLP legacy port deprecation.

To upgrade, point the client instrumentation push endpoint to `:4317` if using
the default OTLP gRPC endpoint.

## v0.19

### Traces: Deprecation of "tempo" in config and metrics. (Deprecation)

The term `tempo` in the config has been deprecated of favor of `traces`. This
change is to make intent clearer.

Example old config:

```yaml
tempo:
  configs:
    - name: default
      receivers:
        jaeger:
          protocols:
            thrift_http:
```

Example of new config:
```yaml
traces:
  configs:
    - name: default
      receivers:
        jaeger:
          protocols:
            thrift_http:
```

Any tempo metrics have been renamed from `tempo_*` to `traces_*`.


### Tempo: split grouping by trace from tail sampling config (Breaking change)

Load balancing traces between agent instances has been moved from an embedded
functionality in tail sampling to its own configuration block.
This is done due to more processor benefiting from receiving consistently
receiving all spans for a trace in the same agent to be processed, such as
service graphs.

As a consequence, `tail_sampling.load_balancing` has been deprecated in favor of
a `load_balancing` block. Also, `port` has been renamed to `receiver_port` and
moved to the new `load_balancing` block.

Example old config:

```yaml
tail_sampling:
  policies:
    - always_sample:
  port: 4318
  load_balancing:
    exporter:
      insecure: true
    resolver:
      dns:
        hostname: agent
        port: 4318
```

Example new config:

```yaml
tail_sampling:
  policies:
    - always_sample:
load_balancing:
  exporter:
    insecure: true
  resolver:
    dns:
      hostname: agent
      port: 4318
  receiver_port: 4318
```

### Metrics: Deprecation of "prometheus" in config. (Deprecation)

The term `prometheus` in the config has been deprecated of favor of `metrics`. This
change is to make it clearer when referring to Prometheus or another
Prometheus-like database, and configuration of Grafana Agent to send metrics to
one of those systems.

Old configs will continue to work for now, but support for the old format will
eventually be removed. To migrate your config, change the `prometheus` key to
`metrics`.

Example old config:

```yaml
prometheus:
  configs:
    - name: default
      host_filter: false
      scrape_configs:
        - job_name: local_scrape
          static_configs:
            - targets: ['127.0.0.1:12345']
              labels:
                cluster: 'localhost'
      remote_write:
        - url: http://localhost:9009/api/prom/push
```

Example new config:

```yaml
metrics:
  configs:
    - name: default
      host_filter: false
      scrape_configs:
        - job_name: local_scrape
          static_configs:
            - targets: ['127.0.0.1:12345']
              labels:
                cluster: 'localhost'
      remote_write:
        - url: http://localhost:9009/api/prom/push
```

### Tempo: prom_instance rename (Breaking change)

As part of `prometheus` being renamed to `metrics`, the spanmetrics
`prom_instance` field has been renamed to `metrics_instance`. This is a breaking
change, and the old name will no longer work.

Example old config:

```yaml
tempo:
  configs:
  - name: default
    spanmetrics:
      prom_instance: default
```

Example new config:

```yaml
tempo:
  configs:
  - name: default
    spanmetrics:
      metrics_instance: default
```

### Logs: Deprecation of "loki" in config. (Deprecation)

The term `loki` in the config has been deprecated of favor of `logs`. This
change is to make it clearer when referring to Grafana Loki, and
configuration of Grafana Agent to send logs to Grafana Loki.

Old configs will continue to work for now, but support for the old format will
eventually be removed. To migrate your config, change the `loki` key to `logs`.

Example old config:

```yaml
loki:
  positions_directory: /tmp/loki-positions
  configs:
  - name: default
    clients:
      - url: http://localhost:3100/loki/api/v1/push
    scrape_configs:
    - job_name: system
      static_configs:
      - targets: ['localhost']
        labels:
          job: varlogs
          __path__: /var/log/*log
```

Example new config:

```yaml
logs:
  positions_directory: /tmp/loki-positions
  configs:
  - name: default
    clients:
      - url: http://localhost:3100/loki/api/v1/push
    scrape_configs:
    - job_name: system
      static_configs:
      - targets: ['localhost']
        labels:
          job: varlogs
          __path__: /var/log/*log
```

### Tempo: Deprecation of "loki" in config. (Deprecation)

As part of the `loki` to `logs` rename, parts of the automatic_logging component
in Tempo have been updated to refer to `logs_instance` instead.

Old configurations using `loki_name`, `loki_tag`, or `backend: loki` will
continue to work as of this version, but support for the old config format
will eventually be removed.

Example old config:

```yaml
tempo:
  configs:
  - name: default
    automatic_logging:
      backend: loki
      loki_name: default
      spans: true
      processes: true
      roots: true
    overrides:
      loki_tag: tempo
```

Example new config:

```yaml
tempo:
  configs:
  - name: default
    automatic_logging:
      backend: logs_instance
      logs_instance_name: default
      spans: true
      processes: true
      roots: true
    overrides:
      logs_instance_tag: tempo
```

## v0.18

### Tempo: Remote write TLS config

Tempo `remote_write` now supports configuring TLS settings in the trace
exporter's client. `insecure_skip_verify` is moved into this setting's block.

Old configurations with `insecure_skip_verify` outside `tls_config` will continue
to work as of this version, but support will eventually be removed.
If both `insecure_skip_verify` and `tls_config.insecure_skip_verify` are used,
then the latter take precedence.

Example old config:

```
tempo:
  configs:
    - name: default
      remote_write:
        - endpoint: otel-collector:55680
          insecure: true
          insecure_skip_verify: true
```

Example new config:

```
tempo:
  configs:
    - name: default
      remote_write:
        - endpoint: otel-collector:55680
          insecure: true
          tls_config:
            insecure_skip_verify: true
```

## v0.15

### Tempo: `automatic_logging` changes

Tempo automatic logging previously assumed that the operator wanted to log
to a Loki instance. With the addition of an option to log to stdout a new
field is required to maintain the old behavior.

Example old config:

```
tempo:
  configs:
  - name: default
    automatic_logging:
      loki_name: <some loki instance>
```

Example new config:

```
tempo:
  configs:
  - name: default
    automatic_logging:
      backend: loki
      loki_name: <some loki instance>
```

## v0.14

### Scraping Service security change

v0.14.0 changes the default behavior of the scraping service config management
API to reject all configuration files that read credentials from a file on disk.
This prevents malicious users from crafting an instance config file that read
arbitrary files on disk and send their contents to remote endpoints.

To revert to the old behavior, add `dangerous_allow_reading_files: true` in your
`scraping_service` config.

Example old config:

```yaml
prometheus:
  scraping_service:
    # ...
```

Example new config:

```yaml
prometheus:
  scraping_service:
    dangerous_allow_reading_files: true
    # ...
```

### SigV4 config change

v0.14.0 updates the internal Prometheus dependency to 2.26.0, which includes
native support for SigV4, but uses a slightly different configuration structure
than the Grafana Agent did.

To migrate, remove the `enabled` key from your `sigv4` configs. If `enabled` was
the only key, define sigv4 as an empty object: `sigv4: {}`.

Example old config:

```yaml
sigv4:
  enabled: true
  region: us-east-1
```

Example new config:

```yaml
sigv4:
  region: us-east-1
```

### Tempo: `push_config` deprecation

`push_config` is now deprecated in favor of a `remote_write` array which allows for sending spans to multiple endpoints.
`push_config` will be removed in a future release, and it is recommended to migrate to `remote_write` as soon as possible.

To migrate, move the batch options outside the `push_config` block.
Then, add a `remote_write` array and move the remaining of your `push_config` block inside it.

Example old config:

```yaml
tempo:
  configs:
    - name: default
      receivers:
        otlp:
          protocols:
            gpc:
      push_config:
        endpoint: otel-collector:55680
        insecure: true
        batch:
          timeout: 5s
          send_batch_size: 100
```

Example migrated config:

```yaml
tempo:
  configs:
    - name: default
      receivers:
        otlp:
          protocols:
            gpc:
      remote_write:
        - endpoint: otel-collector:55680
          insecure: true
      batch:
        timeout: 5s
        send_batch_size: 100
```


## v0.12

v0.12 had two breaking changes: the `tempo` and `loki` sections have been changed to require a list of `tempo`/`loki` configs rather than just one.

### Tempo Config Change

The Tempo config (`tempo` in the config file) has been changed to store
configs within a `configs` list. This allows for defining multiple Tempo
instances for collecting traces and forwarding them to different OTLP
endpoints.

To migrate, add a `configs:` array and move your existing config inside of it.
Give the element a `name: default` field.

Each config must have a unique non-empty name. `default` is recommended for users
that don't have other configs. The name of the config will be added as a
`tempo_config` label for metrics.

Example old config:

```yaml
tempo:
  receivers:
    jaeger:
      protocols:
        thrift_http:
  attributes:
    actions:
    - action: upsert
      key: env
      value: prod
  push_config:
    endpoint: otel-collector:55680
    insecure: true
    batch:
      timeout: 5s
      send_batch_size: 100
```

Example migrated config:

```yaml
tempo:
  configs:
  - name: default
    receivers:
      jaeger:
        protocols:
          thrift_http:
    attributes:
      actions:
      - action: upsert
        key: env
        value: prod
    push_config:
      endpoint: otel-collector:55680
      insecure: true
      batch:
        timeout: 5s
        send_batch_size: 100
```

### Loki Promtail Config Change

The Loki Promtail config (`loki` in the config file) has been changed to store
configs within a `configs` list. This allows for defining multiple Loki
Promtail instances for collecting logs and forwarding them to different Loki
servers.

To migrate, add a `configs:` array and move your existing config inside of it.
Give the element a `name: default` field.

Each config must have a unique non-empty name. `default` is recommended for users
that don't have other configs. The name of the config will be added as a
`loki_config` label for Loki Promtail metrics.

Example old config:

```yaml
loki:
  positions:
    filename: /tmp/positions.yaml
  clients:
    - url: http://loki:3100/loki/api/v1/push
  scrape_configs:
  - job_name: system
    static_configs:
      - targets:
        - localhost
        labels:
          job: varlogs
          __path__: /var/log/*log
```

Example migrated config:

```yaml
loki:
  configs:
  - name: default
    positions:
      filename: /tmp/positions.yaml
    clients:
      - url: http://loki:3100/loki/api/v1/push
    scrape_configs:
    - job_name: system
      static_configs:
        - targets:
          - localhost
          labels:
            job: varlogs
            __path__: /var/log/*log
```
