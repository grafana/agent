---
aliases:
- ../../../configuration/integrations/elasticsearch-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/elasticsearch-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/elasticsearch-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/elasticsearch-exporter-config/
description: Learn about elasticsearch_exporter_config
title: elasticsearch_exporter_config
---

# elasticsearch_exporter_config

The `elasticsearch_exporter_config` block configures the `elasticsearch_exporter` integration,
which is an embedded version of
[`elasticsearch_exporter`](https://github.com/prometheus-community/elasticsearch_exporter). This allows for
the collection of metrics from ElasticSearch servers.

Note that currently, an Agent can only collect metrics from a single ElasticSearch server.
However, the exporter is able to collect the metrics from all nodes through that server configured.

We strongly recommend that you configure a separate user for the Agent, and give it only the strictly mandatory
security privileges necessary for monitoring your node, as per the [official documentation](https://github.com/prometheus-community/elasticsearch_exporter#elasticsearch-7x-security-privileges).

Full reference of options:

```yaml
  # Enables the elasticsearch_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured ElasticSearch server address
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the hostname portion
  # of address.
  [instance: <string>]

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

  #
  # Exporter-specific configuration options
  #

  # HTTP API address of an Elasticsearch node.
  [ address: <string> | default = "http://localhost:9200" ]

  # Timeout for trying to get stats from Elasticsearch.
  [ timeout: <duration> | default = "5s" ]

  # Export stats for all nodes in the cluster. If used, this flag will override the flag `node`.
  [ all: <boolean> ]

  # Node's name of which metrics should be exposed.
  [ node: <string> ]

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

  # Include informational aliases metrics.
  [ aliases: <boolean> ]

  # Export stats for Data Streams.
  [ data_stream: <boolean> ]

  # Export stats for SLM (Snapshot Lifecycle Management).
  [ slm: <boolean> ]

  # Sets the `Authorization` header on every ES probe with the
  # configured username and password.
  # password and password_file are mutually exclusive.
  basic_auth:
    [ username: <string> ]
    [ password: <secret> ]
    [ password_file: <string> ]
```
