---
description: Learn how to install Grafana Agent in flow mode on Docker
title: Run Grafana Agent in flow mode in a Docker container
menuTitle: Docker
weight: 100
aliases:
 - ../../install/docker/
---

# Run Grafana Agent in flow mode in a Docker container

Grafana Agent is available as a Docker container image on the following platforms:

* [Linux containers][] for AMD64 and ARM64.
* [Windows containers][] for AMD64.

[Linux containers]: #run-a-linux-docker-container
[Windows containers]: #run-a-windows-docker-container

## Before you begin

* Install [Docker][] on your computer.
* Create and save a Grafana Agent River configuration file on your computer, for example:

  ```river
  logging {
    level  = "info"
    format = "logfmt"
  }
  ```

[Docker]: https://docker.io

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
Refer to the documentation for [run][] for more information about the options available to the `run` command.

{{% admonition type="note" %}}
Make sure you pass `--server.http.listen-addr=0.0.0.0:12345` as an argument as shown in the example above.
If you don't pass this argument, the [debugging UI](../../monitoring/debugging.md#grafana-agent-flow-ui) won't be available outside of the Docker container.
{{% /admonition %}}

[run]: {{< relref "../../reference/cli/run.md" >}}

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
Refer to the documentation for [run][] for more information about the options available to the `run` command.

{{% admonition type="note" %}}
Make sure you pass `--server.http.listen-addr=0.0.0.0:12345` as an argument as shown in the example above.
If you don't pass this argument, the [debugging UI](../../monitoring/debugging.md#grafana-agent-flow-ui) won't be available outside of the Docker container.
{{% /admonition %}}

[run]: {{< relref "../../reference/cli/run.md" >}}

## Verify

To verify that Grafana Agent is running successfully, navigate to <http://localhost:12345> and make sure the Grafana Agent [UI][] loads without error.

[UI]: {{< relref "../../monitoring/debugging.md#grafana-agent-flow-ui" >}}
