---
aliases:
- ../../../dynamic-configuration/looping/
title: Looping
weight: 200
aliases:
  - ../../../../dynamic-configuration/looping/
  - ../../../../cookbook/dynamic-configuration/looping/
  - ../../../../dynamic-configuration/02_templates/01_looping/
  - ../../../../cookbook/dynamic-configuration/02_templates/01_looping/
---

# 01 Basics of Templating

The templating is based on the excellent [gomplate](https://docs.gomplate.ca/) library. Currently using a custom fork to allow loading gomplate as a library in addition to some new commands. This will NOT try to cover the full range of gomplate, would recommend reading the documentation for full knowledge.

`docker run -v ${PWD}/:/etc/grafana grafana/agentctl:latest template-parse file:///etc/grafana/01_config.yml`

## Looping

[agent-1.yml](https://github.com/grafana/agent/blob/main/docs/sources/cookbook/dynamic-configuration/02_Templates/01_assets/agent-1.yml)

```yaml
server:
  log_level: debug
metrics:
  wal_directory: /tmp/grafana-agent-normal
  global:
    scrape_interval: 60s
    remote_write:
      - url: https://prometheus-us-central1.grafana.net/api/prom/push
        basic_auth:
          username: xyz
          password: secretpassword
  configs:
    - name: default
  {{ range slice "apple" "banana" "pear" }}
    - name: {{ . }}
  {{ end }}
```

The templating engine uses directives that are wrapped in `{{ command }}`, in the above the dynamic configuration engine will loop over the three values, and those values can be accessed by `{{ . }}` which means current value.

## Final

[final.yml](https://github.com/grafana/agent/blob/main/docs/sources/cookbook/dynamic-configuration/02_Templates/01_assets/final.yml)

The final.yml contains 4 prometheus configs

- default
- apple
- banana
- pear
