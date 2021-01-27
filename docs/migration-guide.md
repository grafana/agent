# Migration Guide

This is a guide detailing all breaking changes that have happened in prior
releases and how to migrate to newer versions.

# v0.12.0

## Loki Promtail Config Change 

The Loki Promtail config (`loki` in the config file) has been changed to move to
a `configs` key. 

To migrate, add a `configs:` array and move your existing config inside of it.
Give the element a `name: default` field.

Example old config:

```yaml
loki:
  positions:
    filename: /tmp/positions.yaml
  clients:
    - url: http://loki:3100/loki/api/v1/push
  scrape_configs:
  - job_name: system
    static_configs:
      - targets:
        - localhost
        labels:
          job: varlogs
          __path__: /var/log/*log
```

Example migrated config:

```yaml
loki:
  configs:
  - name: default
    positions:
      filename: /tmp/positions.yaml
    clients:
      - url: http://loki:3100/loki/api/v1/push
    scrape_configs:
    - job_name: system
      static_configs:
        - targets:
          - localhost
          labels:
            job: varlogs
            __path__: /var/log/*log
```
