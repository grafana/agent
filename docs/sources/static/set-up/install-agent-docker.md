---
title: Run Grafana Agent on Docker
weight: 110
aliases:
- ../../set-up/install-agent-docker/
---

## Run Grafana Agent on Docker

Install Grafana Agent and get it up and running on Docker.

### Before you begin

 - Ensure that you have Docker installed.
 - Ensure that you have created a configuration file. In the case of Docker, you install and run the Grafana Agent with a single command. You therefore need to create a configuration file before running Grafana Agent on Docker. For more information on creating a configuration file, refer to [Create a configuration file]({{< relref "../configuration/create-config-file/" >}}).

### Run Agent in a Linux Container

1. Copy and paste the following commands into your command line.
```
docker run \
  -v /tmp/agent:/etc/agent/data \
  -v /path/to/config.yaml:/etc/agent/agent.yaml \
  grafana/agent:v0.32.1
```

2. Replace `/tmp/agent` with the folder you want to store WAL data in.

    WAL data is where metrics are stored before they are sent to Prometheus. Old WAL data is cleaned up every hour and is used for recovery if the process happens to crash.

3. Replace `/path/to/config.yaml` with a path pointing to a valid configuration file.

Note that using paths on your host machine must be exposed to the Docker
container through a bind mount for the flags to work properly.

### Run Agent in a Windows Container

1. Copy and paste the following commands into your command line.
    ```
    docker run ^
      -v c:\grafana-agent-data:c:\etc\grafana-agent\data ^
      -v c:\workspace\config\grafana-agent:c:\etc\grafana-agent ^
      grafana/agent:v0.32.1-windows
    ```

2. Replace `c:\grafana-agent-data` with the folder you want to store WAL data in.

    WAL data is where metrics are stored before they are sent to Prometheus. Old WAL data is cleaned up every hour and is used for recovery if the process happens to crash.

3. Replace `c:\workspace\config\grafana-agent` with a path containing to a valid configuration file. The config file must be named grafana-agent.yaml inside the directory.

Note that using paths on your host machine must be exposed to the Docker
container through a bind mount for the flags to work properly.

### Result

Docker containers run the Grafana Agent using this configuration file.




