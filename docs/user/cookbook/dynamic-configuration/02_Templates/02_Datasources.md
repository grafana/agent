---
aliases:
- ../../../dynamic-configuration/datasources/
title: Datasources
weight: 210
---

# 02 Datasources

Datasources are a powerful concept in gomplate. They allow you to reach out to other files, systems and resources to pull data.

`docker run -v ${PWD}/:/etc/grafana grafana/agentctl:latest template-parse file:///etc/grafana/02_config.yml`


## Config

The [config.yml](https://github.com/grafana/agent/blob/main/docs/user/cookbook/dynamic-configuration/02_Templates/02_config.yml) adds a new field `sources`. Sources can be any number of things defined in the gomplate [datasources](https://docs.gomplate.ca/datasources/) documentation. In this example using fruit.

```yaml
template_paths:
  - "file:///etc/grafana/01_assets"
datasources:
  - name: fruit
    url: "file://etc/grafana/01_assets/fruit.json"
```

[fruit.json](https://github.com/grafana/agent/blob/main/docs/user/cookbook/dynamic-configuration/02_Templates/02_assets/fruit.json)

```json
["mango","peach","orange"]
```

## Usage

[agent-1.yml](https://github.com/grafana/agent/blob/main/docs/user/cookbook/dynamic-configuration/02_Templates/02_assets/agent-1.yml)

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
  {{ range (datasource "fruit") }}
    - name: {{ . }}
  {{ end }}
```

A Datasource is reference by name and in this case it is an array and used exactly like the looping example.
