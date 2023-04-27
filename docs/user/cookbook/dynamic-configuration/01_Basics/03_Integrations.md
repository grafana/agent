---
aliases:
- ../../../dynamic-configuration/integrations/
title: Integrations
weight: 120
---

# 03 Integrations

Dynamic configuration requires the use of `integrations-next` feature flag, to allow arrays of integrations. In this we will load integrations of various types. This is all built on the previous examples.

`docker run -v ${PWD}/:/etc/grafana grafana/agentctl:latest template-parse file:///etc/grafana/03_config.yml`

## Dynamic Configuration

[config.yml](https://github.com/grafana/agent/blob/main/docs/user/cookbook/dynamic-configuration/01_Basics/03_config.yml)

Tells the Grafana Agent where to load files from.

## Integrations

Integrations are loaded from files matching `integrations-*.yml` and are combined together. You can declare for example multiple sets of `redis_configs` across several files.

[integrations-node.yml](https://github.com/grafana/agent/blob/main/docs/user/cookbook/dynamic-configuration/01_Basics/03_assets/integrations-node.yml)

Note: You do NOT have to name the above file `integrations-node.yml` with `node`, `integrations-1.yml` would work the same. The name does NOT determine the type of integrations a template can contain and a template can contain integrations of different types.

```yaml
node_exporter: {}
```

[integrations-redis.yml](https://github.com/grafana/agent/blob/main/docs/user/cookbook/dynamic-configuration/01_Basics/03_assets/integrations-redis.yml)

```yaml
redis_configs:
  - redis_addr: localhost:6379
    autoscrape:
      metric_relabel_configs:
        - source_labels: [__address__]
          target_label: "banana"
          replacement: "apple"
  - redis_addr: localhost:6380
```

## Final

[final.yml](https://github.com/grafana/agent/blob/main/docs/user/cookbook/dynamic-configuration/01_Basics/03_assets/final.yml)

The final result should have 3 integrations enabled, 1 node_exporter and 2 redis.

