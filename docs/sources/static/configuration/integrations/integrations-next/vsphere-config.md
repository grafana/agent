---
aliases:
- ../../../../configuration/integrations/integrations-next/vsphere-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/integrations-next/vsphere-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/integrations-next/vsphere-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/integrations-next/vsphere-config/
description: Learn about vsphere_config next
menuTitle: vsphere_config next
title: vsphere config (beta) next
---

# vsphere config (beta) next

The `vsphere_config` block configures the `vmware_exporter` integration, an embedded
version of [`vmware_exporter`](https://github.com/grafana/vmware_exporter), configured
to collect vSphere metrics. This integration is considered beta.

Configuration reference:

```yaml
  autoscrape:
    # Enables autoscrape of integrations.
    [enable: <boolean> | default = true]

    # Specifies the metrics instance name to send metrics to. Instance
    # names are located at metrics.configs[].name from the top-level config.
    # The instance must exist.
    #
    # As it is common to use the name "default" for your primary instance,
    # we assume the same here.
    [metrics_instance: <string> | default = "default"]

    # Autoscrape interval and timeout. Defaults are inherited from the global
    # section of the top-level metrics config.
    [scrape_interval: <duration> | default = <metrics.global.scrape_interval>]
    [scrape_timeout: <duration> | default = <metrics.global.scrape_timeout>]

  # Integration instance name. This will default to the host:port of the configured
  # vsphere_url.
  [instance: <string> | default = <vsphere_url>]

  # Number of managed objects to include in each request to vsphere when
  # fetching performance counters.
  [request_chunk_size: <int> | default = 256]

  # Number of concurrent requests to vsphere when fetching performance counters.
  [collect_concurrency: <int> | default = 8]

  # Interval on which to run vsphere managed object discovery. Setting this to a
  # non-zero value will result in object discovery running in the background. Each
  # scrape will use object data gathered during the last discovery.
  # When this value is 0, object discovery occurs per scrape.
  [discovery_interval: <duration> | default = 0]
  [enable_exporter_metrics: <boolean> | default = true]

  # The url of the vCenter SDK endpoint
  vsphere_url: <string>

  # vCenter username
  vsphere_user: <string>

  # vCenter password
  vsphere_password: <string>

```

## Quick configuration example

```yaml
integrations:
  vsphere_configs:
    - vsphere_url: https://127.0.0.1:8989/sdk
      vsphere_user: user
      vsphere_password: pass
      request_chunk_size: 256
      collect_concurrency: 8
      instance: vsphere
      autoscrape:
        enable: true
        metrics_instance: default

metrics:
  wal_directory: /tmp/grafana-agent-wal
server:
  log_level: debug
```
