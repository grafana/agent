---
aliases:
- ../../../configuration/integrations/mongodb_exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/mongodb_exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/mongodb_exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/mongodb_exporter-config/
description: Learn about mongodb_exporter_config
title: mongodb_exporter_config
---

# mongodb_exporter_config

The `mongodb_exporter_config` block configures the `mongodb_exporter` integration, which is an embedded version of percona's [`mongodb_exporter`](https://github.com/percona/mongodb_exporter).

In order for this integration to work properly, you have to connect each node of your mongoDB cluster to an agent instance.
That's because this exporter does not collect metrics from multiple nodes.
Additionally, you need to define two custom label for your metrics using relabel_configs.
The first one is service_name, which is how you identify this node in your cluster (example: ReplicaSet1-Node1).
The second one is mongodb_cluster, which is the name of your mongodb cluster, and must be set the same value for all nodes composing the cluster (example: prod-cluster).
Here`s an example:

```yaml
relabel_configs:
    - source_labels: [__address__]
      target_label: service_name
      replacement: 'replicaset1-node1'
    - source_labels: [__address__]
      target_label: mongodb_cluster
      replacement: 'prod-cluster'
```

We strongly recommend that you configure a separate user for the Agent, and give it only the strictly mandatory
security privileges necessary for monitoring your node, as per the [official documentation](https://github.com/percona/mongodb_exporter#permissions).

Besides that, there's not much to configure. Please refer to the full reference of options:

```yaml
  # Enables the mongodb_exporter integration
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the hostname
  # portion of the mongodb_uri field.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the mongodb_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/mongodb_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # metrics.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # metrics.global.scrape_timeout.
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
  # Exporter-specific configuration options
  #

  # MongoDB node connection URL, which must be in the [`Standard Connection String Format`](https://docs.mongodb.com/manual/reference/connection-string/#std-label-connections-standard-connection-string-format)
  [mongodb_uri: <string>]

  # Whether or not a direct connect should be made. Direct connections are not valid if multiple hosts are specified or an SRV URI is used
  [direct_connect: <boolean> | default = true]

  # Enable autodiscover collections
  [discovering_mode: <boolean> | default = false]

  # Path to the file having Prometheus TLS config for basic auth. Only enable if you want to use TLS based authentication.
  [tls_basic_auth_config_path: <string> | default = ""]
```

For `tls_basic_auth_config_path`, check [`tls_config`](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#tls_config) for reference on the file format to be used.
