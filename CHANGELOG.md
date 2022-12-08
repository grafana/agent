# Changelog

> _Contributors should read our [contributors guide][] for instructions on how
> to update the changelog._

This document contains a historical list of changes between releases. Only
changes that impact end-user behavior are listed; changes to documentation or
internal API changes are not present.

Main (unreleased)
-----------------

> **DEPRECATIONS**: This release has deprecations. Please read entries
> carefully and consult the [upgrade guide][] for specific instructions.

### Deprecations

- The `EXPERIMENTAL_ENABLE_FLOW` environment variable is deprecated in favor of
  `AGENT_MODE=flow`. Support for `EXPERIMENTAL_ENABLE_FLOW` will be removed in
  v0.32. (@rfratto)

- The `ebpf_exporter` integration has been removed due to issues with static
  linking. It may be brought back once these are resolved. (@tpaschalis)

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

### Enhancements

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
  native histograms from the WAL to the specified endpoints.

### Bugfixes

- Fix issue where whitespace was being sent as part of password when using a
  password file for `redis_exporter`. (@spartan0x117)

- Flow UI: Fix issue where a configuration block referencing a component would
  cause the graph page to fail to load. (@rfratto)

- Remove duplicate `oauth2` key from `metricsinstances` CRD. (@daper)

- Fix issue where on checking whether to restart integrations the Integration Manager was comparing
  configs with secret values scrubbed, preventing reloads if only secrets were updated. (@spartan0x117)

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

- `extra-scrape-metrics` can now be enabled with the `--enable-features=extra-scrape-metrics` feature flag. See https://prometheus.io/docs/prometheus/2.31/feature_flags/#extra-scrape-metrics for details. (@rlankfo)

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

> *NOTE*: The fixes in this patch are only present in v0.20.1 and >=v0.21.2.

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

- Operator: rename Prometheus* CRDs to Metrics* and Prometheus* fields to
  Metrics*. (@rfratto)

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

- Added [Github exporter](https://github.com/infinityworks/github-exporter)
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

- Added [Github exporter](https://github.com/infinityworks/github-exporter)
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

- Update Cortex dependency to to d382e1d80eaf. This is a non-release build, and
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

* The Scraping service API will now reject configs that read credentials from
  disk by default. This prevents malicious users from reading arbitrary files
  and sending their contents over the network. The old behavior can be
  re-enabled by setting `dangerous_allow_reading_files: true` in the scraping
  service config. (@rfratto)

### Breaking changes

* Configuration for SigV4 has changed. (@rfratto)

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

* The configuration format for the `loki` block has changed. (@rfratto)

* The configuration format for the `tempo` block has changed. (@rfratto)

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
  info displayed will match the build information of the Agent and *not* the
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
  computed by remote_write. This change should not negatively effect existing
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

* Support for scraping Prometheus metrics and sharding the agent through the
  presence of a `host_filter` flag within the Agent configuration file.

[upgrade guide]: https://grafana.com/docs/agent/latest/upgrade-guide/
[contributors guide]: ./docs/developer/contributing.md#updating-the-changelog
