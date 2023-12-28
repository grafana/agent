---
aliases:
- ../../../configuration/integrations/github-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/github-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/github-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/github-exporter-config/
description: Learn about github_exporter_config
title: github_exporter_config
---

# github_exporter_config

The `github_exporter_config` block configures the `github_exporter` integration,
which is an embedded version of
[`github_exporter`](https://github.com/githubexporter/github-exporter). This allows for the collection of metrics from the GitHub api.

We strongly recommend that you configure a separate authentication token for the Agent, and give it only the strictly mandatory
security privileges necessary for monitoring your repositories, as per the [official documentation](https://docs.github.com/en/rest/reference/permissions-required-for-github-apps).
We also recommend that you use `api_token_file` parameter, to avoid setting the authentication token directly on the Agent config file.

Full reference of options:

```yaml
  # Enables the github_exporter integration, allowing the Agent to automatically
  # collect metrics for the specified GitHub objects.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the hostname portion
  # of api_url.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the github_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/github_exporter/metrics and can be scraped by an external
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

  # The full URI of the GitHub API.
  [api_url: <string> | default = "https://api.github.com"]

  # A list of GitHub repositories for which to collect metrics.
  repositories:
    [ - <string> ]

  # A list of GitHub organizations for which to collect metrics.
  organizations:
    [ - <string> ]

  # A list of GitHub users for which to collect metrics.
  users:
    [ - <string> ]

  # A GitHub authentication token that allows the API to be queried more often.
  # Optional, but recommended.
  [api_token: <string>]

  # A path to a file containing a GitHub authentication token that allows the
  # API to be queried more often. If supplied, this supersedes `api_token`
  # Optional, but recommended.
  [api_token_file: <string>]
```
