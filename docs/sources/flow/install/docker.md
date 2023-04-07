---
title: Docker
weight: 200
---

# Install Grafana Agent Flow on Docker

Grafana Agent Flow is available as Docker containers on the following
platforms:

* [Linux containers][] for AMD64 and ARM64 machines.
* [Windows containers][] for AMD64 machines.

[Linux containers]: #run-a-linux-docker-container
[Windows containers]: #run-a-windows-docker-container

## Before you begin

* Ensure that [Docker][] is installed and running on your machine.

* Ensure that you have an existing Grafana Agent Flow configuration file to
  use saved on your host system, such as:

  ```river
  logging {
    level  = "info"
    format = "logfmt"
  }
  ```

[Docker]: https://docker.io

## Run a Linux Docker container

To run Grafana Agent Flow as a Linux Docker container, perform the following
steps:

1. Run the following command in a terminal:

   ```shell
   docker run \
     -e AGENT_MODE=flow \
     -v CONFIG_FILE_PATH:/etc/agent/config.river \
     -p 12345:12345 \
     grafana/agent:latest \
       run --server.http.listen-addr=0.0.0.0:12345 --config.file=/etc/agent/config.river
   ```

   Replace `CONFIG_FILE_PATH` with the path of the configuration file to use on
   your host system.

The last line may be modified to change the arguments passed to the Grafana
Agent Flow binary. To see the set of options available to the `run` command,
refer to the documentation for [run][].

[run]: {{< relref "../reference/cli/run.md" >}}

## Run a Windows Docker container

To run Grafana Agent Flow as a Windows Docker container, perform the following
steps:

1. Run the following command in a terminal:

   ```shell
   docker run \
     -e AGENT_MODE=flow \
     -v CONFIG_FILE_PATH:C:\etc\grafana-agent\config.river \
     -p 12345:12345 \
     grafana/agent:latest-windows \
       run --server.http.listen-addr=0.0.0.0:12345 --config.file=C:\etc\grafana-agent\config.river
   ```

   Replace `CONFIG_FILE_PATH` with the path of the configuration file to use on
   your host system.

The last line may be modified to change the arguments passed to the Grafana
Agent Flow binary. To see the set of options available to the `run` command,
refer to the documentation for [run][].

[run]: {{< relref "../reference/cli/run.md" >}}

## Result

After following these steps, Grafana Agent Flow should be successfully running
in Docker.

To validate that Grafana Agent Flow is running successfully, navigate to
<http://localhost:12345> to ensure that the Grafana Agent Flow [UI][] loads
without error.

[UI]: {{< relref "../monitoring/debugging.md#grafana-agent-flow-ui" >}}

