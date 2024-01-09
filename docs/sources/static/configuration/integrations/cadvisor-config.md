---
aliases:
- ../../../configuration/integrations/cadvisor-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/cadvisor-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/cadvisor-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/cadvisor-config/
description: Learn about cadvisor_config
title: cadvisor_config
---

# cadvisor_config

The `cadvisor_config` block configures the `cadvisor` integration,
which is an embedded version of
[`cadvisor`](https://github.com/google/cadvisor). This allows for the collection of container utilization metrics.

The cAdvisor integration requires some broad privileged permissions to the host. Without these permissions the metrics will not be accessible. This means that the agent must *also* have those elevated permissions.

A good example of the required file, and system permissions can be found in the docker run command published in the [cAdvisor docs](https://github.com/google/cadvisor#quick-start-running-cadvisor-in-a-docker-container).

Full reference of options:

```yaml
  # Enables the cadvisor integration, allowing the Agent to automatically
  # collect metrics for the specified github objects.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  [instance: <string> | default = <integrations_config.instance>]

  # Automatically collect metrics from this integration. If disabled,
  # the cadvisor integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/cadvisor/metrics and can be scraped by an external
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

  #
  # cAdvisor-specific configuration options
  #

  # Convert container labels and environment variables into labels on Prometheus metrics for each container. If false, then the only metrics exported are container name, first alias, and image name. `.` aren't valid in Prometheus label names, so if there are any in the container label, they will transformed to `_` when converted to the Prometheus label. 
  [store_container_labels: <boolean> | default = true]

  # List of container labels to be converted to labels on Prometheus metrics for each container. store_container_labels must be set to false for this to take effect. This must match the format of the container label, not the converted Prometheus label (`.` are converted to `_` in the Prometheus label).   
  allowlisted_container_labels:
    [ - <string> ]

  # List of environment variable keys matched with specified prefix that needs to be collected for containers, only support containerd and docker runtime for now.
  env_metadata_allowlist:
    [ - <string> ]

  # List of cgroup path prefix that needs to be collected even when docker_only is specified.
  raw_cgroup_prefix_allowlist:
    [ - <string> ]

  # Path to a JSON file containing configuration of perf events to measure. Empty value disabled perf events measuring.
  [perf_events_config: <boolean>]

  # resctrl mon groups updating interval. Zero value disables updating mon groups.
  [resctrl_interval: <int> | default = 0]

  # List of `metrics` to be disabled. If set, overrides the default disabled metrics.
  disabled_metrics:
    [ - <string> ]

  # List of `metrics` to be enabled. If set, overrides disabled_metrics
  enabled_metrics:
    [ - <string> ]

  # Length of time to keep data stored in memory
  [storage_duration: <duration> | default = "2m"]

  # Containerd endpoint
  [containerd: <string> | default = "/run/containerd/containerd.sock"]

  # Containerd namespace
  [containerd_namespace: <string> | default = "k8s.io"]

  # Docker endpoint
  [docker: <string> | default = "unix:///var/run/docker.sock"]

  # Use TLS to connect to docker
  [docker_tls: <boolean> | default = false]

  # Path to client certificate for TLS connection to docker
  [docker_tls_cert: <string> | default = "cert.pem"]

  # Path to private key for TLS connection to docker
  [docker_tls_key: <string> | default = "key.pem"]

  # Path to a trusted CA for TLS connection to docker
  [docker_tls_ca: <string> | default = "ca.pem"]

  # Only report docker containers in addition to root stats
  [docker_only: <boolean> | default = false]

  # Disable collecting root Cgroup stats
  [disable_root_cgroup_stats: <boolean> | default = false]
```
