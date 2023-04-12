---
title: logs_config
weight: 300
aliases:
- ../../configuration/loki-config/
- ../../configuration/logs-config/
---

# logs_config

The `logs_config` block configures how the Agent collects logs and sends them to
a Loki push API endpoint. `logs_config` is identical to how Promtail is
configured, except deprecated fields have been removed and the server_config is
not supported.

Refer to the
[Promtail documentation](https://grafana.com/docs/loki/latest/clients/promtail/configuration/#clients)
for the supported values for these fields.

```yaml
# Directory to store Loki Promtail positions files in. Positions files are
# required to read logs, and are used to store the last read offset of log
# sources. The positions files will be stored in
# <positions_directory>/<logs_instance_config.name>.yml.
#
# Optional only if every config has a positions.filename manually provided.
#
# This directory will be automatically created if it doesn't exist.
[positions_directory: <string>]

# Configure values for all Loki Promtail instances.
[global: <global_config>]

# Loki Promtail instances to run for log collection.
configs:
  - [<logs_instance_config>]
```

## global_config

The `global_config` block configures global values for all launched Loki Promtail
instances.

```yaml
clients:
  - [<promtail.client_config>]
```

> **Note:** More information on the following types can be found on the
> documentation for Promtail:
>
> * [`promtail.client_config`](https://grafana.com/docs/loki/latest/clients/promtail/configuration/#clients)

## logs_instance_config

The `logs_instance_config` block is an individual instance of Promtail with its
own set of scrape rules and where to forward logs. It is identical to how
Promtail is configured, except deprecated fields have been removed and the
`server_config` block is not supported.

```yaml
# Name of this config. Required, and must be unique across all Loki configs.
# The name of the config will be the value of a logs_config label for all
# Loki Promtail metrics.
name: <string>

clients:
  - [<promtail.client_config>]

# Optional configuration for where to store the positions files. If
# positions.filename is left empty, the file will be stored in
# <logs_config.positions_directory>/<logs_instance_config.name>.yml.
#
# The directory of the positions file will automatically be created on start up
# if it doesn't already exist..
[positions: <promtail.position_config>]

scrape_configs:
  - [<promtail.scrape_config>]

[target_config: <promtail.target_config>]

[limits_config: <promtail.limits_config>]
```
> **Note:** More information on the following types can be found on the
> documentation for Promtail:
>
> * [`promtail.client_config`](https://grafana.com/docs/loki/latest/clients/promtail/configuration/#clients)
> * [`promtail.scrape_config`](https://grafana.com/docs/loki/latest/clients/promtail/configuration/#scrape_configs)
> * [`promtail.target_config`](https://grafana.com/docs/loki/latest/clients/promtail/configuration/#target_config)
> * [`promtail.limits_config`](https://grafana.com/docs/loki/latest/clients/promtail/configuration/#limits_config)

> **Note:** Backticks in values are not supported.

> **Note:**  Because of how YAML treats backslashes in double-quoted strings,
> all backslashes in a regex expression must be escaped when using double
> quotes. But because of double processing, in Grafana Agent config file
> you must use quadruple backslash (`\\\\`) construction to add backslashes
> into regular expressions, here is example for `name=(\w+)\s` regex:
```
  selector: '{app="my-app"} |~ "name=(\\\\w+)\\\\s"'
```

Using single or double backslash construction produces the error:
```
failed to make file target manager: invalid match stage config: invalid selector syntax for match stage: parse error at line 1, col 40: literal not terminated
```
Using backticks produces the error:
```
invalid match stage config: invalid selector syntax for match stage: parse error at line 1, col 51: syntax error: unexpected IDENTIFIER, expecting STRING"
```

I feel this page could use a little bit of a revamp. It talks a lot about the "template" to be used in a Grafana agent configuration file for the logs section but does not actually has examples of a configuration file with a logs section.

On top of that, it feels a little misleading, when it says "logs_config" whereas in the configuration file, there is only the "logs" block. Simple stuff like this, is what we get asked in the field. Because many folks are using this technology for the first time and they are not super hands-on. This comment would be applicable for metrics and traces blocks too.

As per examples, including something like we have on this page - https://grafana.com/docs/grafana-cloud/data-configuration/logs/collect-logs-with-agent/ - would be really helpful.

Or I actually created a sample Grafana Agent config focusing on the logs block, see below -

# Sample config for Grafana Agent
# For a full configuration reference, see: https://grafana.com/docs/agent/latest/configuration/.
server:
  log_level: warn

metrics:
  global:
    scrape_interval: 1m
  wal_directory: '/var/lib/grafana-agent'
  configs:
    # Example Prometheus scrape configuration to scrape the agent itself for metrics.
    # This is not needed if the agent integration is enabled.
    # - name: agent
    #   host_filter: false
    #   scrape_configs:
    #     - job_name: agent
    #       static_configs:
    #         - targets: ['127.0.0.1:9090']

logs:
  configs:
  - name: nb-td-example
    positions:
      filename: /tmp/positions.yaml
    scrape_configs:
    - job_name: dmesg-example
      static_configs:
        - targets: [localhost]
          labels:
            job: dmesg
            __path__: /var/log/dmesg
    - job_name: varlogs-example
      static_configs:
        - targets: [localhost]
          labels:
            job: varlogs
            __path__: /var/log/*.log
    - job_name: nb-log-file-example
      static_configs:
        - targets: [localhost]
          labels:
            job: nb-sample
            __path__: /home/navish_bahl_grafana_com/sample-log/nb*
    clients:
      - url: <Your Grafana logs endpoint>
        basic_auth:
            username: <Your Grafana logs username>
            password: <API-Key-With-Metrics-Publisher-Role>

integrations:
  agent:
    enabled: true
  node_exporter:
    enabled: true
    include_exporter_metrics: true
    disable_collectors:
      - "mdadm"
  prometheus_remote_write:
  - url: <<Your Grafana metricss endpoint>>
    basic_auth:
      username: <Your Grafana metrics username>
      password: <API-Key-With-Metrics-Publisher-Role>

Something like this would make the docs much more practical to use for the end customer.
