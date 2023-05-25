---
title: Docker
weight: 110
aliases:
- ../../set-up/install-agent-docker/
---

## Install Grafana Agent on Docker

Grafana Agent Flow is available as Docker images on the following platforms:

* [Linux containers][] for AMD64 and ARM64 machines.
* [Windows containers][] for AMD64 machines.

[Linux containers]: #run-agent-in-a-linux-container
[Windows containers]: #run-agent-in-a-windows-container

### Before you begin

 - Ensure that [Docker][] is installed and running on your machine.
 - Ensure that you have an existing Grafana Agent configuration file. 
   You start Grafana Agent on Docker with a single command. You must [create a configuration file]({{< relref "../configuration/create-config-file/" >}}) before you start the Grafana Agent on Docker.

[Docker]: https://docker.io
### Run Agent in a Linux Container

1. Run the following command in a terminal:

   ```
   docker run \
     -v WAL_DATA_DIRECTORY:/etc/agent/data \
     -v CONFIG_FILE_PATH:/etc/agent/agent.yaml \
     grafana/agent:v0.33.2
   ```
   
   - Replace `CONFIG_FILE_PATH` with the configuration file path on your host system.
   - Replace `WAL_DATA_DIRECTORY` with the directory where your metrics are stored before they are sent to Prometheus. Old WAL data is cleaned up every hour and is used for recovery if the process crashes.

     {{% admonition type="note" %}}The paths on your host machine must be exposed to the Docker container through a bind mount for the flags to work properly.{{%/admonition %}}

### Run Agent in a Windows Container

1. Copy and paste the following commands into your command line.

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

### Result

Docker containers run the Grafana Agent using this configuration file.
