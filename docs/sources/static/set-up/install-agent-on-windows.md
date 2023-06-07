---
title: Windows
weight: 120
aliases:
- ../../set-up/install-agent-on-windows/
---

# Install static mode on Windows

You can install Grafana Agent on Microsoft Windows.

## Download Grafana Agent for Windows

1. Navigate to the [latest release](https://github.com/grafana/agent/releases).
1. Scroll down to the **Assets** section.
1. Download the file called `grafana-agent-installer.exe.zip`.

   You can also download the `grafana-agent-installer.exe.zip` asset directly from https://github.com/grafana/agent/releases/latest/download/grafana-agent-installer.exe.zip

## Standard installation

1. Unzip the downloaded file.
1. Double-click on `grafana-agent-installer.exe` to install Grafana Agent.

   Grafana Agent is installed into the default directory `C:\Program Files\Grafana Agent`.
   The [windows_exporter integration](https://grafana.com/docs/agent/latest/static/configuration/integrations/windows-exporter-config) can be enabled with all default windows_exporter options.

## Silent installation

1. Unzip the downloaded file.
1. Run the following command in PowerShell or Command Prompt:

   ```shell
   PATH_TO_INSTALLER/grafana-agent-flow-installer.exe /S
   ```
## Silent installation with `remote_write`

If you are using `remote_write` you must enable Windows Exporter and set the global remote_write configuration. 

1. Unzip the downloaded file.
1. Run the following command in PowerShell or Command Prompt:

   ```shell
   PATH_TO_INSTALLER/grafana-agent-flow-installer.exe /S /EnableExporter true /Username <username> /Password <password> /Url "http://example.com"
   ```

   If you are using Powershell, make sure you use triple quotes `"""http://example.com"""` around the URL parameter.

## Verify the installation

1. Make sure you can access `http://localhost:12345/-/healthy` and `http://localhost:12345/agent/api/v1/metrics/targets`.
1. Optional: You can adjust `C:\Program Files\Grafana Agent\agent-config.yaml` to meet your specific needs. After changing the configuration file, restart the Grafana Agent service to load changes to the configuration.

Existing configuration files are kept when re-installing or upgrading the Grafana Agent.

## Security

A configuration file for the Grafana Agent is provided by default at `C:\Program Files\Grafana Agent`. Depending on your configuration, you can modify the default permissions of the file or move it to another directory.

If you change the location of the configuration file, ensure you complete the following steps.

1. Update the Grafana Agent service to load the new path.
1. Run the following in an elevated prompt, replacing `<new_path>` with the full path holding `agent-config.yaml`:

```shell
sc config "Grafana Agent" binpath= "INSTALLED_DIRECTORY\agent-windows-amd64.exe -config.file=\"<new_path>\agent-config.yaml\""
```

## Uninstall Grafana Agent

If you installed Grafana Agent using the Windows installer, you can uninstall it using Windows' Remove Programs or `C:\Program Files\Grafana Agent\uninstaller.exe`.
Uninstalling Grafana Agent will stop the service and remove it from disk. This includes any configuration files in the installation directory.
Grafana Agent can also be silently uninstalled by executing `uninstall.exe /S` as Administrator.

## Logs

When Grafana Agent runs as a Windows Service, it writes logs to Windows Event Logs. When running as executable, Grafana Agent will write to standard out. The logs will be written with the event source name of `Grafana Agent`.

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

Refer to [windows_events](https://grafana.com/docs/loki/latest/clients/promtail/configuration/#windows_events) for additional configuration details.
