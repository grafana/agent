# Next (master/unreleased)

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
