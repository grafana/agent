---
title: Install Grafana Agent in static mode on Docker
menuTitle: Docker
weight: 200
aliases:
- ../../set-up/install-agent-docker/
- ../install-agent-docker/
---

# Run Grafana Agent in a Docker container

Grafana Agent is available as a Docker container image on the following platforms:

* [Linux containers][] for AMD64 and ARM64 computers.
* [Windows containers][] for AMD64 computers.

[Linux containers]: #run-a-linux-docker-container
[Windows containers]: #run-a-windows-docker-container

## Before you begin

* Install [Docker][] on your computer.
* Create and save a Grafana Agent YAML [configuration file]({{< relref "../../configuration/create-config-file/" >}}) on youur computer.

[Docker]: https://docker.io

## Run a Linux Docker container

To run a Grafana Agent Docker container on Linux, run the following command in a terminal window:

```shell
docker run \
  -v WAL_DATA_DIRECTORY:/etc/agent/data \
  -v CONFIG_FILE_PATH:/etc/agent/agent.yaml \
  grafana/agent:v0.34.3
```

  Replace `CONFIG_FILE_PATH` with the configuration file path on your Linux host system.

{{% admonition type="note" %}}
For the flags to work correctly, you must expose the paths on your Linux host to the Docker container through a bind mount.
{{%/admonition %}}

## Run a Windows Docker container

To run a Grafana Agent Docker container on Windows, run the following command in a Windows command prompt:

```shell
docker run ^
  -v WAL_DATA_DIRECTORY:c:\etc\grafana-agent\data ^
  -v CONFIG_FILE_PATH:c:\etc\grafana-agent ^
  grafana/agent:v0.34.3-windows
```

Replace the following:

* `CONFIG_FILE_PATH` with the configuration file path on your Windows host system.

* `WAL_DATA_DIRECTORY` with the directory used to store your metrics before sending them to Prometheus. Old WAL data is cleaned up every hour and is used for recovery if the process crashes.

{{% admonition type="note" %}}
For the flags to work correctly, you must expose the paths on your Windows host to the Docker container through a bind mount.
{{%/admonition %}}

## Next steps

- [Start Grafana Agent]({{< relref "../start-agent/" >}})
- [Configure Grafana Agent]({{< relref "../../configuration/" >}})
