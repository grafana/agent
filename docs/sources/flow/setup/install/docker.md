---
description: Learn how to install Grafana Agent in flow mode on Docker
title: Install Grafana Agent in flow mode on Docker
menuTitle: Docker
weight: 100
aliases:
 - ../../install/docker/
---

# Install Grafana Agent in flow mode on Docker

Grafana Agent is available as Docker images on the following platforms:

* [Linux containers][] for AMD64 and ARM64 machines.
* [Windows containers][] for AMD64 machines.

[Linux containers]: #run-a-linux-docker-container
[Windows containers]: #run-a-windows-docker-container

## Before you begin

* Ensure that [Docker][] is installed and running on your machine.

* Ensure that you have an existing Grafana Agent River configuration file on your host system, such as:

  ```river
  logging {
    level  = "info"
    format = "logfmt"
  }
  ```

[Docker]: https://docker.io

## Run a Linux Docker container

To run Grafana Agent in flow mode as a Linux Docker container, perform the following
steps:

1. Run the following command in a terminal:

   ```shell
   docker run \
     -e AGENT_MODE=flow \
     -v CONFIG_FILE_PATH:/etc/agent/config.river \
     -p 12345:12345 \
     grafana/agent:latest \
       run --server.http.listen-addr=0.0.0.0:12345 /etc/agent/config.river
   ```

   Replace `CONFIG_FILE_PATH` with the path of the configuration file on
   your host system.

The last line may be modified to change the arguments passed to the Grafana
Agent binary. To see the set of options available to the `run` command,
refer to the documentation for [run][].

{{% admonition type="note" %}}
Make sure to pass `--server.http.listen-addr=0.0.0.0:12345` as an
argument like in the example above, otherwise the [debugging UI][] won't be
available outside of the Docker container.
{{% /admonition %}}

[debugging UI]: {{< relref "../../monitoring/debugging.md#grafana-agent-flow-ui" >}}
[run]: {{< relref "../../reference/cli/run.md" >}}

## Run a Windows Docker container

To run Grafana Agent in flow mode as a Windows Docker container, perform the following
steps:

1. Run the following command in a terminal:

   ```shell
   docker run \
     -e AGENT_MODE=flow \
     -v CONFIG_FILE_PATH:C:\etc\grafana-agent\config.river \
     -p 12345:12345 \
     grafana/agent:latest-windows \
       run --server.http.listen-addr=0.0.0.0:12345 C:\etc\grafana-agent\config.river
   ```

   Replace `CONFIG_FILE_PATH` with the path of the configuration file on
   your host system.

The last line may be modified to change the arguments passed to the Grafana
Agent binary. To see the set of options available to the `run` command,
refer to the documentation for [run][].

{{% admonition type="note" %}}
Make sure to pass `--server.http.listen-addr=0.0.0.0:12345` as an
argument like in the example above, otherwise the [debugging UI][] won't be
available outside of the Docker container.
{{% /admonition %}}

[debugging UI]: {{< relref "../../monitoring/debugging.md#grafana-agent-flow-ui" >}}
[run]: {{< relref "../../reference/cli/run.md" >}}

## Result

After following these steps, Grafana Agent should be successfully running
in Docker.

To validate that Grafana Agent is running successfully, navigate to
<http://localhost:12345> to ensure that the Grafana Agent [UI][] loads
without error.

[UI]: {{< relref "../../monitoring/debugging.md#grafana-agent-flow-ui" >}}
