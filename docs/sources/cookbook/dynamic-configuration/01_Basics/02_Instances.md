---
aliases:
- ../../../dynamic-configuration/instances/
title: Instances
weight: 110
---

# 02 Instances

Dynamic configuration allows multiple prometheus instances to be loaded with a parent metric. This uses
the same agent-1 and server-1 yml from 01.

`docker run -v ${PWD}/:/etc/grafana grafana/agentctl:latest template-parse file:///etc/grafana/02_config.yml`

## Dynamic Configuration

[config.yml](https://github.com/grafana/agent/blob/main/docs/sources/cookbook/dynamic-configuration/01_Basics/02_config.yml)

Tells the Grafana Agent where to load files from.

## Metrics

Dynamic Configuration will find the first file matching pattern `metrics-*.yml` and load that as the base. You can only have one metrics template.

[metrics-1.yml](https://github.com/grafana/agent/blob/main/docs/sources/cookbook/dynamic-configuration/01_Basics/02_assets/metrics-1.yml)

```yaml
configs:
  - name: default
global:
  scrape_interval: 60s
  scrape_timeout: 20s
wal_directory: /tmp/grafana-agent-wal
```

## Metrics Instances

You can have any number of metrics_instances and they are added to any existing metrics instances defined previously.

[metrics_instances-1.yml](https://github.com/grafana/agent/blob/main/docs/sources/cookbook/dynamic-configuration/01_Basics/02_assets/metrics_instances-1.yml)

```yaml
name: instance1
scrape_configs:
  - job_name: instance1_job
    static_configs:
      - targets:
          - localhost:4000
```

[metrics_instances-2.yml](https://github.com/grafana/agent/blob/main/docs/sources/cookbook/dynamic-configuration/01_Basics/02_assets/metrics_instances-2.yml)

```yaml
name: instance2
scrape_configs:
  - job_name: instance2_job
    static_configs:
      - targets:
          - localhost:5555
```

## Final

[final.yml](https://github.com/grafana/agent/blob/main/docs/sources/cookbook/dynamic-configuration/01_Basics/02_assets/final.yml)

In the above you will see the `final.yml` includes all the instance configurations
- default
- instance1
- instance2

