+++
title = "logs_config"
weight = 300
aliases = ["/docs/agent/latest/configuration/loki-config/"]
+++

# logs_config

The `logs_config` block configures how the Agent collects logs and sends them to
a Loki push API endpoint. `logs_config` is identical to how Promtail is
configured, except deprecated fields have been removed and the server_config is
not supported.

Refer to the
[Promtail documentation](https://github.com/grafana/loki/tree/master/docs/sources/clients/promtail#client_config)
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

# Loki Promtail instances to run for log collection.
configs:
  - [<logs_instance_config>]
```

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
```

> **Note:** More information on the following types can be found on the
> documentation for Promtail:
>
> * [`promtail.client_config`](https://github.com/grafana/loki/tree/master/docs/sources/clients/promtail#client_config)
> * [`promtail.scrape_config`](https://github.com/grafana/loki/tree/master/docs/sources/clients/promtail#scrape_config)
> * [`promtail.target_config`](https://github.com/grafana/loki/tree/master/docs/sources/clients/promtail#target_config)
