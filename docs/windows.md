# Windows Installation

## Overview

The installer will install Grafana Agent into the default directory `C:\Program Files (x86)\Grafana Agent`. [Windows Exporter](https://github.com/grafana/windows_exporter) can be enabled with all default options. 

## Installation

After installation ensure that you can reach `http://localhost:12345/-/healthy` and `http://localhost:12345/agent/api/v1/targets`. 

If Grafana Agent is re-installed and an agent-config.yaml already exists it will not overwrite the existing one.

After installation, you can adjust `C:\Program Files (x86)\Grafana Agent\agent-config.yaml` to meet your specific needs. After changing the configuration file, the Grafana Agent service must be restarted to load changes to the configuration.

### Silent Installation

Silent installation can be achieved via  `grafana-agent-installer.exe /S  /EnableExporter "true"`. EnableExporter enables or disables Windows Exporter, default is `false`.

## Security

A configuration file for the Agent is provided by default at `C:\Program Files (x86)\Grafana Agent`. Depending on your configuration, you may wish to modify the default permissions of the file or move it to another directory. 

When changing the location of the configuration file, you must update the Grafana Agent service to load the new path. Run the following in an elevated prompt, replacing `<new_path>` with the full path holding `agent-config.yaml`:

```
sc config "Grafana Agent" binpath= "<installed_directory>\agent-windows-amd64.exe -config.file=\"<new_path>\agent-config.yaml\""
```

## Uninstall

Via Remove Programs or uninstaller.exe in the directory the Agent is installed in. This will turn off and remove the Agent then delete any installed files in the applications directory.

## Logs

Logs are currently not stored anywhere for the services version of Agent, in the future logs will be available via Event Viewer.
