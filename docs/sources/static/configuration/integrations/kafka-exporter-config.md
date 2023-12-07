---
aliases:
- ../../../configuration/integrations/kafka-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/kafka-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/kafka-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/kafka-exporter-config/
description: Learn about kafka_exporter_config
title: kafka_exporter_config
---

# kafka_exporter_config

The `kafka_exporter_config` block configures the `kafka_exporter`
integration, which is an embedded version of [`kafka_exporter`](https://github.com/davidmparrott/kafka_exporter).
This allows for the collection of Kafka Lag metrics and exposing them as Prometheus metrics.

We strongly recommend that you configure a separate user for the Agent, and give it only the strictly mandatory
security privileges necessary for monitoring your node, as per the [documentation](https://github.com/lightbend/kafka-lag-exporter#required-permissions-for-kafka-acl).

Full reference of options:

```yaml
  # Enables the kafka_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured dnsmasq server address
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the hostname
  # portion of the first kafka_uri value. If there is more than one string
  # in kafka_uri, the integration will fail to load and an instance value
  # must be manually provided.
  [instance: <string>]

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

  # Address array (host:port) of Kafka server
  [kafka_uris: <[]string>]

  # Connect using SASL/PLAIN
  [use_sasl: <bool>]

  # Only set this to false if using a non-Kafka SASL proxy
  [use_sasl_handshake: <bool> | default = true]

  # SASL user name
  [sasl_username: <string>]

  # SASL user password
  [sasl_password: <string>]

  # The SASL SCRAM SHA algorithm sha256 or sha512 as mechanism
  [sasl_mechanism: <string>]

  # Connect using TLS
  [use_tls: <bool>]

  # The optional certificate authority file for TLS client authentication
  [ca_file: <string>]

  # The optional certificate file for TLS client authentication
  [cert_file: <string>]

  # The optional key file for TLS client authentication
  [key_file: <string>]

  # If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
  [insecure_skip_verify: <bool>]

  # Kafka broker version
  [kafka_version: <string> | default = "2.0.0"]

  # if you need to use a group from zookeeper
  [use_zookeeper_lag: <bool>]

  # Address array (hosts) of zookeeper server.
  [zookeeper_uris: <[]string>]

  # Kafka cluster name
  [kafka_cluster_name: <string>]

  # Metadata refresh interval
  [metadata_refresh_interval: <duration> | default = "1m"]

  # If true, all scrapes will trigger kafka operations otherwise, they will share results. WARN: This should be disabled on large clusters
  [allow_concurrency: <bool> | default = true]

  # Maximum number of offsets to store in the interpolation table for a partition
  [max_offsets: <int> | default = 1000]

  # How frequently should the interpolation table be pruned, in seconds
  [prune_interval_seconds: <int> | default = 30]

  # Regex filter for topics to be monitored
  [topics_filter_regex: <string> | default = ".*"]

  # Regex filter for consumer groups to be monitored
  [groups_filter_regex: <string> | default = ".*"]

```
