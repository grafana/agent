# Next (master/unreleased)

# v0.2.0 (2020-04-09)

- [FEATURE] The Prometheus remote write protocol will now send scraped metadata (metric name, help, type and unit). This results in almost negligent bytes sent increase as metadata is only sent every minute. It is on by default.

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
