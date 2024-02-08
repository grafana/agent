---
aliases:
- ../../set-up/install-agent-docker/
- ../set-up/install-agent-docker/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/set-up/install/install-agent-docker/
- /docs/grafana-cloud/send-data/agent/static/set-up/install/install-agent-docker/
canonical: https://grafana.com/docs/agent/latest/static/set-up/install/install-agent-docker/
description: Learn how to run Grafana Agent in static mode in a Docker container
menuTitle: Docker
title: Run Grafana Agent in static mode in a Docker container
weight: 200
---

# Run Grafana Agent in static mode in a Docker container

Grafana Agent is available as a Docker container image on the following platforms:

* [Linux containers][] for AMD64 and ARM64.
* [Windows containers][] for AMD64.

[Linux containers]: #run-a-linux-docker-container
[Windows containers]: #run-a-windows-docker-container

## Before you begin

* Install [Docker][] on your computer.
* Create and save a Grafana Agent YAML [configuration file][configure] on your computer.

[Docker]: https://docker.io

## Run a Linux Docker container

To run a Grafana Agent Docker container on Linux, run the following command in a terminal window:

```shell
docker run \
  -v WAL_DATA_DIRECTORY:/etc/agent/data \
  -v CONFIG_FILE_PATH:/etc/agent/agent.yaml \
  grafana/agent:{{< param "AGENT_RELEASE" >}}
```

Replace `CONFIG_FILE_PATH` with the configuration file path on your Linux host system.

{{< admonition type="note" >}}
For the flags to work correctly, you must expose the paths on your Linux host to the Docker container through a bind mount.
{{< /admonition >}}

## Run a Windows Docker container

To run a Grafana Agent Docker container on Windows, run the following command in a Windows command prompt:

```shell
docker run ^
  -v WAL_DATA_DIRECTORY:C:\etc\grafana-agent\data ^
  -v CONFIG_FILE_PATH:C:\etc\grafana-agent ^
  grafana/agent:{{< param "AGENT_RELEASE" >}}-windows
```

Replace the following:

* `CONFIG_FILE_PATH`: The configuration file path on your Windows host system.
* `WAL_DATA_DIRECTORY`: the directory used to store your metrics before sending them to Prometheus. Old WAL data is cleaned up every hour and is used for recovery if the process crashes.

{{< admonition type="note" >}}
For the flags to work correctly, you must expose the paths on your Windows host to the Docker container through a bind mount.
{{< /admonition >}}

## Next steps

- [Start Grafana Agent][start]
- [Configure Grafana Agent][configure]

{{% docs/reference %}}
[start]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/set-up/start-agent"
[start]: "/docs/grafana-cloud/ -> ../start-agent"
[configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/create-config-file"
[configure]: "/docs/grafana-cloud/ -> ../../configuration/create-config-file"
{{% /docs/reference %}}
