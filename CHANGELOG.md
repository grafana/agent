# Next (master/unreleased)

NOTE: FreeBSD builds will not be included for this release. There is a bug in an
upstream library preventing cross-compilation of the Grafana Cloud Agent for
this platform. FreeBSD builds will return in a future release.

- [FEATURE] The Grafana Cloud Agent can now collect logs and send to Loki. This
  is done by embedding Promtail, the official Loki log collection client.
  (@rfratto)

- [FEATURE] Integrations can now be enabled without scraping. Set
  scrape_integrations to `false` at the `integrations` key or within the
  specific integration you don't want to scrape. This is useful when another
  Agent or Prometheus server will scrape the integration. (@rfratto)

- [FEATURE] [process-exporter](https://github.com/ncabatoff/process-exporter) is
  now embedded as `process_exporter`. The hypen has been changed to an
  underscore in the config file to retain consistency with `node_exporter`.
  (@rfratto)

- [ENHANCEMENT] A new config option, `replace_instance_label`, is now available
  for use with integrations. When this is true, the instance label for all
  metrics coming from an integration will be replaced with the machine's
  hostname rather than 127.0.0.1. (@rfratto)

- [EHANCEMENT] The embedded Prometheus version has been updated to 2.20.1.
  (@rfratto, @gotjosh)

- [ENHANCEMENT] The User-Agent header written by the Agent when remote_writing
  will now be `GrafanaCloudAgent/<Version>` instead of `Prometheus/<Prometheus Version>`.
  (@rfratto)

- [ENHANCEMENT] The subsystems of the Agent (`prometheus`, `loki`) are now made
  optional. Enabling integrations also implicitly enables the associated
  subsystem. For example, enabling the `agent` or `node_exporter` integration will
  force the `prometheus` subsystem to be enabled.  (@rfratto)

- [BUGFIX] The documentation for Tanka configs is now correct. (@amckinley)

- [BUGFIX] Minor corrections and spelling issues have been fixed in the Overview
  documentation. (@amckinley)

- [BUGFIX] The new default of `shared` instances mode broke the metric value for
  `agent_prometheus_active_configs`, which was tracking the number of combined
  configs (i.e., number of launched instances). This metric has been fixed and
  a new metric, `agent_prometheus_active_instances`, has been added to track
  the numbger of launched instances. If instance sharing is not enabled, both
  metrics will share the same value. (@rfratto)

- [BUGFIX] The Configs API will now disallow two instance configs having
  multiple `scrape_configs` with the same `job_name`. THIS IS A BREAKING CHANGE.
  This was needed for the instance sharing mode, where combined instances may
  have duplicate `job_names` across their `scrape_configs`. This brings the
  scraping service more in line with Prometheus, where `job_names` must globally
  be unique. This change also disallows concurrent requests to the put/apply
  config API endpoint to prevent a race condition of two conflicting configs
  being applied at the same time. (@rfratto)

- [BUGFIX] `remote_write` names in a group will no longer be copied from the
  remote_write names of the first instance in the group. Rather, all
  remote_write names will be generated based on the first 6 characters of the
  group hash and the first six characters of the remote_write hash. (@rfratto)

- [BUGFIX] Fix a panic that may occur during shutdown if the WAL is closed in
  the middle of the WAL being truncated. (@rfratto)

- [DEPRECATION] `use_hostname_label` is now supplanted by
  `replace_instance_label`. `use_hostname_label` will be removed in a future
  version. (@rfratto)

# v0.5.0 (2020-08-12)

- [FEATURE] A [scrape targets API](https://github.com/grafana/agent/blob/master/docs/api.md#list-current-scrape-targets)
  has been added to show every target the Agent is currently scraping, when it
  was last scraped, how long it took to scrape, and errors from the last scrape,
  if any. (@rfratto)

- [FEATURE]  "Shared Instance Mode" is the new default mode for spawning
  Prometheus instances, and will improve CPU and memory usage for users of
  integrations and the scraping service. (@rfratto)

- [ENHANCEMENT] Memory stability and utilization of the WAL has been improved,
  and the reported number of active series in the WAL will stop double-counting
  recently churned series. (@rfratto)

- [ENHANCEMENT] Changing scrape_configs and remote_write configs for an instance
  will now be dynamically applied without restarting the instance. This will
  result in less missing metrics for users of the scraping service that change a
  config. (@rfratto)

- [ENHANCEMENT] The Tanka configuration now uses k8s-alpha. (@duologic)

- [BUGFIX] The Tanka configuration will now also deploy a single-replica
  deployment specifically for scraping the Kubernetes API. This deployment acts
  together with the Daemonset to scrape the full cluster and the control plane.
  (@gotjosh)

- [BUGFIX] The node_exporter filesystem collector will now work on Linux systems
  without needing to manually set the blocklist and allowlist of filesystems.
  (@rfratto)

# v0.4.0 (2020-06-18)

- [FEATURE] Support for integrations has been added. Integrations can be any
  embedded tool, but are currently used for embedding exporters and generating
  scrape configs. (@rfratto)

- [FEATURE] node_exporter has been added as an integration. This is the full
  version of node_exporter with the same configuration options. (@rfratto)

- [FEATURE] An Agent integration that makes the Agent automatically scrape
  itself has been added. (@rfratto)

- [ENHANCEMENT] The WAL can now be truncated if running the Agent without any
  remote_write endpoints. (@rfratto)

- [ENHANCEMENT] Clarify server_config description in documentation. (@rfratto)

- [ENHANCEMENT] Clarify wal_truncate_frequency and remote_flush_deadline in
  documentation. (@rfratto)

- [ENHANCEMENT] Document /agent/api/v1/instances endpoint (@rfratto)

- [ENHANCEMENT] Be explicit about envsubst requirement for Kubernetes install
  script. (@robx)

- [BUGFIX] Prevent the Agent from crashing when a global Prometheus config
  stanza is not provided. (@robx)

- [BUGFIX] Enable agent host_filter in the Tanka configs, which was disabled by
  default by mistake. (@rfratto)

# v0.3.2 (2020-05-29)

- [FEATURE] Tanka configs that deploy the scraping service mode are now
  available (@rfratto)

- [FEATURE] A k3d example has been added as a counterpart to the docker-compose
  example. (@rfratto)

- [ENHANCEMENT] Labels provided by the default deployment of the Agent
  (Kubernetes and Tanka) have been changed to align with the latest changes to
  grafana/jsonnet-libs. The old `instance` label is now called `pod`, and the
  new `instance` label is unique. A `container` label has also been added. The
  Agent mixin has been subsequently updated to also incorporate these label
  changes. (@rfratto)

- [ENHANCEMENT] The `remote_write` and `scrape_config` sections now share the
  same validations as Prometheus (@rfratto)

- [ENHANCEMENT] Setting `wal_truncation_frequency` to less than the scrape
  interval is now disallowed (@rfratto)

- [BUGFIX] A deadlock in scraping service mode when updating a config that
  shards to the same node has been fixed (@rfratto)

- [BUGFIX] `remote_write` config stanzas will no longer ignore `password_file`
  (@rfratto)

- [BUGFIX] `scrape_config` client secrets (e.g., basic auth, bearer token,
  `password_file`) will now be properly retained in scraping service mode
  (@rfratto)

- [BUGFIX] Labels for CPU, RX, and TX graphs in the Agent Operational dashboard
  now correctly show the pod name of the Agent instead of the exporter name.
  (@rfratto)

# v0.3.1 (2020-05-20)

- [BUGFIX] A typo in the Tanka configs and Kubernetes manifests that prevents
  the Agent launching with v0.3.0 has been fixed (@captncraig)

- [BUGFIX] Fixed a bug where Tanka mixins could not be used due to an issue with
  the folder placement enhancement (@rfratto)

- [ENHANCEMENT] `agentctl` and the config API will now validate that the YAML
  they receive are valid instance configs. (@rfratto)

- [FEATURE] The Agent has upgraded its vendored Prometheus to v2.18.1
  (@gotjosh, @rfratto)

# v0.3.0 (2020-05-13)

- [FEATURE] A third operational mode called "scraping service mode" has been
  added. A KV store is used to store instance configs which are distributed
  amongst a clustered set of Agent processes, dividing the total scrape load
  across each agent. An API is exposed on the Agents to list, create, update,
  and delete instance configurations from the KV store. (@rfratto)

- [FEATURE] An "agentctl" binary has been released to interact with the new
  instance config management API created by the "scraping service mode."
  (@rfratto, @hoenn)

- [FEATURE] The Agent now includes readiness and healthiness endpoints.
  (@rfratto)

- [ENHANCEMENT] The YAML files are now parsed strictly and an invalid YAML will
  generate an error at runtime. (@hoenn)

- [ENHANCEMENT] The default build mode for the Docker containers is now release,
  not debug. (@rfratto)

- [ENHANCEMENT] The Grafana Agent Tanka Mixins now are placed in an "Agent"
  folder within Grafana. (@cyriltovena)

# v0.2.0 (2020-04-09)

- [FEATURE] The Prometheus remote write protocol will now send scraped metadata (metric name, help, type and unit). This results in almost negligent bytes sent increase as metadata is only sent every minute. It is on by default. (@gotjosh)

  These metrics are available to monitor metadata being sent:
    - `prometheus_remote_storage_succeeded_metadata_total`
    - `prometheus_remote_storage_failed_metadata_total`
    - `prometheus_remote_storage_retried_metadata_total`
    - `prometheus_remote_storage_sent_batch_duration_seconds` and
      `prometheus_remote_storage_sent_bytes_total` have a new label “type” with
      the values of `metadata` or `samples`.

- [FEATURE] The Agent has upgraded its vendored Prometheus to v2.17.1 (@rfratto)

- [BUGFIX] Invalid configs passed to the agent will now stop the process after they are logged as invalid; previously the Agent process would continue. (@rfratto)

- [BUGFIX] Enabling host_filter will now allow metrics from node role Kubernetes service discovery to be scraped properly (e.g., cAdvisor, Kubelet). (@rfratto)

# v0.1.1 (2020-03-16)

- Nits in documentation (@sh0rez)
- Fix various dashboard mixin problems from v0.1.0 (@rfratto)
- Pass through release tag to `docker build` (@rfratto)

# v0.1.0 (2020-03-16)

First (beta) release!

This release comes with support for scraping Prometheus metrics and
sharding the agent through the presence of a `host_filter` flag within the
Agent configuration file.

Note that enabling the `host_filter` flag currently works best when using our
preferred Kubernetes deployment, as it deploys the agent as a DaemonSet.
