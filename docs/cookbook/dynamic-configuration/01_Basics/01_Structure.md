# 01 Structure

Dynamic Configuration uses a series of files to load templates. This example will show how they all combine together. 
Running the below command will combine all the templates into the final.yml

`docker run -v ${PWD}/:/etc/grafana grafana/agentctl:latest template-parse /etc/grafana/02_config.yml`

## Dynamic Configuration

[config.yml](./01_config.yml)

Tells the Grafana Agent where to load files from.

## Agent

Dynamic Configuration will find the first file matching pattern `agent-*.yml` and load that as the base. You can only have
one agent template.

[agent-1.yml](./01_assets/agent-1.yml)

```yaml
server:
  http_listen_port: 12345
  log_level: debug
metrics:
  wal_directory: /tmp/grafana-agent-normal
  global:
    scrape_interval: 60s
    remote_write:
      - url: https://prometheus-us-central1.grafana.net/api/prom/push
        basic_auth:
          username: 12345
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

[server-1.yml](./01_assets/server-1.yml)


```yaml
http_listen_port: 12345
log_level: info
```

## Final

[final.yml](./01_assets/final.yml)

In the above example the `log_level: debug` block will be replaced with `log_level: info` from the server-1.yml
