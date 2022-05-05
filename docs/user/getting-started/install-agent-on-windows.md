+++
title = "Install Agent on Windows"
weight = 120
+++

# Install Agent on Windows

The installer will install Grafana Agent into the default directory `C:\Program Files\Grafana Agent`.
The [windows_exporter integration](https://github.com/prometheus-community/windows_exporter)
can be enabled with all default windows_exporter options. The installer can be
found by either:

1. Downloading the `grafana-agent-installer.exe` asset from the current [release](https://github.com/grafana/agent/releases/latest) page.
2. Downloading the `grafana-agent-installer.exe` asset directly from https://github.com/grafana/agent/releases/latest/download/grafana-agent-installer.exe.

## Installation

After installation, ensure that you can reach `http://localhost:12345/-/healthy` and `http://localhost:12345/agent/api/v1/metrics/targets`.

After installation, you can adjust `C:\Program Files\Grafana Agent\agent-config.yaml` to meet your specific needs. After changing the configuration file, the Grafana Agent service must be restarted to load changes to the configuration.

Existing configuration files will be kept when re-installing or upgrading the Grafana Agent.

### Silent Installation

Silent installation can be achieved via  `grafana-agent-installer.exe /S /EnableExporter true /Username xyz /Password password /Url "http://example.com" `. EnableExporter enables or disables Windows Exporter, default is `false`. Username, Password and Url set the global remote_write configuration, they can be omitted if not using remote_write. When using powershell use triple quotes `"""http://example.com"""` around the url parameter.

## Security

A configuration file for the Agent is provided by default at `C:\Program Files\Grafana Agent`. Depending on your configuration, you may wish to modify the default permissions of the file or move it to another directory.

When changing the location of the configuration file, you must update the Grafana Agent service to load the new path. Run the following in an elevated prompt, replacing `<new_path>` with the full path holding `agent-config.yaml`:

```
sc config "Grafana Agent" binpath= "<installed_directory>\agent-windows-amd64.exe -config.file=\"<new_path>\agent-config.yaml\""
```

## Uninstall

If the Grafana Agent is installed using the installer, it can be uninstalled via Windows' Remove Programs or `C:\Program Files\Grafana Agent\uninstaller.exe`. Uninstalling the Agent will stop the service and remove it from disk. This will include any configuration files in the installation directory. Grafana Agent can be silently uninstalled by executing `uninstall.exe /S` while running as Administrator.

## Logs

When Grafana Agent runs as a Windows Service, the Grafana Agent will write logs to Windows Event Logs. When running as executable, Grafana Agent will write to standard out. The logs will be written with the event source name of `Grafana Agent`.

## Pushing Windows logs to Grafana Loki

Grafana Agent can use the embedded [promtail](https://grafana.com/docs/loki/latest/clients/promtail/) to push Windows Event Logs to [Grafana Loki](https://github.com/grafana/loki). Example configuration below:

```yaml
server:
  log_level: debug
logs:
  # Choose a directory to save the last read position of log files at.
  # This directory will be created if it doesn't already exist.
  positions_directory: "C:\\path\\to\\directory"
  configs:
    - name: windows
      # Loki endpoint to push logs to
      clients:
        - url: https://example.com
      scrape_configs:
      - job_name: windows
        windows_events:
          # Note the directory structure must already exist but the file will be created on demand
          bookmark_path: "C:\\path\\to\\bookmark\\directory\\bookmark.xml"
          use_incoming_timestamp: false
          eventlog_name: "Application"
          # Filter for logs
          xpath_query: '*'
          labels:
            job: windows
```

Additional windows_events configuration details can be found [here](https://grafana.com/docs/loki/latest/clients/promtail/configuration/#windows_events).
