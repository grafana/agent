---
aliases:
- ../../install/docker/
- /docs/grafana-cloud/agent/flow/setup/install/docker/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/setup/install/docker/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/setup/install/docker/
canonical: https://grafana.com/docs/agent/latest/flow/setup/install/docker/
description: Learn how to install Grafana Agent in flow mode on Docker
menuTitle: Docker
title: Run Grafana Agent in flow mode in a Docker container
weight: 100
refs:
  run:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/cli/run/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/cli/run/
  ui:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/monitoring/debugging/#grafana-agent-flow-ui
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/monitoring/debugging/#grafana-agent-flow-ui
---

# Run Grafana Agent in flow mode in a Docker container

Grafana Agent is available as a Docker container image on the following platforms:

* [Linux containers][] for AMD64 and ARM64.
* [Windows containers][] for AMD64.

## Before you begin

* Install [Docker][] on your computer.
* Create and save a Grafana Agent River configuration file on your computer, for example:

  ```river
  logging {
    level  = "info"
    format = "logfmt"
  }
  ```

## Run a Linux Docker container

To run Grafana Agent in flow mode as a Linux Docker container, run the following command in a terminal window:

```shell
docker run \
  -e AGENT_MODE=flow \
  -v CONFIG_FILE_PATH:/etc/agent/config.river \
  -p 12345:12345 \
  grafana/agent:latest \
    run --server.http.listen-addr=0.0.0.0:12345 /etc/agent/config.river
```

Replace `CONFIG_FILE_PATH` with the path of the configuration file on your host system.

You can modify the last line to change the arguments passed to the Grafana Agent binary.
Refer to the documentation for [run](ref:run) for more information about the options available to the `run` command.

> **Note:** Make sure you pass `--server.http.listen-addr=0.0.0.0:12345` as an argument as shown in the example above.
> If you don't pass this argument, the [debugging UI](ref:ui) won't be available outside of the Docker container.


## Run a Windows Docker container

To run Grafana Agent in flow mode as a Windows Docker container, run the following command in a terminal window:

```shell
docker run \
  -e AGENT_MODE=flow \
  -v CONFIG_FILE_PATH:C:\etc\grafana-agent\config.river \
  -p 12345:12345 \
  grafana/agent:latest-windows \
    run --server.http.listen-addr=0.0.0.0:12345 C:\etc\grafana-agent\config.river
```

Replace `CONFIG_FILE_PATH` with the path of the configuration file on your host system.

You can modify the last line to change the arguments passed to the Grafana Agent binary.
Refer to the documentation for [run](ref:run) for more information about the options available to the `run` command.


> **Note:** Make sure you pass `--server.http.listen-addr=0.0.0.0:12345` as an argument as shown in the example above.
> If you don't pass this argument, the [debugging UI](ref:ui) won't be available outside of the Docker container.

## Verify

To verify that Grafana Agent is running successfully, navigate to <http://localhost:12345> and make sure the [Grafana Agent UI](ref:ui) loads without error.

[Linux containers]: #run-a-linux-docker-container
[Windows containers]: #run-a-windows-docker-container
[Docker]: https://docker.io

