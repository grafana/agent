---
title: Docker
weight: 110
aliases:
- ../../set-up/install-agent-docker/
---

# Install Grafana Agent on Docker

Grafana Agent is available as a Docker image on the following platforms:

* [Linux][]
* [Windows][]

[Linux]: #run-a-linux-docker-container
[Windows]: #run-a-windows-docker-container

## Before you begin

 - Ensure that [Docker][] is installed and running on your machine.
 - Ensure that you have created a Grafana Agent [configuration file]({{< relref "../configuration/create-config-file/" >}}).

[Docker]: https://docker.io
## Run a Linux Docker container

1. To install a Grafana Agent Docker container on Linux, run the following command in a terminal:

   ```
   docker run \
     -v WAL_DATA_DIRECTORY:/etc/agent/data \
     -v CONFIG_FILE_PATH:/etc/agent/agent.yaml \
     grafana/agent:v0.33.2
   ```
   
   - Replace `CONFIG_FILE_PATH` with the configuration file path on your host system.
   - Replace `WAL_DATA_DIRECTORY` with the directory where your metrics are stored before they are sent to Prometheus. Old WAL data is cleaned up every hour and is used for recovery if the process crashes.

     {{% admonition type="note" %}}The paths on your host machine must be exposed to the Docker container through a bind mount for the flags to work properly.{{%/admonition %}}

## Run a Windows Docker container

1. To install a Grafana Agent Docker container onWindows, run the following command in a terminal:

   ```
   docker run ^
     -v c:\grafana-agent-data:c:\etc\grafana-agent\data ^
     -v CONFIG_FILE_PATH:c:\etc\grafana-agent ^
     grafana/agent:v0.33.2-windows
   ```

   - Replace `CONFIG_FILE_PATH` with the configuration file path on your host system.
   - Replace `WAL_DATA_DIRECTORY` with the directory where your metrics are stored before they are sent to Prometheus. Old WAL data is cleaned up every hour and is used for recovery if the process crashes.

   {{% admonition type="note" %}}
   The paths on your host machine must be exposed to the Docker container through a bind mount for the flags to work properly. 
   {{%/admonition %}}

## Result

Docker containers run the Grafana Agent using this configuration file.
