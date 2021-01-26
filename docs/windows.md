# Windows Installation

## Overview

The installer will install Grafana Agent into the default directory (C:\Program Files (x86)\Grafana Agent) and setting Agent as a service via [NSSM](https://nssm.cc/), optionally you can select to install the [Windows Exporter](https://github.com/prometheus-community/windows_exporter) with all default options. 

## Installation

![](./assets/remote_options.png)

**Remote Write** is used to specify any compatible Prometheus Endpoint. **User** and **Password** are used for basic auth. This will generate a configuration snippet like the below.

```
  prometheus_remote_write:
    - url: https://example.com
      basic_auth:
        username: "legit_username"
        password: "legit_password"
```

<br>

![](./assets/exporter.png)

Selecting the checkbox will install the Windows Exporter and start the Windows Exporter has a service, serving metrics from `localhost:9182/metrics`.

After installation ensure that you can reach `http://localhost:12345/-/healthy` and `http://localhost:12345/agent/api/v1/targets`. 

## Security

The config by default is installed alongside the Agent itself, depending on your configuration you may not want that for security reasons and instead make it protected. You can do that by changing the files attributes or changing the config via `nssm.exe" set "Grafana Agent" AppParameters "--config.file=\"CustomDirectory\agent-config.yaml\""`, then restarting the service.

## Uninstall

Via Remove Programs or uninstaller.exe in the directory the Agent is installed in. This will turn off and remove the Agent and Windows Export services then delete any installed files in the default directory.
