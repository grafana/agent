---
aliases:
- ../../../dynamic-configuration/objects/
title: Objects
weight: 220
---

# 02 Datasources

Datasources can also access objects.

`docker run -v ${PWD}/:/etc/grafana grafana/agentctl:latest template-parse file:///etc/grafana/03_config.yml`


## Config

The [config.yml](https://github.com/grafana/agent/blob/main/docs/sources/cookbook/dynamic-configuration/02_Templates/02_config.yml) adds a new field `sources`. Sources can be any number of things defined in the gomplate [datasources](https://docs.gomplate.ca/datasources/) documentation. In this example using fruit.

```yaml
template_paths:
  - "file:///etc/grafana/03_assets"
datasources:
  - name: computers
    url: "file:///etc/grafana/03_assets/computers.json"
```

[computers.json](https://github.com/grafana/agent/blob/main/docs/sources/cookbook/dynamic-configuration/02_Templates/03_assets/computers.json)

```json
[
  {
    "name": "webhost1",
    "ip" : "192.168.1.1",
    "enabled": true
  },
  {
    "name": "webhost2",
    "ip" : "192.168.1.2",
    "enabled": false

  },
  {
    "name": "webhost3",
    "ip" : "192.168.1.3",
    "enabled": true
  }
]
```

## Usage

[agent-1.yml](https://github.com/grafana/agent/blob/main/docs/sources/cookbook/dynamic-configuration/02_Templates/02_assets/agent-1.yml)

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
      # Check for length so that if it is 0, we dont write any scrape configs
  {{ if $length := len (datasource "computers") }}
  {{ if gt $length 0 }}
  scrape_configs:
  {{ end }}
  {{ end }}
  {{ range (datasource "computers") }}
  # Only add if the computers are enabled
  # the . references our current object
  {{ if eq .enabled true }}
  - job_name: {{ .name }}
    static_configs:
      - targets:
          - {{ .ip }}
  {{ end }}
  {{ end }}
```

This is a much more complex example, in the above we are doing:

- comparisons
- creating and setting variables
- looping over objects

The final output will only list `webhost1` and `webhost3` since `webhost2` is not enabled.
