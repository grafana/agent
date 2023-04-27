---
aliases:
- ../../../dynamic-configuration/structure/
title: Structure
weight: 100
---

# 01 Structure

Dynamic Configuration uses a series of files to load templates. This example will show how they all combine together. Running the below command will combine all the templates into the final.yml. Any failure while loading the config will revert to the original config, or if this is the initial load Grafana Agent will quit.

`docker run -v ${PWD}/:/etc/grafana grafana/agentctl:latest template-parse file:///etc/grafana/01_config.yml`

## Dynamic Configuration

[config.yml](https://github.com/grafana/agent/blob/main/docs/user/cookbook/dynamic-configuration/01_Basics/01_config.yml)

```yaml
template_paths:
  - "file:///etc/grafana/01_assets"
```

Tells the Grafana Agent where to load files from. It is important to note that dynamic configuration does NOT traverse directories. It will look at the directory specified only, if you need more directories then add them to the `template_paths` array. NOTE, if no protocol specified ie `file://` above, then file access will be assumed. `file:///etc/grafana/01_assets` is equivalent to `//etc/grafana/01_assets`

## Agent

Dynamic Configuration will find the first file matching pattern `agent-*.yml` and load that as the base. You can only have one agent template. If multiple matching templates are found then the configuration will fail to load.

[agent-1.yml](https://github.com/grafana/agent/blob/main/docs/user/cookbook/dynamic-configuration/01_Basics/01_assets/agent-1.yml)

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
integrations:
  node_exporter:
    enabled: true
  agent:
    enabled: true
```

## Server

Dynamic configuration will find the first file matching pattern `server-*.yml` and replace the `Server` config block in
the Agent Configuration. Note that you do NOT include the `server:` tag, dynamic configuration knows by the name that it
is a configuration block.

You can only have 1 server template.

[server-1.yml](https://github.com/grafana/agent/blob/main/docs/user/cookbook/dynamic-configuration/01_Basics/01_assets/server-1.yml)


```yaml
log_level: info
```

## Final

[final.yml](https://github.com/grafana/agent/blob/main/docs/user/cookbook/dynamic-configuration/01_Basics/01_assets/final.yml)

In the above example the `log_level: debug` block will be replaced with `log_level: info` from the server-1.yml
