# Changelog

> _Contributors should read our [contributors guide][] for instructions on how
> to update the changelog._

This document contains a historical list of changes between releases. Only
changes that impact end-user behavior are listed; changes to documentation or
internal API changes are not present.

Main (unreleased)
-----------------

### Features

- Add support for converting Prometheus `file_sd_config` to `discovery.file`. (@erikbaranowski)

### Bugfixes

- Fix issue where `remote.http` incorrectly had a status of "Unknown" until the
  period specified by the polling frquency elapsed. (@rfratto)

### Enhancements

- Add [godeltaprof](https://github.com/grafana/godeltaprof) profiling types (`godeltaprof_memory`, `godeltaprof_mutex`, `godeltaprof_block`) to `pyroscope.scrape` component

v0.35.0-rc.0 (2023-07-13)
-------------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Breaking changes

- The algorithm for the "hash" action of `otelcol.processor.attributes` has changed.
  The change was made in PR [#22831](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22831) of opentelemetry-collector-contrib. (@ptodev)

- `otelcol.exporter.loki` now includes the instrumentation scope in its output. (@ptodev)

- `otelcol.extension.jaeger_remote_sampling` removes the `/` HTTP endpoint. The `/sampling` endpoint is still functional.
  The change was made in PR [#18070](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/18070) of opentelemetry-collector-contrib. (@ptodev)

- The field `version` and `auth` struct block from `walk_params` in `prometheus.exporter.snmp` and SNMP integration have been removed. The auth block now can be configured at top level, together with `modules` (@marctc)

- Rename `discovery.file` to `local.file_match` to make it more clear that it
  discovers file on the local filesystem, and so it doesn't get confused with
  Prometheus' file discovery. (@rfratto)

- Remove the `discovery_target_decode` function in favor of using discovery
  components to better match the behavior of Prometheus' service discovery.
  (@rfratto)

- In the traces subsystem for Static mode, some metrics are removed and others are renamed. (@ptodev)
  - Removed metrics:
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

  - Renamed metrics:
    - "traces_receiver_refused_spans" is renamed to "traces_receiver_refused_spans_total"
    - "traces_receiver_accepted_spans" is renamed to "traces_receiver_refused_spans_total"
    - "traces_exporter_sent_metric_points" is renamed to "traces_exporter_sent_metric_points_total"

- The `remote_sampling` block has been removed from `otelcol.receiver.jaeger`. (@ptodev)

### Deprecations

- `otelcol.exporter.jaeger` has been deprecated and will be removed in Agent v0.38.0. (@ptodev)

### Features

- The Pyroscope scrape component computes and sends delta profiles automatically when required to reduce bandwidth usage. (@cyriltovena)

- Support `stage.geoip` in `loki.process`. (@akselleirv)

- Integrations: Introduce the `squid` integration. (@armstrmi)

- Support custom fields in MMDB file for `stage.geoip`. (@akselleirv)

- Added json_path function to river stdlib. (@jkroepke)

- Add `format`, `join`, `tp_lower`, `replace`, `split`, `trim`, `trim_prefix`, `trim_suffix`, `trim_space`, `to_upper` functions to river stdlib. (@jkroepke)

- Flow UI: Add a view for listing the Agent's peers status when clustering is enabled. (@tpaschalis)

- Add a new CLI command `grafana-agent convert` for converting a river file from supported formats to river. (@erikbaranowski)

- Add support to the `grafana-agent run` CLI for converting a river file from supported formats to river. (@erikbaranowski)

- Add boringcrypto builds and docker images for Linux arm64 and x64. (@mattdurham)

- New Grafana Agent Flow components:

  - `discovery.file` discovers scrape targets from files. (@spartan0x117)
  - `discovery.kubelet` collect scrape targets from the Kubelet API. (@gcampbell12)
  - `module.http` runs a Grafana Agent Flow module loaded from a remote HTTP endpoint. (@spartan0x117)
  - `otelcol.processor.attributes` accepts telemetry data from other `otelcol`
    components and modifies attributes of a span, log, or metric. (@ptodev)
  - `prometheus.exporter.cloudwatch` - scrape AWS CloudWatch metrics (@thepalbi)
  - `prometheus.exporter.elasticsearch` collects metrics from Elasticsearch. (@marctc)
  - `prometheus.exporter.kafka` collects metrics from Kafka Server. (@oliver-zhang)
  - `prometheus.exporter.mongodb` collects metrics from MongoDB. (@marctc)
  - `prometheus.exporter.squid` collects metrics from a squid server. (@armstrmi)
  - `prometheus.operator.probes` - discovers Probe resources in your Kubernetes
    cluster and scrape the targets they reference. (@captncraig)
  - `pyroscope.ebpf` collects system-wide performance profiles from the current
    host (@korniltsev)
  - `otelcol.exporter.loadbalancing` - export traces and logs to multiple OTLP gRPC
    endpoints in a load-balanced way. (@ptodev)

- New Grafana Agent Flow command line utilities:

  - `grafana-agent tools prometheus.remote_write` holds a collection of remote
    write-specific tools. These have been ported over from the `agentctl` command. (@rfratto)

- A new `action` argument for `otelcol.auth.headers`. (@ptodev)

- New `metadata_keys` and `metadata_cardinality_limit` arguments for `otelcol.processor.batch`. (@ptodev)

- New `boolean_attribute` and `ottl_condition` sampling policies for `otelcol.processor.tail_sampling`. (@ptodev)

- A new `initial_offset` argument for `otelcol.receiver.kafka`. (@ptodev)

### Enhancements

- Attributes and blocks set to their default values will no longer be shown in the Flow UI. (@rfratto)

- Tanka config: retain cAdvisor metrics for system processes (Kubelet, Containerd, etc.) (@bboreham)

- Update cAdvisor dependency to v0.47.0. (@jcreixell)

- Upgrade and improve Cloudwatch exporter integration (@thepalbi)

- Update `node_exporter` dependency to v1.6.0. (@spartan0x117)

- Enable `prometheus.relabel` to work with Prometheus' Native Histograms. (@tpaschalis)

- Update `dnsmasq_exporter` to last version. (@marctc)

- Add deployment spec options to describe operator's Prometheus Config Reloader image. (@alekseybb197)

- Update `module.git` with basic and SSH key authentication support. (@djcode)

- Support `clustering` block in `prometheus.operator.servicemonitors` and `prometheus.operator.podmonitors` components to distribute
  targets amongst clustered agents. (@captncraig)

- Update `redis_exporter` dependency to v1.51.0. (@jcreixell)

- The Grafana Agent mixin now includes a dashboard for the logs pipeline. (@thampiotr)

- The Agent Operational dashboard of Grafana Agent mixin now has more descriptive panel titles, Y-axis units

- Add `write_relabel_config` to `prometheus.remote_write` (@jkroepke)

- Update OpenTelemetry Collector dependencies from v0.63.0 to v0.80.0. (@ptodev)

- Allow setting the node name for clustering with a command-line flag. (@tpaschalis)

- Allow `prometheus.exporter.snmp` and SNMP integration to be configured passing a YAML block. (@marctc)

- Some metrics have been added to the traces subsystem for Static mode. (@ptodev)
  - "traces_processor_batch_batch_send_size" (histogram)
  - "traces_processor_batch_batch_size_trigger_send_total" (counter)
  - "traces_processor_batch_metadata_cardinality" (gauge)
  - "traces_processor_batch_timeout_trigger_send_total" (counter)
  - "traces_rpc_server_duration" (histogram)
  - "traces_exporter_send_failed_metric_points_total" (counter)
  - "traces_exporter_send_failed_spans_total" (counter)
  - "traces_exporter_sent_spans_total" (counter)

- Added support for custom `length` time setting in Cloudwatch component and integration. (@thepalbi)

### Bugfixes

- Add signing region to remote.s3 component for use with custom endpoints so that Authorization Headers work correctly when
  proxying requests. (@mattdurham)

- Fix oauth default scope in `loki.source.azure_event_hubs`. (@akselleirv)

- Fix bug where `otelcol.exporter.otlphttp` ignores configuration for `traces_endpoint`, `metrics_endpoint`, and `logs_endpoint` attributes. (@SimoneFalzone)

- Fix issue in `prometheus.remote_write` where the `queue_config` and
  `metadata_config` blocks used incorrect defaults when not specified in the
  config file. (@rfratto)

- Fix issue where published RPMs were not signed. (@rfratto)

- Fix issue where flow mode exports labeled as "string or secret" could not be
  used in a binary operation. (@rfratto)

- Fix Grafana Agent mixin's "Agent Operational" dashboard expecting pods to always have `grafana-agent-.*` prefix. (@thampiotr)

- Change the HTTP Path and Data Path from the controller-local ID to the global ID for components loaded from within a module loader. (@spartan0x117)

- Fix bug where `stage.timestamp` in `loki.process` wasn't able to correctly
  parse timezones. This issue only impacts the dedicated `grafana-agent-flow`
  binary. (@rfratto)

- Fix bug where JSON requests to `loki.source.api` would not be handled correctly. This adds `/loki/api/v1/raw` and `/loki/api/v1/push` endpoints to `loki.source.api` and maps the `/api/v1/push` and `/api/v1/raw` to
  the `/loki` prefixed endpoints. (@mattdurham)

### Other changes

- Mongodb integration has been re-enabled. (@jcreixell, @marctc)
- Build with go 1.20.6 (@captncraig)

- Clustering for Grafana Agent in flow mode has graduated from experimental to beta.

v0.34.3 (2023-06-27)
--------------------

### Bugfixes

- Fixes a bug in conversion of OpenTelemetry histograms when exported to Prometheus. (@grcevski)
- Enforce sha256 digest signing for rpms enabling installation on FIPS-enabled OSes. (@kfriedrich123)
- Fix panic from improper startup ordering in `prometheus.operator.servicemonitors`. (@captncraig)

v0.34.2 (2023-06-20)
--------------------

### Enhancements

- Replace map cache in prometheus.relabel with an LRU cache. (@mattdurham)
- Integrations: Extend `statsd` integration to configure relay endpoint. (@arminaaki)

### Bugfixes

- Fix a bug where `prometheus.relabel` would not correctly relabel when there is a cache miss. (@thampiotr)
- Fix a bug where `prometheus.relabel` would not correctly relabel exemplars or metadata. (@tpaschalis)
- Fixes several issues with statsd exporter. (@jcreixell, @marctc)

### Other changes

- Mongodb integration has been disabled for the time being due to licensing issues. (@jcreixell)

v0.34.1 (2023-06-12)
--------------------

### Bugfixes

- Fixed application of sub-collector defaults using the `windows_exporter` integration or `prometheus.exporter.windows`. (@mattdurham)

- Fix issue where `remote.http` did not fail early if the initial request
  failed. This caused failed requests to initially export empty values, which
  could lead to propagating issues downstream to other components which expect
  the export to be non-empty. (@rfratto)

- Allow `bearerTokenFile` field to be used in ServiceMonitors. (@captncraig)

- Fix issue where metrics and traces were not recorded from components within modules. (@mattdurham)

- `service_name` label is inferred from discovery meta labels in `pyroscope.scrape` (@korniltsev)

### Other changes

- Add logging to failed requests in `remote.http`. (@rfratto)

v0.34.0 (2023-06-08)
--------------------

### Breaking changes

- The experimental dynamic configuration feature has been removed in favor of Flow mode. (@mattdurham)

- The `oracledb` integration configuration has removed a redundant field `metrics_scrape_interval`. Use the `scrape_interval` parameter of the integration if a custom scrape interval is required. (@schmikei)

- Upgrade the embedded windows_exporter to commit 79781c6. (@jkroepke)

- Prometheus exporters in Flow mode now set the `instance` label to a value similar to the one they used to have in Static mode (<hostname> by default, customized by some integrations). (@jcreixell)

- `phlare.scrape` and `phlare.write` have been renamed to `pyroscope.scrape` and `pyroscope.scrape`. (@korniltsev)

### Features

- New Grafana Agent Flow components:
  - `loki.source.api` - receive Loki log entries over HTTP (e.g. from other agents). (@thampiotr)
  - `prometheus.operator.servicemonitors` discovers ServiceMonitor resources in your Kubernetes cluster and scrape
    the targets they reference. (@captncraig, @marctc, @jcreixell)
  - `prometheus.receive_http` - receive Prometheus metrics over HTTP (e.g. from other agents). (@thampiotr)
  - `remote.vault` retrieves a secret from Vault. (@rfratto)
  - `prometheus.exporter.snowflake` collects metrics from a snowflake database (@jonathanWamsley)
  - `prometheus.exporter.mssql` collects metrics from Microsoft SQL Server (@jonathanwamsley)
  - `prometheus.exporter.oracledb` collects metrics from oracledb (@jonathanwamsley)
  - `prometheus.exporter.dnsmasq` collects metrics from a dnsmasq server. (@spartan0x117)
  - `loki.source.awsfirehose` - receive Loki log entries from AWS Firehose via HTTP (@thepalbi)
  - `discovery.http` service discovery via http. (@captncraig)

- Added new functions to the River standard library:
  - `coalesce` returns the first non-zero value from a list of arguments. (@jkroepke)
  - `nonsensitive` converts a River secret back into a string. (@rfratto)

### Enhancements

- Support to attach node metadata to pods and endpoints targets in
  `discovery.kubernetes`. (@laurovenancio)

- Support ability to add optional custom headers to `loki.write` endpoint block (@aos)

- Support in-memory HTTP traffic for Flow components. `prometheus.exporter`
  components will now export a target containing an internal HTTP address.
  `prometheus.scrape`, when given that internal HTTP address, will connect to
  the server in-memory, bypassing the network stack. Use the new
  `--server.http.memory-addr` flag to customize which address is used for
  in-memory traffic. (@rfratto)
- Disable node_exporter on Windows systems (@jkroepke)
- Operator support for OAuth 2.0 Client in LogsClientSpec (@DavidSpek)

- Support `clustering` block in `phlare.scrape` components to distribute
  targets amongst clustered agents. (@rfratto)

- Delete stale series after a single WAL truncate instead of two. (@rfratto)

- Update OracleDB Exporter dependency to 0.5.0 (@schmikei)

- Embed Google Fonts on Flow UI (@jkroepke)

- Enable Content-Security-Policies on Flow UI (@jkroepke)

- Update azure-metrics-exporter to v0.0.0-20230502203721-b2bfd97b5313 (@kgeckhart)

- Update azidentity dependency to v1.3.0. (@akselleirv)

- Add custom labels to journal entries in `loki.source.journal` (@sbhrule15)

- `prometheus.operator.podmonitors` and `prometheus.operator.servicemonitors` can now access cluster secrets for authentication to targets. (@captncraig)

### Bugfixes

- Fix `loki.source.(gcplog|heroku)` `http` and `grpc` blocks were overriding defaults with zero-values
  on non-present fields. (@thepalbi)

- Fix an issue where defining `logging` or `tracing` blocks inside of a module
  would generate a panic instead of returning an error. (@erikbaranowski)

- Fix an issue where not specifying either `http` nor `grpc` blocks could result
  in a panic for `loki.source.heroku` and `loki.source.gcplog` components. (@thampiotr)

- Fix an issue where build artifacts for IBM S390x were being built with the
  GOARCH value for the PPC64 instead. (tpaschalis)

- Fix an issue where the Grafana Agent Flow RPM used the wrong path for the
  environment file, preventing the service from loading. (@rfratto)

- Fix an issue where the cluster advertise address was overwriting the join
  addresses. (@laurovenancio)

- Fix targets deduplication when clustering mode is enabled. (@laurovenancio)

- Fix issue in operator where any version update will restart all agent pods simultaneously. (@captncraig)

- Fix an issue where `loki.source.journald` did not create the positions
  directory with the appropriate permissions. (@tpaschalis)

- Fix an issue where fanning out log entries to multiple `loki.process`
  components lead to a race condition. (@tpaschalis)

- Fix panic in `prometheus.operator.servicemonitors` from relabel rules without certain defaults. (@captncraig)

- Fix issue in modules export cache throwing uncomparable errors. (@mattdurham)

- Fix issue where the UI could not navigate to components loaded by modules. (@rfratto)

- Fix issue where using exporters inside modules failed due to not passing the in-memory address dialer. (@mattdurham)

- Add signing region to remote.s3 component for use with custom endpoints so that Authorization Headers work correctly when
  proxying requests. (@mattdurham)

- Fix missing `instance` key for `prometheus.exporter.dnsmasq` component. (@spartan0x117)

### Other changes

- Add metrics when clustering mode is enabled. (@rfratto)
- Document debug metric `loki_process_dropped_lines_by_label_total` in loki.process. (@akselleirv)

- Add `agent_wal_out_of_order_samples_total` metric to track samples received
  out of order. (@rfratto)

- Add CLI flag `--server.http.enable-pprof` to grafana-agent-flow to conditionally enable `/debug/pprof` endpoints (@jkroepke)

- Use Go 1.20.4 for builds. (@tpaschalis)

- Integrate the new ExceptionContext which was recently added to the Faro Web-SDK in the
  app_agent_receiver Payload. (@codecapitano)

- Flow clustering: clusters will now use 512 tokens per node for distributing
  work, leading to better distribution. However, rolling out this change will
  cause some incorrerct or missing assignments until all nodes are updated. (@rfratto)

- Change the Docker base image for Linux containers to `ubuntu:lunar`.
  (@rfratto)

v0.33.2 (2023-05-11)
--------------------

### Bugfixes

- Fix issue where component evaluation time was overridden by a "default
  health" message. (@rfratto)

- Honor timeout when trying to establish a connection to another agent in Flow
  clustering mode. (@rfratto)

- Fix an issue with the grafana/agent windows docker image entrypoint
  not targeting the right location for the config. (@erikbaranowski)

- Fix issue where the `node_exporter` integration and
  `prometheus.exporter.unix` `diskstat_device_include` component could not set
  the allowlist field for the diskstat collector. (@tpaschalis)

- Fix an issue in `loki.source.heroku` where updating the `labels` or `use_incoming_timestamp`
  would not take effect. (@thampiotr)

- Flow: Fix an issue within S3 Module where the S3 path was not parsed correctly when the
  path consists of a parent directory. (@jastisriradheshyam)

- Flow: Fix an issue on Windows where `prometheus.remote_write` failed to read
  WAL checkpoints. This issue led to memory leaks once the initial checkpoint
  was created, and prevented a fresh process from being able to deliver metrics
  at all. (@rfratto)

- Fix an issue where the `loki.source.kubernetes` component could lead to
  the Agent crashing due to a race condition. (@tpaschalis)

### Other changes

- The `phlare.scrape` Flow component `fetch profile failed` log has been set to
  `debug` instead of `error`. (@erikbaranowski)

v0.33.1 (2023-05-01)
--------------------

### Bugfixes

- Fix spelling of the `frequency` argument on the `local.file` component.
  (@tpaschalis)

- Fix bug where some capsule values (such as Prometheus receivers) could not
  properly be used as an argument to a module. (@rfratto)

- Fix version information not displaying correctly when passing the `--version`
  flag or in the `agent_build_info` metric. (@rfratto)

- Fix issue in `loki.source.heroku` and `loki.source.gcplog` where updating the
  component would cause Grafana Agent Flow's Prometheus metrics endpoint to
  return an error until the process is restarted. (@rfratto)

- Fix issue in `loki.source.file` where updating the component caused
  goroutines to leak. (@rfratto)

### Other changes

- Support Bundles report the status of discovered log targets. (@tpaschalis)

v0.33.0 (2023-04-25)
--------------------

### Breaking changes

- Support for 32-bit ARM builds is removed for the foreseeable future due to Go
  compiler issues. We will consider bringing back 32-bit ARM support once our Go
  compiler issues are resolved and 32-bit ARM builds are stable. (@rfratto)

- Agent Management: `agent_management.api_url` config field has been replaced by
`agent_management.host`. The API path and version is now defined by the Agent. (@jcreixell)

- Agent Management: `agent_management.protocol` config field now allows defining "http" and "https" explicitly. Previously, "http" was previously used for both, with the actual protocol used inferred from the api url, which led to confusion. When upgrading, make sure to set to "https" when replacing `api_url` with `host`. (@jcreixell)

- Agent Management: `agent_management.remote_config_cache_location` config field has been replaced by
`agent_management.remote_configuration.cache_location`. (@jcreixell)

- Remove deprecated symbolic links to to `/bin/agent*` in Docker containers,
  as planned in v0.31. (@tpaschalis)

### Deprecations

- [Dynamic Configuration](https://grafana.com/docs/agent/latest/cookbook/dynamic-configuration/) will be removed in v0.34. Grafana Agent Flow supersedes this functionality. (@mattdurham)

### Features

- New Grafana Agent Flow components:

  - `discovery.dns` DNS service discovery. (@captncraig)
  - `discovery.ec2` service discovery for aws ec2. (@captncraig)
  - `discovery.lightsail` service discovery for aws lightsail. (@captncraig)
  - `discovery.gce` discovers resources on Google Compute Engine (GCE). (@marctc)
  - `discovery.digitalocean` provides service discovery for DigitalOcean. (@spartan0x117)
  - `discovery.consul` service discovery for Consul. (@jcreixell)
  - `discovery.azure` provides service discovery for Azure. (@spartan0x117)
  - `module.file` runs a Grafana Agent Flow module loaded from a file on disk.
    (@erikbaranowski)
  - `module.git` runs a Grafana Agent Flow module loaded from a file within a
    Git repository. (@rfratto)
  - `module.string` runs a Grafana Agent Flow module passed to the component by
    an expression containing a string. (@erikbaranowski, @rfratto)
  - `otelcol.auth.oauth2` performs OAuth 2.0 authentication for HTTP and gRPC
    based OpenTelemetry exporters. (@ptodev)
  - `otelcol.extension.jaeger_remote_sampling` provides an endpoint from which to
    pull Jaeger remote sampling documents. (@joe-elliott)
  - `otelcol.exporter.logging` accepts OpenTelemetry data from other `otelcol` components and writes it to the console. (@erikbaranowski)
  - `otelcol.auth.sigv4` performs AWS Signature Version 4 (SigV4) authentication
    for making requests to AWS services via `otelcol` components that support
    authentication extensions. (@ptodev)
  - `prometheus.exporter.blackbox` collects metrics from Blackbox exporter. (@marctc)
  - `prometheus.exporter.mysql` collects metrics from a MySQL database. (@spartan0x117)
  - `prometheus.exporter.postgres` collects metrics from a PostgreSQL database. (@spartan0x117)
  - `prometheus.exporter.statsd` collects metrics from a Statsd instance. (@gaantunes)
  - `prometheus.exporter.snmp` collects metrics from SNMP exporter. (@marctc)
  - `prometheus.operator.podmonitors` discovers PodMonitor resources in your Kubernetes cluster and scrape
    the targets they reference. (@captncraig, @marctc, @jcreixell)
  - `prometheus.exporter.windows` collects metrics from a Windows instance. (@jkroepke)
  - `prometheus.exporter.memcached` collects metrics from a Memcached server. (@spartan0x117)
  - `loki.source.azure_event_hubs` reads messages from Azure Event Hub using Kafka and forwards them to other   `loki` components. (@akselleirv)

- Add support for Flow-specific system packages:

  - Flow-specific DEB packages. (@rfratto, @robigan)
  - Flow-specific RPM packages. (@rfratto, @robigan)
  - Flow-specific macOS Homebrew Formula. (@rfratto)
  - Flow-specific Windows installer. (@rfratto)

  The Flow-specific packages allow users to install and run Grafana Agent Flow
  alongside an existing installation of Grafana Agent.

- Agent Management: Add support for integration snippets. (@jcreixell)

- Flow: Introduce a gossip-over-HTTP/2 _clustered mode_. `prometheus.scrape`
  component instances can opt-in to distributing scrape load between cluster
  peers. (@tpaschalis)

### Enhancements

- Flow: Add retries with backoff logic to Phlare write component. (@cyriltovena)

- Operator: Allow setting runtimeClassName on operator-created pods. (@captncraig)

- Operator: Transparently compress agent configs to stay under size limitations. (@captncraig)

- Update Redis Exporter Dependency to v1.49.0. (@spartan0x117)

- Update Loki dependency to the k144 branch. (@andriikushch)

- Flow: Add OAUTHBEARER mechanism to `loki.source.kafka` using Azure as provider. (@akselleirv)

- Update Process Exporter dependency to v0.7.10. (@spartan0x117)

- Agent Management: Introduces backpressure mechanism for remote config fetching (obeys 429 request
  `Retry-After` header). (@spartan0x117)

- Flow: support client TLS settings (CA, client certificate, client key) being
  provided from other components for the following components:

  - `discovery.docker`
  - `discovery.kubernetes`
  - `loki.source.kafka`
  - `loki.source.kubernetes`
  - `loki.source.podlogs`
  - `loki.write`
  - `mimir.rules.kubernetes`
  - `otelcol.auth.oauth2`
  - `otelcol.exporter.jaeger`
  - `otelcol.exporter.otlp`
  - `otelcol.exporter.otlphttp`
  - `otelcol.extension.jaeger_remote_sampling`
  - `otelcol.receiver.jaeger`
  - `otelcol.receiver.kafka`
  - `phlare.scrape`
  - `phlare.write`
  - `prometheus.remote_write`
  - `prometheus.scrape`
  - `remote.http`

- Flow: support server TLS settings (client CA, server certificate, server key)
  being provided from other components for the following components:

  - `loki.source.syslog`
  - `otelcol.exporter.otlp`
  - `otelcol.extension.jaeger_remote_sampling`
  - `otelcol.receiver.jaeger`
  - `otelcol.receiver.opencensus`
  - `otelcol.receiver.zipkin`

- Flow: Define custom http method and headers in `remote.http` component (@jkroepke)

- Flow: Add config property to `prometheus.exporter.blackbox` to define the config inline (@jkroepke)

- Update Loki Dependency to k146 which includes configurable file watchers (@mattdurham)

### Bugfixes

- Flow: fix issue where Flow would return an error when trying to access a key
  of a map whose value was the zero value (`null`, `0`, `false`, `[]`, `{}`).
  Whether an error was returned depended on the internal type of the value.
  (@rfratto)

- Flow: fix issue where using the `jaeger_remote` sampler for the `tracing`
  block would fail to parse the response from the remote sampler server if it
  used strings for the strategy type. This caused sampling to fall back
  to the default rate. (@rfratto)

- Flow: fix issue where components with no arguments like `loki.echo` were not
  viewable in the UI. (@rfratto)

- Flow: fix deadlock in `loki.source.file` where terminating tailers would hang
  while flushing remaining logs, preventing `loki.source.file` from being able
  to update. (@rfratto)

- Flow: fix deadlock in `loki.process` where a component with no stages would
  hang forever on handling logs. (@rfratto)

- Fix issue where a DefaultConfig might be mutated during unmarshaling. (@jcreixell)

- Fix issues where CloudWatch Exporter cannot use FIPS Endpoints outside of USA regions (@aglees)

- Fix issue where scraping native Prometheus histograms would leak memory.
  (@rfratto)

- Flow: fix issue where `loki.source.docker` component could deadlock. (@tpaschalis)

- Flow: fix issue where `prometheus.remote_write` created unnecessary extra
  child directories to store the WAL in. (@rfratto)

- Fix internal metrics reported as invalid by promtool's linter. (@tpaschalis)

- Fix issues with cri stage which treats partial line coming from any stream as same. (@kavirajk @aglees)

- Operator: fix for running multiple operators with different `--agent-selector` flags. (@captncraig)

- Operator: respect FilterRunning on PodMonitor and ServiceMonitor resources to only scrape running pods. (@captncraig)

- Fixes a bug where the github exporter would get stuck in an infinite loop under certain conditions. (@jcreixell)

- Fix bug where `loki.source.docker` always failed to start. (@rfratto)

### Other changes

- Grafana Agent Docker containers and release binaries are now published for
  s390x. (@rfratto)

- Use Go 1.20.3 for builds. (@rfratto)

- Change the Docker base image for Linux containers to `ubuntu:kinetic`.
  (@rfratto)

- Update prometheus.remote_write defaults to match new prometheus
  remote-write defaults. (@erikbaranowski)

v0.32.1 (2023-03-06)
--------------------

### Bugfixes

- Flow: Fixes slow reloading of targets in `phlare.scrape` component. (@cyriltovena)

- Flow: add a maximum connection lifetime of one hour when tailing logs from
  `loki.source.kubernetes` and `loki.source.podlogs` to recover from an issue
  where the Kubernetes API server stops responding with logs without closing
  the TCP connection. (@rfratto)

- Flow: fix issue in `loki.source.kubernetes` where `__pod__uid__` meta label
  defaulted incorrectly to the container name, causing tailers to never
  restart. (@rfratto)

v0.32.0 (2023-02-28)
--------------------

### Breaking changes

- Support for the embedded Flow UI for 32-bit ARMv6 builds is temporarily
  removed. (@rfratto)

- Node Exporter configuration options changed to align with new upstream version (@Thor77):

  - `diskstats_ignored_devices` is now `diskstats_device_exclude` in agent configuration.
  - `ignored_devices` is now `device_exclude` in flow configuration.

- Some blocks in Flow components have been merged with their parent block to make the block hierarchy smaller:

  - `discovery.docker > http_client_config` is merged into the `discovery.docker` block. (@erikbaranowski)
  - `discovery.kubernetes > http_client_config` is merged into the `discovery.kubernetes` block. (@erikbaranowski)
  - `loki.source.kubernetes > client > http_client_config` is merged into the `client` block. (@erikbaranowski)
  - `loki.source.podlogs > client > http_client_config` is merged into the `client` block. (@erikbaranowski)
  - `loki.write > endpoint > http_client_config` is merged into the `endpoint` block. (@erikbaranowski)
  - `mimir.rules.kubernetes > http_client_config` is merged into the `mimir.rules.kubernetes` block. (@erikbaranowski)
  - `otelcol.receiver.opencensus > grpc` is merged into the `otelcol.receiver.opencensus` block. (@ptodev)
  - `otelcol.receiver.zipkin > http` is merged into the `otelcol.receiver.zipkin` block. (@ptodev)
  - `phlare.scrape > http_client_config` is merged into the `phlare.scrape` block. (@erikbaranowski)
  - `phlare.write > endpoint > http_client_config` is merged into the `endpoint` block. (@erikbaranowski)
  - `prometheus.remote_write > endpoint > http_client_config` is merged into the `endpoint` block. (@erikbaranowski)
  - `prometheus.scrape > http_client_config` is merged into the `prometheus.scrape` block. (@erikbaranowski)

- The `loki.process` component now uses a combined name for stages, simplifying
  the block hierarchy. For example, the `stage > json` block hierarchy is now a
  single block called `stage.json`. All stage blocks in `loki.process` have
  been updated to use this simplified hierarchy. (@tpaschalis)

- `remote.s3` `client_options` block has been renamed to `client`. (@mattdurham)

- Renamed `prometheus.integration.node_exporter` to `prometheus.exporter.unix`. (@jcreixell)

- As first announced in v0.30, support for the `EXPERIMENTAL_ENABLE_FLOW`
  environment variable has been removed in favor of `AGENT_MODE=flow`.
  (@rfratto)

### Features

- New integrations:

  - `oracledb` (@schmikei)
  - `mssql` (@binaryfissiongames)
  - `cloudwatch metrics` (@thepalbi)
  - `azure` (@kgeckhart)
  - `gcp` (@kgeckhart, @ferruvich)

- New Grafana Agent Flow components:

  - `loki.echo` writes received logs to stdout. (@tpaschalis, @rfratto)
  - `loki.source.docker` reads logs from Docker containers and forwards them to
    other `loki` components. (@tpaschalis)
  - `loki.source.kafka` reads logs from Kafka events and forwards them to other
    `loki` components. (@erikbaranowski)
  - `loki.source.kubernetes_events` watches for Kubernetes Events and converts
    them into log lines to forward to other `loki` components. It is the
    equivalent of the `eventhandler` integration. (@rfratto)
  - `otelcol.processor.tail_sampling` samples traces based on a set of defined
    policies from `otelcol` components before forwarding them to other
    `otelcol` components. (@erikbaranowski)
  - `prometheus.exporter.apache` collects metrics from an apache web server
    (@captncraig)
  - `prometheus.exporter.consul` collects metrics from a consul installation
    (@captncraig)
  - `prometheus.exporter.github` collects metrics from GitHub (@jcreixell)
  - `prometheus.exporter.process` aggregates and collects metrics by scraping
    `/proc`. (@spartan0x117)
  - `prometheus.exporter.redis` collects metrics from a redis database
    (@spartan0x117)

### Enhancements

- Flow: Support `keepequal` and `dropequal` actions for relabeling. (@cyriltovena)

- Update Prometheus Node Exporter integration to v1.5.0. (@Thor77)

- Grafana Agent Flow will now reload the config file when `SIGHUP` is sent to
  the process. (@rfratto)

- If using the official RPM and DEB packages for Grafana Agent, invoking
  `systemctl reload grafana-agent` will now reload the configuration file.
  (@rfratto)

- Flow: the `loki.process` component now implements all the same processing
  stages as Promtail's pipelines. (@tpaschalis)

- Flow: new metric for `prometheus.scrape` -
  `agent_prometheus_scrape_targets_gauge`. (@ptodev)

- Flow: new metric for `prometheus.scrape` and `prometheus.relabel` -
  `agent_prometheus_forwarded_samples_total`. (@ptodev)

- Flow: add `constants` into the standard library to expose the hostname, OS,
  and architecture of the system Grafana Agent is running on. (@rfratto)

- Flow: add timeout to loki.source.podlogs controller setup. (@polyrain)

### Bugfixes

- Fixed a reconciliation error in Grafana Agent Operator when using `tlsConfig`
  on `Probe`. (@supergillis)

- Fix issue where an empty `server:` config stanza would cause debug-level logging.
  An empty `server:` is considered a misconfiguration, and thus will error out.
  (@neomantra)

- Flow: fix an error where some error messages that crossed multiple lines
  added extra an extra `|` character when displaying the source file on the
  starting line. (@rfratto)

- Flow: fix issues in `agent fmt` where adding an inline comment on the same
  line as a `[` or `{` would cause indentation issues on subsequent lines.
  (@rfratto)

- Flow: fix issues in `agent fmt` where line comments in arrays would be given
  the wrong identation level. (@rfratto)

- Flow: fix issues with `loki.file` and `loki.process` where deadlock contention or
  logs fail to process. (@mattdurham)

- Flow: `oauth2 > tls_config` was documented as a block but coded incorrectly as
  an attribute. This is now a block in code. This impacted `discovery.docker`,
  `discovery.kubernetes`, `loki.source.kubernetes`, `loki.write`,
  `mimir.rules.kubernetes`, `phlare.scrape`, `phlare.write`,
  `prometheus.remote_write`, `prometheus.scrape`, and `remote.http`
  (@erikbaranowski)

- Flow: Fix issue where using `river:",label"` causes the UI to return nothing. (@mattdurham)

### Other changes

- Use Go 1.20 for builds. (@rfratto)

- The beta label from Grafana Agent Flow has been removed. A subset of Flow
  components are still marked as beta or experimental:

  - `loki.echo` is explicitly marked as beta.
  - `loki.source.kubernetes` is explicitly marked as experimental.
  - `loki.source.podlogs` is explicitly marked as experimental.
  - `mimir.rules.kubernetes` is explicitly marked as beta.
  - `otelcol.processor.tail_sampling` is explicitly marked as beta.
  - `otelcol.receiver.loki` is explicitly marked as beta.
  - `otelcol.receiver.prometheus` is explicitly marked as beta.
  - `phlare.scrape` is explicitly marked as beta.
  - `phlare.write` is explicitly marked as beta.

v0.31.3 (2023-02-13)
--------------------

### Bugfixes

- `loki.source.cloudflare`: fix issue where the `zone_id` argument
  was being ignored, and the `api_token` argument was being used for the zone
  instead. (@rfratto)

- `loki.source.cloudflare`: fix issue where `api_token` argument was not marked
  as a sensitive field. (@rfratto)

v0.31.2 (2023-02-08)
--------------------

### Other changes

- In the Agent Operator, upgrade the `prometheus-config-reloader` dependency
  from version 0.47.0 to version 0.62.0. (@ptodev)

v0.31.1 (2023-02-06)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Breaking changes

- All release Windows `.exe` files are now published as a zip archive.
  Previously, `grafana-agent-installer.exe` was unzipped. (@rfratto)

### Other changes

- Support Go 1.20 for builds. Official release binaries are still produced
  using Go 1.19. (@rfratto)

v0.31.0 (2023-01-31)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Breaking changes

- Release binaries (including inside Docker containers) have been renamed to be
  prefixed with `grafana-` (@rfratto):

  - `agent` is now `grafana-agent`.
  - `agentctl` is now `grafana-agentctl`.
  - `agent-operator` is now `grafana-agent-operator`.

### Deprecations

- A symbolic link in Docker containers from the old binary name to the new
  binary name has been added. These symbolic links will be removed in v0.33. (@rfratto)

### Features

- New Grafana Agent Flow components:

  - `loki.source.cloudflare` reads logs from Cloudflare's Logpull API and
    forwards them to other `loki` components. (@tpaschalis)
  - `loki.source.gcplog` reads logs from GCP cloud resources using Pub/Sub
    subscriptions and forwards them to other `loki` components. (@tpaschalis)
  - `loki.source.gelf` listens for Graylog logs. (@mattdurham)
  - `loki.source.heroku` listens for Heroku messages over TCP a connection and
    forwards them to other `loki` components. (@erikbaranowski)
  - `loki.source.journal` read messages from systemd journal. (@mattdurham)
  - `loki.source.kubernetes` collects logs from Kubernetes pods using the
    Kubernetes API. (@rfratto)
  - `loki.source.podlogs` discovers PodLogs resources on Kubernetes and
    uses the Kubernetes API to collect logs from the pods specified by the
    PodLogs resource. (@rfratto)
  - `loki.source.syslog` listens for Syslog messages over TCP and UDP
    connections and forwards them to other `loki` components. (@tpaschalis)
  - `loki.source.windowsevent` reads logs from Windows Event Log. (@mattdurham)
  - `otelcol.exporter.jaeger` forwards OpenTelemetry data to a Jaeger server.
    (@erikbaranowski)
  - `otelcol.exporter.loki` forwards OTLP-formatted data to compatible `loki`
    receivers. (@tpaschalis)
  - `otelcol.receiver.kafka` receives telemetry data from Kafka. (@rfratto)
  - `otelcol.receiver.loki` receives Loki logs, converts them to the OTLP log
    format and forwards them to other `otelcol` components. (@tpaschalis)
  - `otelcol.receiver.opencensus` receives OpenConsensus-formatted traces or
    metrics. (@ptodev)
  - `otelcol.receiver.zipkin` receives Zipkin-formatted traces. (@rfratto)
  - `phlare.scrape` collects application performance profiles. (@cyriltovena)
  - `phlare.write` sends application performance profiles to Grafana Phlare.
    (@cyriltovena)
  - `mimir.rules.kubernetes` discovers `PrometheusRule` Kubernetes resources and
    loads them into a Mimir instance. (@Logiraptor)

- Flow components which work with relabeling rules (`discovery.relabel`,
  `prometheus.relabel` and `loki.relabel`) now export a new value named Rules.
  This value returns a copy of the currently configured rules. (@tpaschalis)

- New experimental feature: agent-management. Polls configured remote API to fetch new configs. (@spartan0x117)

- Introduce global configuration for logs. (@jcreixell)

### Enhancements

- Handle faro-web-sdk `View` meta in app_agent_receiver. (@rlankfo)

- Flow: the targets in debug info from `loki.source.file` are now individual blocks. (@rfratto)

- Grafana Agent Operator: add [promtail limit stage](https://grafana.com/docs/loki/latest/clients/promtail/stages/limit/) to the operator. (@spartan0x117)

### Bugfixes

- Flow UI: Fix the issue with messy layout on the component list page while
  browser window resize (@xiyu95)

- Flow UI: Display the values of all attributes unless they are nil. (@ptodev)

- Flow: `prometheus.relabel` and `prometheus.remote_write` will now error if they have exited. (@ptodev)

- Flow: Fix issue where negative numbers would convert to floating-point values
  incorrectly, treating the sign flag as part of the number. (@rfratto)

- Flow: fix a goroutine leak when `loki.source.file` is passed more than one
  target with identical set of public labels. (@rfratto)

- Fix issue where removing and re-adding log instance configurations causes an
  error due to double registration of metrics (@spartan0x117, @jcreixell)

### Other changes

- Use Go 1.19.4 for builds. (@erikbaranowski)

- New windows containers for agent and agentctl. These can be found moving forward with the ${Version}-windows tags for grafana/agent and grafana/agentctl docker images (@erikbaranowski)

v0.30.2 (2023-01-11)
--------------------

### Bugfixes

- Flow: `prometheus.relabel` will no longer modify the labels of the original
  metrics, which could lead to the incorrect application of relabel rules on
  subsequent relabels. (@rfratto)

- Flow: `loki.source.file` will no longer deadlock other components if log
  lines cannot be sent to Loki. `loki.source.file` will wait for 5 seconds per
  file to finish flushing read logs to the client, after which it will drop
  them, resulting in lost logs. (@rfratto)

- Operator: Fix the handling of the enableHttp2 field as a boolean in
  `pod_monitor` and `service_monitor` templates. (@tpaschalis)

v0.30.1 (2022-12-23)
--------------------

### Bugfixes

- Fix issue where journald support was accidentally removed. (@tpaschalis)

- Fix issue where some traces' metrics where not collected. (@marctc)

v0.30.0 (2022-12-20)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Breaking changes

- The `ebpf_exporter` integration has been removed due to issues with static
  linking. It may be brought back once these are resolved. (@tpaschalis)

### Deprecations

- The `EXPERIMENTAL_ENABLE_FLOW` environment variable is deprecated in favor of
  `AGENT_MODE=flow`. Support for `EXPERIMENTAL_ENABLE_FLOW` will be removed in
  v0.32. (@rfratto)

### Features

- `grafana-agent-operator` supports oauth2 as an authentication method for
  remote_write. (@timo-42)

- Grafana Agent Flow: Add tracing instrumentation and a `tracing` block to
  forward traces to `otelcol` component. (@rfratto)

- Grafana Agent Flow: Add a `discovery_target_decode` function to decode a JSON
  array of discovery targets corresponding to Prometheus' HTTP and file service
  discovery formats. (@rfratto)

- New Grafana Agent Flow components:

  - `remote.http` polls an HTTP URL and exposes the response body as a string
    or secret to other components. (@rfratto)

  - `discovery.docker` discovers Docker containers from a Docker Engine host.
    (@rfratto)

  - `loki.source.file` reads and tails files for log entries and forwards them
    to other `loki` components. (@tpaschalis)

  - `loki.write` receives log entries from other `loki` components and sends
    them over to a Loki instance. (@tpaschalis)

  - `loki.relabel` receives log entries from other `loki` components and
    rewrites their label set. (@tpaschalis)

  - `loki.process` receives log entries from other `loki` components and runs
    one or more processing stages. (@tpaschalis)

  - `discovery.file` discovers files on the filesystem following glob
    patterns. (@mattdurham)

- Integrations: Introduce the `snowflake` integration. (@binaryfissiongames)

### Enhancements

- Update agent-loki.yaml to use environment variables in the configuration file (@go4real)

- Integrations: Always use direct connection in mongodb_exporter integration. (@v-zhuravlev)

- Update OpenTelemetry Collector dependency to v0.63.1. (@tpaschalis)

- riverfmt: Permit empty blocks with both curly braces on the same line.
  (@rfratto)

- riverfmt: Allow function arguments to persist across different lines.
  (@rfratto)

- Flow: The HTTP server will now start before the Flow controller performs the
  initial load. This allows metrics and pprof data to be collected during the
  first load. (@rfratto)

- Add support for using a [password map file](https://github.com/oliver006/redis_exporter/blob/master/contrib/sample-pwd-file.json) in `redis_exporter`. (@spartan0x117)

- Flow: Add support for exemplars in Prometheus component pipelines. (@rfratto)

- Update Prometheus dependency to v2.40.5. (@rfratto)

- Update Promtail dependency to k127. (@rfratto)

- Native histograms are now supported in the static Grafana Agent and in
  `prometheus.*` Flow components. Native histograms will be automatically
  collected from supported targets. remote_write must be configured to forward
  native histograms from the WAL to the specified endpoints. (@rfratto)

- Flow: metrics generated by upstream OpenTelemetry Collector components are
  now exposed at the `/metrics` endpoint of Grafana Agent Flow. (@rfratto)

### Bugfixes

- Fix issue where whitespace was being sent as part of password when using a
  password file for `redis_exporter`. (@spartan0x117)

- Flow UI: Fix issue where a configuration block referencing a component would
  cause the graph page to fail to load. (@rfratto)

- Remove duplicate `oauth2` key from `metricsinstances` CRD. (@daper)

- Fix issue where on checking whether to restart integrations the Integration
  Manager was comparing configs with secret values scrubbed, preventing reloads
  if only secrets were updated. (@spartan0x117)

### Other changes

- Grafana Agent Flow has graduated from experimental to beta.

v0.29.0 (2022-11-08)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Breaking changes

- JSON-encoded traces from OTLP versions earlier than 0.16.0 are no longer
  supported. (@rfratto)

### Deprecations

- The binary names `agent`, `agentctl`, and `agent-operator` have been
  deprecated and will be renamed to `grafana-agent`, `grafana-agentctl`, and
  `grafana-agent-operator` in the v0.31.0 release.

### Features

- Add `agentctl test-logs` command to allow testing log configurations by redirecting
  collected logs to standard output. This can be useful for debugging. (@jcreixell)

- New Grafana Agent Flow components:

  - `otelcol.receiver.otlp` receives OTLP-formatted traces, metrics, and logs.
    Data can then be forwarded to other `otelcol` components. (@rfratto)

  - `otelcol.processor.batch` batches data from `otelcol` components before
    forwarding it to other `otelcol` components. (@rfratto)

  - `otelcol.exporter.otlp` accepts data from `otelcol` components and sends
    it to a gRPC server using the OTLP protocol. (@rfratto)

  - `otelcol.exporter.otlphttp` accepts data from `otelcol` components and
    sends it to an HTTP server using the OTLP protocol. (@tpaschalis)

  - `otelcol.auth.basic` performs basic authentication for `otelcol`
    components that support authentication extensions. (@rfratto)

  - `otelcol.receiver.jeager` receives Jaeger-formatted traces. Data can then
    be forwarded to other `otelcol` components. (@rfratto)

  - `otelcol.processor.memory_limiter` periodically checks memory usage and
    drops data or forces a garbage collection if the defined limits are
    exceeded. (@tpaschalis)

  - `otelcol.auth.bearer` performs bearer token authentication for `otelcol`
    components that support authentication extensions. (@rfratto)

  - `otelcol.auth.headers` attaches custom request headers to `otelcol`
    components that support authentication extensions. (@rfratto)

  - `otelcol.receiver.prometheus` receives Prometheus metrics, converts them
    to the OTLP metric format and forwards them to other `otelcol` components.
    (@tpaschalis)

  - `otelcol.exporter.prometheus` forwards OTLP-formatted data to compatible
    `prometheus` components. (@rfratto)

- Flow: Allow config blocks to reference component exports. (@tpaschalis)

- Introduce `/-/support` endpoint for generating 'support bundles' in static
  agent mode. Support bundles are zip files of commonly-requested information
  that can be used to debug a running agent. (@tpaschalis)

### Enhancements

- Update OpenTelemetry Collector dependency to v0.61.0. (@rfratto)

- Add caching to Prometheus relabel component. (@mattdurham)

- Grafana Agent Flow: add `agent_resources_*` metrics which explain basic
  platform-agnostic metrics. These metrics assist with basic monitoring of
  Grafana Agent, but are not meant to act as a replacement for fully featured
  components like `prometheus.integration.node_exporter`. (@rfratto)

- Enable field label in TenantStageSpec of PodLogs pipeline. (@siiimooon)

- Enable reporting of enabled integrations. (@marctc)

- Grafana Agent Flow: `prometheus.remote_write` and `prometheus.relabel` will
  now export receivers immediately, removing the need for dependant components
  to be evaluated twice at process startup. (@rfratto)

- Add missing setting to configure instance key for Eventhandler integration. (@marctc)

- Update Prometheus dependency to v2.39.1. (@rfratto)

- Update Promtail dependency to weekly release k122. (@rfratto)

- Tracing: support the `num_traces` and `expected_new_traces_per_sec` configuration parameters in the tail_sampling processor. (@ptodev)

### Bugfixes

- Remove empty port from the `apache_http` integration's instance label. (@katepangLiu)

- Fix identifier on target creation for SNMP v2 integration. (@marctc)

- Fix bug when specifying Blackbox's modules when using Blackbox integration. (@marctc)

- Tracing: fix a panic when the required `protocols` field was not set in the `otlp` receiver. (@ptodev)

- Support Bearer tokens for metric remote writes in the Grafana Operator (@jcreixell, @marctc)

### Other changes

- Update versions of embedded Prometheus exporters used for integrations:

  - Update `github.com/prometheus/statsd_exporter` to `v0.22.8`. (@captncraig)

  - Update `github.com/prometheus-community/postgres_exporter` to `v0.11.1`. (@captncraig)

  - Update `github.com/prometheus/memcached_exporter` to `v0.10.0`. (@captncraig)

  - Update `github.com/prometheus-community/elasticsearch_exporter` to `v1.5.0`. (@captncraig)

  - Update `github.com/prometheus/mysqld_exporter` to `v0.14.0`. (@captncraig)

  - Update `github.com/prometheus/consul_exporter` to `v0.8.0`. (@captncraig)

  - Update `github.com/ncabatoff/process-exporter` to `v0.7.10`. (@captncraig)

  - Update `github.com/prometheus-community/postgres_exporter` to `v0.11.1`. (@captncraig)

- Use Go 1.19.3 for builds. (@rfratto)

v0.28.1 (2022-11-03)
--------------------

### Security

- Update Docker base image to resolve OpenSSL vulnerabilities CVE-2022-3602 and
  CVE-2022-3786. Grafana Agent does not use OpenSSL, so we do not believe it is
  vulnerable to these issues, but the base image has been updated to remove the
  report from image scanners. (@rfratto)

v0.28.0 (2022-09-29)
--------------------

### Features

- Introduce Grafana Agent Flow, an experimental "programmable pipeline" runtime
  mode which improves how to configure and debug Grafana Agent by using
  components. (@captncraig, @karengermond, @marctc, @mattdurham, @rfratto,
  @rlankfo, @tpaschalis)

- Introduce Blackbox exporter integration. (@marctc)

### Enhancements

- Update Loki dependency to v2.6.1. (@rfratto)

### Bugfixes

### Other changes

- Fix relabel configs in sample agent-operator manifests (@hjet)

- Operator no longer set the `SecurityContext.Privileged` flag in the `config-reloader` container. (@hsyed-dojo)

- Add metrics for config reloads and config hash (@jcreixell)

v0.27.1 (2022-09-09)
--------------------

> **NOTE**: ARMv6 Docker images are no longer being published.
>
> We have stopped publishing Docker images for ARMv6 platforms.
> This is due to the new Ubuntu base image we are using that does not support ARMv6.
> The new Ubuntu base image has less reported CVEs, and allows us to provide more
> secure Docker images. We will still continue to publish ARMv6 release binaries and
> deb/rpm packages.

### Other Changes

- Switch docker image base from debian to ubuntu. (@captncraig)

v0.27.0 (2022-09-01)
--------------------

### Features

- Integrations: (beta) Add vmware_exporter integration (@rlankfo)

- App agent receiver: add Event kind to payload (@domasx2)

### Enhancements

- Tracing: Introduce a periodic appender to the remotewriteexporter to control sample rate. (@mapno)

- Tracing: Update OpenTelemetry dependency to v0.55.0. (@rfratto, @mapno)

- Add base agent-operator jsonnet library and generated manifests (@hjet)

- Add full (metrics, logs, K8s events) sample agent-operator jsonnet library and gen manifests (@hjet)

- Introduce new configuration fields for disabling Keep-Alives and setting the
  IdleConnectionTimeout when scraping. (@tpaschalis)

- Add field to Operator CRD to disable report usage functionality. (@marctc)

### Bugfixes

- Tracing: Fixed issue with the PromSD processor using the `connection` method to discover the IP
  address.  It was failing to match because the port number was included in the address string. (@jphx)

- Register prometheus discovery metrics. (@mattdurham)

- Fix seg fault when no instance parameter is provided for apache_http integration, using integrations-next feature flag. (@rgeyer)

- Fix grafanacloud-install.ps1 web request internal server error when fetching config. (@rlankfo)

- Fix snmp integration not passing module or walk_params parameters when scraping. (@rgeyer)

- Fix unmarshal errors (key "<walk_param name>" already set in map) for snmp integration config when walk_params is defined, and the config is reloaded. (@rgeyer)

### Other changes

- Update several go dependencies to resolve warnings from certain security scanning tools. None of the resolved vulnerabilities were known to be exploitable through the agent. (@captncraig)

- It is now possible to compile Grafana Agent using Go 1.19. (@rfratto)

v0.26.1 (2022-07-25)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Breaking changes

- Change windows certificate store so client certificate is no longer required in store. (@mattdurham)

### Bugfixes

- Operator: Fix issue where configured `targetPort` ServiceMonitors resulted in
  generating an incorrect scrape_config. (@rfratto)

- Build the Linux/AMD64 artifacts using the opt-out flag for the ebpf_exporter. (@tpaschalis)

v0.26.0 (2022-07-18)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Breaking changes

- Deprecated `server` YAML block fields have now been removed in favor of the
  command-line flags that replaced them. These fields were originally
  deprecated in v0.24.0. (@rfratto)

- Changed tail sampling policies to be configured as in the OpenTelemetry
  Collector. (@mapno)

### Features

- Introduce Apache HTTP exporter integration. (@v-zhuravlev)

- Introduce eBPF exporter integration. (@tpaschalis)

### Enhancements

- Truncate all records in WAL if repair attempt fails. (@rlankfo)

### Bugfixes

- Relative symlinks for promtail now work as expected. (@RangerCD, @mukerjee)

- Fix rate limiting implementation for the app agent receiver integration. (@domasx2)

- Fix mongodb exporter so that it now collects all metrics. (@mattdurham)

v0.25.1 (2022-06-16)
--------------------

### Bugfixes

- Integer types fail to unmarshal correctly in operator additional scrape configs. (@rlankfo)

- Unwrap replayWAL error before attempting corruption repair. (@rlankfo)

v0.25.0 (2022-06-06)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Breaking changes

- Traces: Use `rpc.grpc.status_code` attribute to determine
  span failed in the service graph processor (@rcrowe)

### Features

- Add HTTP endpoints to fetch active instances and targets for the Logs subsystem.
  (@marctc)

- (beta) Add support for using windows certificate store for TLS connections. (@mattdurham)

- Grafana Agent Operator: add support for integrations through an `Integration`
  CRD which is discovered by `GrafanaAgent`. (@rfratto)

- (experimental) Add app agent receiver integration. This depends on integrations-next being enabled
  via the `integrations-next` feature flag. Use `-enable-features=integrations-next` to use
  this integration. (@kpelelis, @domas)

- Introduce SNMP exporter integration. (@v-zhuravlev)

- Configure the agent to report the use of feature flags to grafana.com. (@marctc)

### Enhancements

- integrations-next: Integrations using autoscrape will now autoscrape metrics
  using in-memory connections instead of connecting to themselves over the
  network. As a result of this change, the `client_config` field has been
  removed. (@rfratto)

- Enable `proxy_url` support on `oauth2` for metrics and logs (update **prometheus/common** dependency to `v0.33.0`). (@martin-jaeger-maersk)

- `extra-scrape-metrics` can now be enabled with the `--enable-features=extra-scrape-metrics` feature flag. See <https://prometheus.io/docs/prometheus/2.31/feature_flags/#extra-scrape-metrics> for details. (@rlankfo)

- Resolved issue in v2 integrations where if an instance name was a prefix of another the route handler would fail to
  match requests on the longer name (@mattdurham)

- Set `include_metadata` to true by default for OTLP traces receivers (@mapno)

### Bugfixes

- Scraping service was not honoring the new server grpc flags `server.grpc.address`.  (@mattdurham)

### Other changes

- Update base image of official Docker containers from Debian buster to Debian
  bullseye. (@rfratto)

- Use Go 1.18 for builds. (@rfratto)

- Add `metrics` prefix to the url of list instances endpoint (`GET
  /agent/api/v1/instances`) and list targets endpoint (`GET
  /agent/api/v1/metrics/targets`). (@marctc)

- Add extra identifying labels (`job`, `instance`, `agent_hostname`) to eventhandler integration. (@hjet)

- Add `extra_labels` configuration to eventhandler integration. (@hjet)

v0.24.2 (2022-05-02)
--------------------

### Bugfixes

- Added configuration watcher delay to prevent race condition in cases where scraping service mode has not gracefully exited. (@mattdurham)

### Other changes

- Update version of node_exporter to include additional metrics for osx. (@v-zhuravlev)

v0.24.1 (2022-04-14)
--------------------

### Bugfixes

- Add missing version information back into `agentctl --version`. (@rlankfo)

- Bump version of github-exporter to latest upstream SHA 284088c21e7d, which
  includes fixes from bugs found in their latest tag. This includes a fix
  where not all releases where retrieved when pulling release information.
  (@rfratto)

- Set the `Content-Type` HTTP header to `application/json` for API endpoints
  returning json objects. (@marctc)

- Operator: fix issue where a `username_file` field was incorrectly set.
  (@rfratto)

- Initialize the logger with default `log_level` and `log_format` parameters.
  (@tpaschalis)

### Other changes

- Embed timezone data to enable Promtail pipelines using the `location` field
  on Windows machines. (@tpaschalis)

v0.24.0 (2022-04-07)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.
>
> **GRAFANA AGENT OPERATOR USERS**: As of this release, Grafana Agent Operator
> does not support versions of Grafana Agent prior to v0.24.0.

### Breaking changes

- The following metrics will now be prefixed with `agent_dskit_` instead of
  `cortex_`: `cortex_kv_request_duration_seconds`,
  `cortex_member_consul_heartbeats_total`, `cortex_member_ring_tokens_owned`,
  `cortex_member_ring_tokens_to_own`, `cortex_ring_member_ownership_percent`,
  `cortex_ring_members`, `cortex_ring_oldest_member_timestamp`,
  `cortex_ring_tokens_owned`, `cortex_ring_tokens_total`. (@rlankfo)

- Traces: the `traces_spanmetrics_calls_total_total` metric has been renamed to
  `traces_spanmetrics_calls_total` (@fredr)

- Two new flags, `-server.http.enable-tls` and `-server.grpc.enable-tls` must
  be provided to explicitly enable TLS support. This is a change of the
  previous behavior where TLS support was enabled when a certificate pair was
  provided. (@rfratto)

- Many command line flags starting with `-server.` block have been renamed.
  (@rfratto)

- The `-log.level` and `-log.format` flags are removed in favor of being set in
  the configuration file. (@rfratto)

- Flags for configuring TLS have been removed in favor of being set in the
  configuration file. (@rfratto)

- Dynamic reload is no longer supported for deprecated server block fields.
  Changing a deprecated field will be ignored and cause the reload to fail.
  (@rfratto)

- The default HTTP listen address is now `127.0.0.1:12345`. Use the
  `-server.http.address` flag to change this value. (@rfratto)

- The default gRPC listen address is now `127.0.0.1:12346`. Use the
  `-server.grpc.address` flag to change this value. (@rfratto)

- `-reload-addr` and `-reload-port` have been removed. They are no longer
  necessary as the primary HTTP server is now static and can't be shut down in
  the middle of a `/-/reload` call. (@rfratto)

- (Only impacts `integrations-next` feature flag) Many integrations have been
  renamed to better represent what they are integrating with. For example,
  `redis_exporter` is now `redis`. This change requires updating
  `integrations-next`-enabled configuration files. This change also changes
  integration names shown in metric labels. (@rfratto)

- The deprecated `-prometheus.*` flags have been removed in favor of
  their `-metrics.*` counterparts. The `-prometheus.*` flags were first
  deprecated in v0.19.0. (@rfratto)

### Deprecations

- Most fields in the `server` block of the configuration file are
  now deprecated in favor of command line flags. These fields will be removed
  in the v0.26.0 release. Please consult the upgrade guide for more information
  and rationale. (@rfratto)

### Features

- Added config read API support to GrafanaAgent Custom Resource Definition.
  (@shamsalmon)

- Added consulagent_sd to target discovery. (@chuckyz)

- Introduce EXPERIMENTAL support for dynamic configuration. (@mattdurham)

- Introduced endpoint that accepts remote_write requests and pushes metrics data directly into an instance's WAL. (@tpaschalis)

- Added builds for linux/ppc64le. (@aklyachkin)

### Enhancements

- Tracing: Exporters can now be configured to use OAuth. (@canuteson)

- Strengthen readiness check for metrics instances. (@tpaschalis)

- Parameterize namespace field in sample K8s logs manifests (@hjet)

- Upgrade to Loki k87. (@rlankfo)

- Update Prometheus dependency to v2.34.0. (@rfratto)

- Update OpenTelemetry-collector dependency to v0.46.0. (@mapno)

- Update cAdvisor dependency to v0.44.0. (@rfratto)

- Update mongodb_exporter dependency to v0.31.2 (@mukerjee)

- Use grafana-agent/v2 Tanka Jsonnet to generate K8s manifests (@hjet)

- Replace agent-bare.yaml K8s sample Deployment with StatefulSet (@hjet)

- Improve error message for `agentctl` when timeout happens calling
  `cloud-config` command (@marctc)

- Enable integrations-next by default in agent-bare.yaml. Please note #1262 (@hjet)

### Bugfixes

- Fix Kubernetes manifests to use port `4317` for OTLP instead of the previous
  `55680` in line with the default exposed port in the agent.

- Ensure singleton integrations are honored in v2 integrations (@mattdurham)

- Tracing: `const_labels` is now correctly parsed in the remote write exporter.
  (@fredr)

- integrations-next: Fix race condition where metrics endpoints for
  integrations may disappear after reloading the config file. (@rfratto)

- Removed the `server.path_prefix` field which would break various features in
  Grafana Agent when set. (@rfratto)

- Fix issue where installing the DEB/RPM packages would overwrite the existing
  config files and environment files. (@rfratto)

- Set `grafanaDashboardFolder` as top level key in the mixin. (@Duologic)

- Operator: Custom Secrets or ConfigMaps to mount will no longer collide with
  the path name of the default secret mount. As a side effect of this bugfix,
  custom Secrets will now be mounted at
  `/var/lib/grafana-agent/extra-secrets/<secret name>` and custom ConfigMaps
  will now be mounted at `/var/lib/grafana-agent/extra-configmaps/<configmap
  name>`. This is not a breaking change as it was previously impossible to
  properly provide these custom mounts. (@rfratto)

- Flags accidentally prefixed with `-metrics.service..` (two `.` in a row) have
  now been fixed to only have one `.`. (@rfratto)

- Protect concurrent writes to the WAL in the remote write exporter (@mapno)

### Other changes

- The `-metrics.wal-directory` flag and `metrics.wal_directory` config option
  will now default to `data-agent/`, the same default WAL directory as
  Prometheus Agent. (@rfratto)

v0.23.0 (2022-02-10)
--------------------

### Enhancements

- Go 1.17 is now used for all builds of the Agent. (@tpaschalis)

- integrations-next: Add `extra_labels` to add a custom set of labels to
  integration targets. (@rfratto)

- The agent no longer appends duplicate exemplars. (@tpaschalis)

- Added Kubernetes eventhandler integration (@hjet)

- Enables sending of exemplars over remote write by default. (@rlankfo)

### Bugfixes

- Fixed issue where Grafana Agent may panic if there is a very large WAL
  loading while old WALs are being deleted or the `/agent/api/v1/targets`
  endpoint is called. (@tpaschalis)

- Fix panic in prom_sd_processor when address is empty (@mapno)

- Operator: Add missing proxy_url field from generated remote_write configs.
  (@rfratto)

- Honor the specified log format in the traces subsystem (@mapno)

- Fix typo in node_exporter for runit_service_dir. (@mattdurham)

- Allow inlining credentials in remote_write url. (@tpaschalis)

- integrations-next: Wait for integrations to stop when starting new instances
  or shutting down (@rfratto).

- Fix issue with windows_exporter mssql collector crashing the agent.
  (@mattdurham)

- The deb and rpm files will now ensure the /var/lib/grafana-agent data
  directory is created with permissions set to 0770. (@rfratto)

- Make agent-traces.yaml Namespace a template-friendly variable (@hjet)

- Disable `machine-id` journal vol by default in sample logs manifest (@hjet)

v0.22.0 (2022-01-13)
--------------------

> This release has deprecations. Please read entries carefully and consult
> the [upgrade guide][] for specific instructions.

### Deprecations

- The node_exporter integration's `netdev_device_whitelist` field is deprecated
  in favor of `netdev_device_include`. Support for the old field name will be
  removed in a future version. (@rfratto)

- The node_exporter integration's `netdev_device_blacklist` field is deprecated
  in favor of `netdev_device_include`. Support for the old field name will be
  removed in a future version. (@rfratto)

- The node_exporter integration's `systemd_unit_whitelist` field is deprecated
  in favor of `systemd_unit_include`. Support for the old field name will be
  removed in a future version. (@rfratto)

- The node_exporter integration's `systemd_unit_blacklist` field is deprecated
  in favor of `systemd_unit_exclude`. Support for the old field name will be
  removed in a future version. (@rfratto)

- The node_exporter integration's `filesystem_ignored_mount_points` field is
  deprecated in favor of `filesystem_mount_points_exclude`. Support for the old
  field name will be removed in a future version. (@rfratto)

- The node_exporter integration's `filesystem_ignored_fs_types` field is
  deprecated in favor of `filesystem_fs_types_exclude`. Support for the old
  field name will be removed in a future version. (@rfratto)

### Features

- (beta) Enable experimental config urls for fetching remote configs.
  Currently, only HTTP/S is supported. Pass the
  `-enable-features=remote-configs` flag to turn this on. (@rlankfo)

- Added [cAdvisor](https://github.com/google/cadvisor) integration. (@rgeyer)

- Traces: Add `Agent Tracing Pipeline` dashboard and alerts (@mapno)

- Traces: Support jaeger/grpc exporter (@nicoche)

- (beta) Enable an experimental integrations subsystem revamp. Pass
  `integrations-next` to `-enable-features` to turn this on. Reading the
  documentation for the revamp is recommended; enabling it causes breaking
  config changes. (@rfratto)

### Enhancements

- Traces: Improved pod association in PromSD processor (@mapno)

- Updated OTel to v0.40.0 (@mapno)

- Remote write dashboard: show in and out sample rates (@bboreham)

- Remote write dashboard: add mean latency (@bboreham)

- Update node_exporter dependency to v1.3.1. (@rfratto)

- Cherry-pick Prometheus PR #10102 into our Prometheus dependency (@rfratto).

### Bugfixes

- Fix usage of POSTGRES_EXPORTER_DATA_SOURCE_NAME when using postgres_exporter
  integration (@f11r)

- Change ordering of the entrypoint for windows service so that it accepts
  commands immediately (@mattdurham)

- Only stop WAL cleaner when it has been started (@56quarters)

- Fix issue with unquoted install path on Windows, that could allow escalation
  or running an arbitrary executable (@mattdurham)

- Fix cAdvisor so it collects all defined metrics instead of the last
  (@pkoenig10)

- Fix panic when using 'stdout' in automatic logging (@mapno)

- Grafana Agent Operator: The /-/ready and /-/healthy endpoints will
  no longer always return 404 (@rfratto).

### Other changes

- Remove log-level flag from systemd unit file (@jpkrohling)

v0.21.2 (2021-12-08)
--------------------

### Security fixes

- This release contains a fix for
  [CVE-2021-41090](https://github.com/grafana/agent/security/advisories/GHSA-9c4x-5hgq-q3wh).

### Other changes

- This release disables the existing `/-/config` and
  `/agent/api/v1/configs/{name}` endpoints by default. Pass the
  `--config.enable-read-api` flag at the command line to opt in to these
  endpoints.

v0.21.1 (2021-11-18)
--------------------

### Bugfixes

- Fix panic when using postgres_exporter integration (@saputradharma)

- Fix panic when dnsamsq_exporter integration tried to log a warning (@rfratto)

- Statsd Integration: Adding logger instance to the statsd mapper
  instantiation. (@gaantunes)

- Statsd Integration: Fix issue where mapped metrics weren't exposed to the
  integration. (@mattdurham)

- Operator: fix bug where version was a required field (@rfratto)

- Metrics: Only run WAL cleaner when metrics are being used and a WAL is
  configured. (@rfratto)

v0.21.0 (2021-11-17)
--------------------

### Enhancements

- Update Cortex dependency to v1.10.0-92-g85c378182. (@rlankfo)

- Update Loki dependency to v2.1.0-656-g0ae0d4da1. (@rlankfo)

- Update Prometheus dependency to v2.31.0 (@rlankfo)

- Add Agent Operator Helm quickstart guide (@hjet)

- Reorg Agent Operator quickstart guides (@hjet)

### Bugfixes

- Packaging: Use correct user/group env variables in RPM %post script (@simonc6372)

- Validate logs config when using logs_instance with automatic logging processor (@mapno)

- Operator: Fix MetricsInstance Service port (@hjet)

- Operator: Create govern service per Grafana Agent (@shturman)

- Operator: Fix relabel_config directive for PodLogs resource (@hjet)

- Traces: Fix `success_logic` code in service graphs processor (@mapno)

### Other changes

- Self-scraped integrations will now use an SUO-specific value for the `instance` label. (@rfratto)

- Traces: Changed service graphs store implementation to improve CPU performance (@mapno)

v0.20.1 (2021-12-08)
--------------------

> _NOTE_: The fixes in this patch are only present in v0.20.1 and >=v0.21.2.

### Security fixes

- This release contains a fix for
  [CVE-2021-41090](https://github.com/grafana/agent/security/advisories/GHSA-9c4x-5hgq-q3wh).

### Other changes

- This release disables the existing `/-/config` and
  `/agent/api/v1/configs/{name}` endpoitns by default. Pass the
  `--config.enable-read-api` flag at the command line to opt in to these
  endpoints.

v0.20.0 (2021-10-28)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Breaking Changes

- push_config is no longer supported in trace's config (@mapno)

### Features

- Operator: The Grafana Agent Operator can now generate a Kubelet service to
  allow a ServiceMonitor to collect Kubelet and cAdvisor metrics. This requires
  passing a `--kubelet-service` flag to the Operator in `namespace/name` format
  (like `kube-system/kubelet`). (@rfratto)

- Service graphs processor (@mapno)

### Enhancements

- Updated mysqld_exporter to v0.13.0 (@gaantunes)

- Updated postgres_exporter to v0.10.0 (@gaantunes)

- Updated redis_exporter to v1.27.1 (@gaantunes)

- Updated memcached_exporter to v0.9.0 (@gaantunes)

- Updated statsd_exporter to v0.22.2 (@gaantunes)

- Updated elasticsearch_exporter to v1.2.1 (@gaantunes)

- Add remote write to silent Windows Installer  (@mattdurham)

- Updated mongodb_exporter to v0.20.7 (@rfratto)

- Updated OTel to v0.36 (@mapno)

- Updated statsd_exporter to v0.22.2 (@mattdurham)

- Update windows_exporter to v0.16.0 (@rfratto, @mattdurham)

- Add send latency to agent dashboard (@bboreham)

### Bugfixes

- Do not immediately cancel context when creating a new trace processor. This
  was preventing scrape_configs in traces from functioning. (@lheinlen)

- Sanitize autologged Loki labels by replacing invalid characters with
  underscores (@mapno)

- Traces: remove extra line feed/spaces/tabs when reading password_file content
  (@nicoche)

- Updated envsubst to v2.0.0-20210730161058-179042472c46. This version has a
  fix needed for escaping values outside of variable substitutions. (@rlankfo)

- Grafana Agent Operator should no longer delete resources matching the names
  of the resources it manages. (@rfratto)

- Grafana Agent Operator will now appropriately assign an
  `app.kubernetes.io/managed-by=grafana-agent-operator` to all created
  resources. (@rfratto)

### Other changes

- Configuration API now returns 404 instead of 400 when attempting to get or
  delete a config which does not exist. (@kgeckhart)

- The windows_exporter now disables the textfile collector by default.
  (@rfratto)

v0.19.0 (2021-09-29)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Breaking Changes

- Reduced verbosity of tracing autologging by not logging `STATUS_CODE_UNSET`
  status codes. (@mapno)

- Operator: rename `Prometheus*` CRDs to `Metrics*` and `Prometheus*` fields to
  `Metrics*`. (@rfratto)

- Operator: CRDs are no longer referenced using a hyphen in the name to be
  consistent with how Kubernetes refers to resources. (@rfratto)

- `prom_instance` in the spanmetrics config is now named `metrics_instance`.
  (@rfratto)

### Deprecations

- The `loki` key at the root of the config file has been deprecated in favor of
  `logs`. `loki`-named fields in `automatic_logging` have been renamed
  accordinly: `loki_name` is now `logs_instance_name`, `loki_tag` is now
  `logs_instance_tag`, and `backend: loki` is now `backend: logs_instance`.
  (@rfratto)

- The `prometheus` key at the root of the config file has been deprecated in
  favor of `metrics`. Flag names starting with `prometheus.` have also been
  deprecated in favor of the same flags with the `metrics.` prefix. Metrics
  prefixed with `agent_prometheus_` are now prefixed with `agent_metrics_`.
  (@rfratto)

- The `tempo` key at the root of the config file has been deprecated in favor
  of `traces`. (@mattdurham)

### Features

- Added [GitHub exporter](https://github.com/infinityworks/github-exporter)
  integration. (@rgeyer)

- Add TLS config options for tempo `remote_write`s. (@mapno)

- Support autologging span attributes as log labels (@mapno)

- Put Tests requiring Network Access behind a -online flag (@flokli)

- Add logging support to the Grafana Agent Operator. (@rfratto)

- Add `operator-detach` command to agentctl to allow zero-downtime upgrades
  when removing an Operator CRD. (@rfratto)

- The Grafana Agent Operator will now default to deploying the matching release
  version of the Grafana Agent instead of v0.14.0. (@rfratto)

### Enhancements

- Update OTel dependency to v0.30.0 (@mapno)

- Allow reloading configuration using `SIGHUP` signal. (@tharun208)

- Add HOSTNAME environment variable to service file to allow for expanding the
  $HOSTNAME variable in agent config.  (@dfrankel33)

- Update jsonnet-libs to 1.21 for Kubernetes 1.21+ compatability. (@MurzNN)

- Make method used to add k/v to spans in prom_sd processor configurable.
  (@mapno)

### Bugfixes

- Regex capture groups like `${1}` will now be kept intact when using
  `-config.expand-env`. (@rfratto)

- The directory of the logs positions file will now properly be created on
  startup for all instances. (@rfratto)

- The Linux system packages will now configure the grafana-agent user to be a
  member of the adm and systemd-journal groups. This will allow logs to read
  from journald and /var/log by default. (@rfratto)

- Fix collecting filesystem metrics on Mac OS (darwin) in the `node_exporter`
  integration default config. (@eamonryan)

- Remove v0.0.0 flags during build with no explicit release tag (@mattdurham)

- Fix issue with global scrape_interval changes not reloading integrations
  (@kgeckhart)

- Grafana Agent Operator will now detect changes to referenced ConfigMaps and
  Secrets and reload the Agent properly. (@rfratto)

- Grafana Agent Operator's object label selectors will now use Kubernetes
  defaults when undefined (i.e., default to nothing). (@rfratto)

- Fix yaml marshalling tag for cert_file in kafka exporter agent config.
  (@rgeyer)

- Fix warn-level logging of dropped targets. (@james-callahan)

- Standardize scrape_interval to 1m in examples. (@mattdurham)

v0.18.4 (2021-09-14)
--------------------

### Enhancements

- Add `agent_prometheus_configs_changed_total` metric to track instance config
  events. (@rfratto)

### Bugfixes

- Fix info logging on windows. (@mattdurham)

- Scraping service: Ensure that a reshard is scheduled every reshard
  interval. (@rfratto)

v0.18.3 (2021-09-08)
--------------------

### Bugfixes

- Register missing metric for configstore consul request duration. (@rfratto)

- Logs should contain a caller field with file and line numbers again
  (@kgeckhart)

- In scraping service mode, the polling configuration refresh should honor
  timeout. (@mattdurham)

- In scraping service mode, the lifecycle reshard should happen using a
  goroutine. (@mattdurham)

- In scraping service mode, scraping service can deadlock when reloading during
  join. (@mattdurham)

- Scraping service: prevent more than one refresh from being queued at a time.
  (@rfratto)

v0.18.2 (2021-08-12)
--------------------

### Bugfixes

- Honor the prefix and remove prefix from consul list results (@mattdurham)

v0.18.1 (2021-08-09)
--------------------

### Bugfixes

- Reduce number of consul calls when ran in scrape service mode (@mattdurham)

v0.18.0 (2021-07-29)
--------------------

### Features

- Added [GitHub exporter](https://github.com/infinityworks/github-exporter)
  integration. (@rgeyer)

- Add support for OTLP HTTP trace exporting. (@mapno)

### Enhancements

- Switch to drone for releases. (@mattdurham)

- Update postgres_exporter to a [branch of](https://github.com/grafana/postgres_exporter/tree/exporter-package-v0.10.0) v0.10.0

### Bugfixes

- Enabled flag for integrations is not being honored. (@mattdurham)

v0.17.0 (2021-07-15)
--------------------

### Features

- Added [Kafka Lag exporter](https://github.com/davidmparrott/kafka_exporter)
  integration. (@gaantunes)

### Bugfixes

- Fix race condition that may occur and result in a panic when initializing
  scraping service cluster. (@rfratto)

v0.16.1 (2021-06-22)
--------------------

### Bugfixes

- Fix issue where replaying a WAL caused incorrect metrics to be sent over
  remote write. (@rfratto)

v0.16.0 (2021-06-17)
--------------------

### Features

- (beta) A Grafana Agent Operator is now available. (@rfratto)

### Enhancements

- Error messages when installing the Grafana Agent for Grafana Cloud will now
  be shown. (@rfratto)

### Bugfixes

- Fix a leak in the shared string interner introduced in v0.14.0. This fix was
  made to a [dependency](https://github.com/grafana/prometheus/pull/21).
  (@rfratto)

- Fix issue where a target will fail to be scraped for the process lifetime if
  that target had gone down for long enough that its series were removed from
  the in-memory cache (2 GC cycles). (@rfratto)

v0.15.0 (2021-06-03)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Breaking Changes

- The configuration of Tempo Autologging has changed. (@mapno)

### Features

- Add support for exemplars. (@mapno)

### Enhancements

- Add the option to log to stdout instead of a Loki instance. (@joe-elliott)

- Update Cortex dependency to v1.8.0.

- Running the Agent as a DaemonSet with host_filter and role: pod should no
  longer cause unnecessary load against the Kubernetes SD API. (@rfratto)

- Update Prometheus to v2.27.0. (@mapno)

- Update Loki dependency to d88f3996eaa2. This is a non-release build, and was
  needed to support exemplars. (@mapno)

- Update Cortex dependency to d382e1d80eaf. This is a non-release build, and
  was needed to support exemplars. (@mapno)

### Bugfixes

- Host filter relabeling rules should now work. (@rfratto)

- Fixed issue where span metrics where being reported with wrong time unit.
  (@mapno)

### Other changes

- Intentionally order tracing processors. (@joe-elliott)

v0.14.0 (2021-05-24)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.
>
> **STABILITY NOTICE**: As of this release, functionality that is not
> recommended for production use and is expected to change will be tagged
> interchangably as "experimental" or "beta."

### Security fixes

- The Scraping service API will now reject configs that read credentials from
  disk by default. This prevents malicious users from reading arbitrary files
  and sending their contents over the network. The old behavior can be
  re-enabled by setting `dangerous_allow_reading_files: true` in the scraping
  service config. (@rfratto)

### Breaking changes

- Configuration for SigV4 has changed. (@rfratto)

### Deprecations

- `push_config` is now supplanted by `remote_block` and `batch`. `push_config`
  will be removed in a future version (@mapno)

### Features

- (beta) New integration: windows_exporter (@mattdurham)

- (beta) Grafana Agent Windows Installer is now included as a release artifact.
  (@mattdurham)

- Official M1 Mac release builds will now be generated! Look for
  `agent-darwin-arm64` and `agentctl-darwin-arm64` in the release assets.
  (@rfratto)

- Add support for running as a Windows service (@mattdurham)

- (beta) Add /-/reload support. It is not recommended to invoke `/-/reload`
  against the main HTTP server. Instead, two new command-line flags have been
  added: `--reload-addr` and `--reload-port`. These will launch a
  `/-/reload`-only HTTP server that can be used to safely reload the Agent's
  state.  (@rfratto)

- Add a /-/config endpoint. This endpoint will return the current configuration
  file with defaults applied that the Agent has loaded from disk. (@rfratto)

- (beta) Support generating metrics and exposing them via a Prometheus exporter
  from span data. (@yeya24)

- Tail-based sampling for tracing pipelines (@mapno)

- Added Automatic Logging feature for Tempo (@joe-elliott)

- Disallow reading files from within scraping service configs by default.
  (@rfratto)

- Add remote write for span metrics (@mapno)

### Enhancements

- Support compression for trace export. (@mdisibio)

- Add global remote_write configuration that is shared between all instances
  and integrations. (@mattdurham)

- Go 1.16 is now used for all builds of the Agent. (@rfratto)

- Update Prometheus dependency to v2.26.0. (@rfratto)

- Upgrade `go.opentelemetry.io/collector` to v0.21.0 (@mapno)

- Add kafka trace receiver (@mapno)

- Support mirroring a trace pipeline to multiple backends (@mapno)

- Add `headers` field in `remote_write` config for Tempo. `headers` specifies
  HTTP headers to forward to the remote endpoint. (@alexbiehl)

- Add silent uninstall to Windows Uninstaller. (@mattdurham)

### Bugfixes

- Native Darwin arm64 builds will no longer crash when writing metrics to the
  WAL. (@rfratto)

- Remote write endpoints that never function across the lifetime of the Agent
  will no longer prevent the WAL from being truncated. (@rfratto)

- Bring back FreeBSD support. (@rfratto)

- agentctl will no longer leak WAL resources when retrieving WAL stats.
  (@rfratto)

- Ensure defaults are applied to undefined sections in config file. This fixes
  a problem where integrations didn't work if `prometheus:` wasn't configured.
  (@rfratto)

- Fixed issue where automatic logging double logged "svc". (@joe-elliott)

### Other changes

- The Grafana Cloud Agent has been renamed to the Grafana Agent. (@rfratto)

- Instance configs uploaded to the Config Store API will no longer be stored
  along with the global Prometheus defaults. This is done to allow globals to
  be updated and re-apply the new global defaults to the configs from the
  Config Store. (@rfratto)

- The User-Agent header sent for logs will now be `GrafanaAgent/<version>`
  (@rfratto)

- Add `tempo_spanmetrics` namespace in spanmetrics (@mapno)

v0.13.1 (2021-04-09)
--------------------

### Bugfixes

- Validate that incoming scraped metrics do not have an empty label set or a
  label set with duplicate labels, mirroring the behavior of Prometheus.
  (@rfratto)

v0.13.0 (2021-02-25)
--------------------

> The primary branch name has changed from `master` to `main`. You may have to
> update your local checkouts of the repository to point at the new branch name.

### Features

- postgres_exporter: Support query_path and disable_default_metrics. (@rfratto)

### Enhancements

- Support other architectures in installation script. (@rfratto)

- Allow specifying custom wal_truncate_frequency per integration. (@rfratto)

- The SigV4 region can now be inferred using the shared config (at
  `$HOME/.aws/config`) or environment variables (via `AWS_CONFIG`). (@rfratto)

- Update Prometheus dependency to v2.25.0. (@rfratto)

### Bugfixes

- Not providing an `-addr` flag for `agentctl config-sync` will no longer
  report an error and will instead use the pre-existing default value.
  (@rfratto)

- Fixed a bug from v0.12.0 where the Loki installation script failed because
  positions_directory was not set. (@rfratto)

- Reduce the likelihood of dataloss during a remote_write-side outage by
  increasing the default wal_truncation_frequency to 60m and preventing the WAL
  from being truncated if the last truncation timestamp hasn't changed. This
  change increases the size of the WAL on average, and users may configure a
  lower wal_truncation_frequency to deliberately choose a smaller WAL over
  write guarantees. (@rfratto)

- Add the ability to read and serve HTTPS integration metrics when given a set
  certificates (@mattdurham)

v0.12.0 (2021-02-05)
--------------------

> **BREAKING CHANGES**: This release has breaking changes. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Breaking Changes

- The configuration format for the `loki` block has changed. (@rfratto)

- The configuration format for the `tempo` block has changed. (@rfratto)

### Features

- Support for multiple Loki Promtail instances has been added. (@rfratto)

- Support for multiple Tempo instances has been added. (@rfratto)

- Added [ElasticSearch exporter](https://github.com/justwatchcom/elasticsearch_exporter)
  integration. (@colega)

### Enhancements

- `.deb` and `.rpm` packages are now generated for all supported architectures.
  The architecture of the AMD64 package in the filename has been renamed to
  `amd64` to stay synchronized with the architecture name presented from other
  release assets. (@rfratto)

- The `/agent/api/v1/targets` API will now include discovered labels on the
  target pre-relabeling in a `discovered_labels` field. (@rfratto)

- Update Loki to 59a34f9867ce. This is a non-release build, and was needed to
  support multiple Loki instances. (@rfratto)

- Scraping service: Unhealthy Agents in the ring will no longer cause job
  distribution to fail. (@rfratto)

- Scraping service: Cortex ring metrics (prefixed with cortex_ring_) will now
  be registered for tracking the state of the hash ring. (@rfratto)

- Scraping service: instance config ownership is now determined by the hash of
  the instance config name instead of the entire config. This means that
  updating a config is guaranteed to always hash to the same Agent, reducing
  the number of metrics gaps. (@rfratto)

- Only keep a handful of K8s API server metrics by default to reduce default
  active series usage. (@hjet)

- Go 1.15.8 is now used for all distributions of the Agent. (@rfratto)

### Bugfixes

- `agentctl config-check` will now work correctly when the supplied config file
  contains integrations. (@hoenn)

v0.11.0 (2021-01-20)
--------------------

### Features

- ARMv6 builds of `agent` and `agentctl` will now be included in releases to
  expand Agent support to cover all models of Raspberry Pis. ARMv6 docker
  builds are also now available. (@rfratto)

- Added `config-check` subcommand for `agentctl` that can be used to validate
  Agent configuration files before attempting to load them in the `agent`
  itself. (@56quarters)

### Enhancements

- A sigv4 install script for Prometheus has been added. (@rfratto)

- NAMESPACE may be passed as an environment variable to the Kubernetes install
  scripts to specify an installation namespace. (@rfratto)

### Bugfixes

- The K8s API server scrape job will use the API server Service name when
  resolving IP addresses for Prometheus service discovery using the "Endpoints"
  role. (@hjet)

- The K8s manifests will no longer include the `default/kubernetes` job twice
  in both the DaemonSet and the Deployment. (@rfratto)

v0.10.0 (2021-01-13)
--------------------

### Features

- Prometheus `remote_write` now supports SigV4 authentication using the
  [AWS default credentials chain](https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html).
  This enables the Agent to send metrics to Amazon Managed Prometheus without
  needing the [SigV4 Proxy](https://github.com/awslabs/aws-sigv4-proxy).
  (@rfratto)

### Enhancements

- Update `redis_exporter` to v1.15.0. (@rfratto)

- `memcached_exporter` has been updated to v0.8.0. (@rfratto)

- `process-exporter` has been updated to v0.7.5. (@rfratto)

- `wal_cleanup_age` and `wal_cleanup_period` have been added to the top-level
  Prometheus configuration section. These settings control how Write Ahead Logs
  (WALs) that are not associated with any instances are cleaned up. By default,
  WALs not associated with an instance that have not been written in the last
  12 hours are eligible to be cleaned up. This cleanup can be disabled by
  setting `wal_cleanup_period` to `0`. (@56quarters)

- Configuring logs to read from the systemd journal should now work on journals
  that use +ZSTD compression. (@rfratto)

### Bugfixes

- Integrations will now function if the HTTP listen address was set to a value
  other than the default. (@mattdurham)

- The default Loki installation will now be able to write its positions file.
  This was prevented by accidentally writing to a readonly volume mount.
  (@rfratto)

v0.9.1 (2021-01-04)
-------------------

### Enhancements

- agentctl will now be installed by the rpm and deb packages as
  `grafana-agentctl`. (@rfratto)

v0.9.0 (2020-12-10)
-------------------

### Features

- Add support to configure TLS config for the Tempo exporter to use
  insecure_skip_verify to disable TLS chain verification. (@bombsimon)

- Add `sample-stats` to `agentctl` to search the WAL and return a summary of
  samples of series matching the given label selector. (@simonswine)

- New integration:
  [postgres_exporter](https://github.com/wrouesnel/postgres_exporter)
  (@rfratto)

- New integration:
  [statsd_exporter](https://github.com/prometheus/statsd_exporter) (@rfratto)

- New integration:
  [consul_exporter](https://github.com/prometheus/consul_exporter) (@rfratto)

- Add optional environment variable substitution of configuration file.
  (@dcseifert)

### Enhancements

- `min_wal_time` and `max_wal_time` have been added to the instance config
  settings, guaranteeing that data in the WAL will exist for at least
  `min_wal_time` and will not exist for longer than `max_wal_time`. This change
  will increase the size of the WAL slightly but will prevent certain scenarios
  where data is deleted before it is sent. To revert back to the old behavior,
  set `min_wal_time` to `0s`. (@rfratto)

- Update `redis_exporter` to v1.13.1. (@rfratto)

- Bump OpenTelemetry-collector dependency to v0.16.0. (@bombsimon)

### Bugfixes

- Fix issue where the Tempo example manifest could not be applied because the
  port names were too long. (@rfratto)

- Fix issue where the Agent Kubernetes manifests may not load properly on AKS.
  (#279) (@rfratto)

### Other changes

- The User-Agent header sent for logs will now be `GrafanaCloudAgent/<version>`
  (@rfratto)

v0.8.0 (2020-11-06)
-------------------

### Features

- New integration: [dnsamsq_exporter](https://github.com/google/dnsamsq_exporter)
  (@rfratto).

- New integration: [memcached_exporter](https://github.com/prometheus/memcached_exporter)
  (@rfratto).

### Enhancements

- Add `<integration name>_build_info` metric to all integrations. The build
  info displayed will match the build information of the Agent and _not_ the
  embedded exporter. This metric is used by community dashboards, so adding it
  to the Agent increases compatibility with existing dashboards that depend on
  it existing. (@rfratto)

- Bump OpenTelemetry-collector dependency to 0.14.0 (@joe-elliott)

### Bugfixes

- Error messages when retrieving configs from the KV store will now be logged,
  rather than just logging a generic message saying that retrieving the config
  has failed. (@rfratto)

v0.7.2 (2020-10-29)
-------------------

### Enhancements

- Bump Prometheus dependency to 2.21. (@rfratto)

- Bump OpenTelemetry-collector dependency to 0.13.0 (@rfratto)

- Bump Promtail dependency to 2.0. (@rfratto)

- Enhance host_filtering mode to support targets from Docker Swarm and Consul.
  Also, add a `host_filter_relabel_configs` to that will apply relabeling rules
  for determining if a target should be dropped. Add a documentation section
  explaining all of this in detail. (@rfratto)

### Bugfixes

- Fix deb package prerm script so that it stops the agent on package removal.
  (@jdbaldry)

- Fix issue where the `push_config` for Tempo field was expected to be
  `remote_write`. `push_config` now works as expected. (@rfratto)

v0.7.1 (2020-10-23)
-------------------

### Bugfixes

- Fix issue where ARM binaries were not published with the GitHub release.

v0.7.0 (2020-10-23)
-------------------

### Features

- Added Tracing Support. (@joe-elliott)

- Add RPM and deb packaging. (@jdbaldry, @simon6372)

- arm64 and arm/v7 Docker containers and release builds are now available for
  `agent` and `agentctl`. (@rfratto)

- Add `wal-stats` and `target-stats` tooling to `agentctl` to discover WAL and
  cardinality issues. (@rfratto)

- [mysqld_exporter](https://github.com/prometheus/mysqld_exporter) is now
  embedded and available as an integration. (@rfratto)

- [redis_exporter](https://github.com/oliver006/redis_exporter) is now embedded
  and available as an integration. (@dafydd-t)

### Enhancements

- Resharding the cluster when using the scraping service mode now supports
  timeouts through `reshard_timeout`. The default value is `30s.` This timeout
  applies to cluster-wide reshards (performed when joining and leaving the
  cluster) and local reshards (done on the `reshard_interval`). (@rfratto)

### Bugfixes

- Fix issue where integrations crashed with instance_mode was set to `distinct`
  (@rfratto)

- Fix issue where the `agent` integration did not work on Windows (@rfratto).

- Support URL-encoded paths in the scraping service API. (@rfratto)

- The instance label written from replace_instance_label can now be overwritten
  with relabel_configs. This bugfix slightly modifies the behavior of what data
  is stored. The final instance label will now be stored in the WAL rather than
  computed by remote_write. This change should not negatively affect existing
  users. (@rfratto)

v0.6.1 (2020-04-11)
-------------------

### Bugfixes

- Fix issue where build information was empty when running the Agent with
  --version. (@rfratto)

- Fix issue where updating a config in the scraping service may fail to pick up
  new targets. (@rfratto)

- Fix deadlock that slowly prevents the Agent from scraping targets at a high
  scrape volume. (@rfratto)

v0.6.0 (2020-09-04)
-------------------

### Breaking Changes

- The Configs API will now disallow two instance configs having multiple
  `scrape_configs` with the same `job_name`. This was needed for the instance
  sharing mode, where combined instances may have duplicate `job_names` across
  their `scrape_configs`. This brings the scraping service more in line with
  Prometheus, where `job_names` must globally be unique. This change also
  disallows concurrent requests to the put/apply config API endpoint to prevent
  a race condition of two conflicting configs being applied at the same time.
  (@rfratto)

### Deprecations

- `use_hostname_label` is now supplanted by `replace_instance_label`.
  `use_hostname_label` will be removed in a future version. (@rfratto)

### Features

- The Grafana Agent can now collect logs and send to Loki. This is done by
  embedding Promtail, the official Loki log collection client. (@rfratto)

- Integrations can now be enabled without scraping. Set scrape_integrations to
  `false` at the `integrations` key or within the specific integration you
  don't want to scrape. This is useful when another Agent or Prometheus server
  will scrape the integration. (@rfratto)

- [process-exporter](https://github.com/ncabatoff/process-exporter) is now
  embedded as `process_exporter`. The hypen has been changed to an underscore
  in the config file to retain consistency with `node_exporter`. (@rfratto)

### Enhancements

- A new config option, `replace_instance_label`, is now available for use with
  integrations. When this is true, the instance label for all metrics coming
  from an integration will be replaced with the machine's hostname rather than
  127.0.0.1. (@rfratto)

- The embedded Prometheus version has been updated to 2.20.1. (@rfratto,
  @gotjosh)

- The User-Agent header written by the Agent when remote_writing will now be
  `GrafanaCloudAgent/<Version>` instead of `Prometheus/<Prometheus Version>`.
  (@rfratto)

- The subsystems of the Agent (`prometheus`, `loki`) are now made optional.
  Enabling integrations also implicitly enables the associated subsystem. For
  example, enabling the `agent` or `node_exporter` integration will force the
  `prometheus` subsystem to be enabled.  (@rfratto)

### Bugfixes

- The documentation for Tanka configs is now correct. (@amckinley)

- Minor corrections and spelling issues have been fixed in the Overview
  documentation. (@amckinley)

- The new default of `shared` instances mode broke the metric value for
  `agent_prometheus_active_configs`, which was tracking the number of combined
  configs (i.e., number of launched instances). This metric has been fixed and
  a new metric, `agent_prometheus_active_instances`, has been added to track
  the numbger of launched instances. If instance sharing is not enabled, both
  metrics will share the same value. (@rfratto)

- `remote_write` names in a group will no longer be copied from the
  remote_write names of the first instance in the group. Rather, all
  remote_write names will be generated based on the first 6 characters of the
  group hash and the first six characters of the remote_write hash. (@rfratto)

- Fix a panic that may occur during shutdown if the WAL is closed in the middle
  of the WAL being truncated. (@rfratto)

v0.5.0 (2020-08-12)
-------------------

### Features

- A [scrape targets API](https://github.com/grafana/agent/blob/main/docs/api.md#list-current-scrape-targets)
  has been added to show every target the Agent is currently scraping, when it
  was last scraped, how long it took to scrape, and errors from the last
  scrape, if any. (@rfratto)

- "Shared Instance Mode" is the new default mode for spawning Prometheus
  instances, and will improve CPU and memory usage for users of integrations
  and the scraping service. (@rfratto)

### Enhancements

- Memory stability and utilization of the WAL has been improved, and the
  reported number of active series in the WAL will stop double-counting
  recently churned series. (@rfratto)

- Changing scrape_configs and remote_write configs for an instance will now be
  dynamically applied without restarting the instance. This will result in less
  missing metrics for users of the scraping service that change a config.
  (@rfratto)

- The Tanka configuration now uses k8s-alpha. (@duologic)

### Bugfixes

- The Tanka configuration will now also deploy a single-replica deployment
  specifically for scraping the Kubernetes API. This deployment acts together
  with the Daemonset to scrape the full cluster and the control plane.
  (@gotjosh)

- The node_exporter filesystem collector will now work on Linux systems without
  needing to manually set the blocklist and allowlist of filesystems.
  (@rfratto)

v0.4.0 (2020-06-18)
-------------------

### Features

- Support for integrations has been added. Integrations can be any embedded
  tool, but are currently used for embedding exporters and generating scrape
  configs. (@rfratto)

- node_exporter has been added as an integration. This is the full version of
  node_exporter with the same configuration options. (@rfratto)

- An Agent integration that makes the Agent automatically scrape itself has
  been added. (@rfratto)

### Enhancements

- The WAL can now be truncated if running the Agent without any remote_write
  endpoints. (@rfratto)

### Bugfixes

- Prevent the Agent from crashing when a global Prometheus config stanza is not
  provided. (@robx)

- Enable agent host_filter in the Tanka configs, which was disabled by default
  by mistake. (@rfratto)

v0.3.2 (2020-05-29)
-------------------

### Features

- Tanka configs that deploy the scraping service mode are now available
  (@rfratto)

- A k3d example has been added as a counterpart to the docker-compose example.
  (@rfratto)

### Enhancements

- Labels provided by the default deployment of the Agent (Kubernetes and Tanka)
  have been changed to align with the latest changes to grafana/jsonnet-libs.
  The old `instance` label is now called `pod`, and the new `instance` label is
  unique. A `container` label has also been added. The Agent mixin has been
  subsequently updated to also incorporate these label changes. (@rfratto)

- The `remote_write` and `scrape_config` sections now share the same
  validations as Prometheus (@rfratto)

- Setting `wal_truncation_frequency` to less than the scrape interval is now
  disallowed (@rfratto)

### Bugfixes

- A deadlock in scraping service mode when updating a config that shards to the
  same node has been fixed (@rfratto)

- `remote_write` config stanzas will no longer ignore `password_file`
  (@rfratto)

- `scrape_config` client secrets (e.g., basic auth, bearer token,
  `password_file`) will now be properly retained in scraping service mode
  (@rfratto)

- Labels for CPU, RX, and TX graphs in the Agent Operational dashboard now
  correctly show the pod name of the Agent instead of the exporter name.
  (@rfratto)

v0.3.1 (2020-05-20)
-------------------

### Features

- The Agent has upgraded its vendored Prometheus to v2.18.1 (@gotjosh,
  @rfratto)

### Bugfixes

- A typo in the Tanka configs and Kubernetes manifests that prevents the Agent
  launching with v0.3.0 has been fixed (@captncraig)

- Fixed a bug where Tanka mixins could not be used due to an issue with the
  folder placement enhancement (@rfratto)

### Enhancements

- `agentctl` and the config API will now validate that the YAML they receive
  are valid instance configs. (@rfratto)

v0.3.0 (2020-05-13)
-------------------

### Features

- A third operational mode called "scraping service mode" has been added. A KV
  store is used to store instance configs which are distributed amongst a
  clustered set of Agent processes, dividing the total scrape load across each
  agent. An API is exposed on the Agents to list, create, update, and delete
  instance configurations from the KV store. (@rfratto)

- An "agentctl" binary has been released to interact with the new instance
  config management API created by the "scraping service mode." (@rfratto,
  @hoenn)

- The Agent now includes readiness and healthiness endpoints. (@rfratto)

### Enhancements

- The YAML files are now parsed strictly and an invalid YAML will generate an
  error at runtime. (@hoenn)

- The default build mode for the Docker containers is now release, not debug.
  (@rfratto)

- The Grafana Agent Tanka Mixins now are placed in an "Agent" folder within
  Grafana. (@cyriltovena)

v0.2.0 (2020-04-09)
-------------------

### Features

- The Prometheus remote write protocol will now send scraped metadata (metric
  name, help, type and unit). This results in almost negligent bytes sent
  increase as metadata is only sent every minute. It is on by default.
  (@gotjosh)

  These metrics are available to monitor metadata being sent:
  - `prometheus_remote_storage_succeeded_metadata_total`
  - `prometheus_remote_storage_failed_metadata_total`
  - `prometheus_remote_storage_retried_metadata_total`
  - `prometheus_remote_storage_sent_batch_duration_seconds` and
    `prometheus_remote_storage_sent_bytes_total` have a new label “type” with
    the values of `metadata` or `samples`.

### Enhancements

- The Agent has upgraded its vendored Prometheus to v2.17.1 (@rfratto)

### Bugfixes

- Invalid configs passed to the agent will now stop the process after they are
  logged as invalid; previously the Agent process would continue. (@rfratto)

- Enabling host_filter will now allow metrics from node role Kubernetes service
  discovery to be scraped properly (e.g., cAdvisor, Kubelet). (@rfratto)

v0.1.1 (2020-03-16)
-------------------

### Other changes

- Nits in documentation (@sh0rez)

- Fix various dashboard mixin problems from v0.1.0 (@rfratto)

- Pass through release tag to `docker build` (@rfratto)

v0.1.0 (2020-03-16)
-------------------

> First release!

### Features

- Support for scraping Prometheus metrics and sharding the agent through the
  presence of a `host_filter` flag within the Agent configuration file.

[upgrade guide]: https://grafana.com/docs/agent/latest/upgrade-guide/
[contributors guide]: ./docs/developer/contributing.md#updating-the-changelog
