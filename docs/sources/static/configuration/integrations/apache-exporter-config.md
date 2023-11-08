---
aliases:
- ../../../configuration/integrations/apache-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/apache-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/apache-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/apache-exporter-config/
description: Learn about apache_http_config
title: apache_http_config
---

# apache_http_config

The `apache_http_config` block configures the `apache_http` integration,
which is an embedded version of
[`apache_exporter`](https://github.com/Lusitaniae/apache_exporter). This allows the collection of Apache [mod_status](https://httpd.apache.org/docs/current/mod/mod_status.html) statistics via HTTP.

Full reference of options:

```yaml
  # Enables the apache_http integration, allowing the Agent to automatically
  # collect metrics for the specified apache http servers.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the hostname portion
  # of api_url.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the apache_http integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/apache_http/metrics and can be scraped by an external
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

  # URI to apache stub status page.
  # If your server-status page is secured by http auth, add the credentials to the scrape URL following this example:
  # http://user:password@localhost/server-status?auto .
  [scrape_uri: <string> | default = "http://localhost/server-status?auto"]

  # Override for HTTP Host header; empty string for no override.
  [host_override: <string> | default = ""]

  # Ignore server certificate if using https.
  [insecure: <bool> | default = false]

```
