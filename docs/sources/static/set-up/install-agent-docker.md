---
title: Docker
weight: 110
aliases:
- ../../set-up/install-agent-docker/
---

# Run Grafana Agent in a Docker container

Grafana Agent is available as a Docker container image on the following platforms:

* [Linux][]
* [Windows][]

[Linux]: #run-a-linux-docker-container
[Windows]: #run-a-windows-docker-container

## Before you begin

 - Ensure that [Docker][] is installed and running on your machine.
 - Ensure that you have created a Grafana Agent [configuration file]({{< relref "../configuration/create-config-file/" >}}).

[Docker]: https://docker.io
## Run a Linux Docker container

1. To run a Grafana Agent Docker container on Linux, run the following command in a terminal:

   ```
   docker run \
     -v WAL_DATA_DIRECTORY:/etc/agent/data \
     -v CONFIG_FILE_PATH:/etc/agent/agent.yaml \
     grafana/agent:v0.34.0
   ```
   
   - Replace `CONFIG_FILE_PATH` with the configuration file path on your Linux host system.
   - Replace `WAL_DATA_DIRECTORY` with the directory used to store your metrics before sending them to Prometheus. Old WAL data is cleaned up every hour and is used for recovery if the process crashes.

     {{% admonition type="note" %}}
     For the flags to work correctly, you must expose the paths on your Linux host to the Docker container through a bind mount.
     {{%/admonition %}}

## Run a Windows Docker container

1. To run a Grafana Agent Docker container on Windows, run the following command in a Windows command prompt:

   ```
   docker run ^
     -v WAL_DATA_DIRECTORY:c:\etc\grafana-agent\data ^
     -v CONFIG_FILE_PATH:c:\etc\grafana-agent ^
     grafana/agent:v0.34.0-windows
   ```

   - Replace `CONFIG_FILE_PATH` with the configuration file path on your Windows host system.
   - Replace `WAL_DATA_DIRECTORY` with the directory used to store your metrics before sending them to Prometheus. Old WAL data is cleaned up every hour and is used for recovery if the process crashes.

   {{% admonition type="note" %}}
   For the flags to work correctly, you must expose the paths on your Windows host to the Docker container through a bind mount. 
   {{%/admonition %}}

## Result

Docker containers run the Grafana Agent using this configuration file.
