---
aliases:
  - /docs/agent/latest/configuration/integrations/integrations-next/vsphere-config/
title: vsphere_config
---

# vsphere config (beta)

The `vsphere_config` block configures the `vmware_exporter` integration, an embedded
version of [`vmware_exporter`](https://github.com/grafana/vmware_exporter), configured
to collect vSphere metrics. This integration is considered beta.

## Quick configuration example

```yaml
integrations:
  vsphere_configs:
    - vsphere_url: https://127.0.0.1:8989/sdk
      vsphere_user: user
      vsphere_password: pass
      chunk_size: 256
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
