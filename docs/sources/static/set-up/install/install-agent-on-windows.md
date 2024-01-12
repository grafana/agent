---
aliases:
- ../../set-up/install-agent-on-windows/
- ../set-up/install-agent-on-windows/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/set-up/install/install-agent-on-windows/
- /docs/grafana-cloud/send-data/agent/static/set-up/install/install-agent-on-windows/
canonical: https://grafana.com/docs/agent/latest/static/set-up/install/install-agent-on-windows/
description: Learn how to install Grafana Agent in static mode on Windows
menuTitle: Windows
title: Install Grafana Agent in static mode on Windows
weight: 600
---

# Install Grafana Agent in static mode on Windows

You can install Grafana Agent in static mode on Windows as a standard graphical install, or as a silent install.

## Standard install

To do a standard graphical install of Grafana Agent on Windows, perform the following steps.

1. Navigate to the [latest release](https://github.com/grafana/agent/releases) on GitHub.

1. Scroll down to the **Assets** section.

1. Download the file called `grafana-agent-installer.exe.zip`.

1. Unzip the downloaded file.

1. Double-click on `grafana-agent-installer.exe` to install Grafana Agent.

   Grafana Agent is installed into the default directory `C:\Program Files\Grafana Agent`.
   
   The following options are available:
   
   - The [windows_exporter integration][windows_exporter_config] can be enabled with all default windows_exporter options. 
   - The [-config.expand-env][flags] command line flag can be enabled.

## Silent install

To do a silent install of Grafana Agent on Windows, perform the following steps.

1. Navigate to the [latest release](https://github.com/grafana/agent/releases) on GitHub.

1. Scroll down to the **Assets** section.

1. Download the file called `grafana-agent-installer.exe.zip`.

1. Unzip the downloaded file.

1. Run the following command in PowerShell or Command Prompt:

   ```shell
   PATH_TO_INSTALLER/grafana-agent-installer.exe /S
   ```

   Replace `PATH_TO_INSTALLER` with the path where the unzipped installer executable is located.

## Silent install with `remote_write`

If you are using `remote_write` you must enable Windows Exporter and set the global remote_write configuration.

1. Navigate to the [latest release](https://github.com/grafana/agent/releases) on GitHub.

1. Scroll down to the **Assets** section.

1. Download the file called `grafana-agent-installer.exe.zip`.

1. Unzip the downloaded file.

1. Run the following command in PowerShell or Command Prompt:

   ```shell
   PATH_TO_INSTALLER/grafana-agent-installer.exe /S /EnableExporter true /Username USERNAME /Password PASSWORD /Url "http://example.com"
   ```

   Replace the following:

   - `PATH_TO_INSTALLER`: The path where the unzipped installer executable is located.
   - `USERNAME`: Your username
   - `PASSWORD`: Your password

   If you are using Powershell, make sure you use triple quotes `"""http://example.com"""` around the URL parameter.

## Silent install with `-config.expand_env`

You can enable [-config.expand-env][flags] during a silent install.

1. Navigate to the [latest release](https://github.com/grafana/agent/releases) on GitHub.

1. Scroll down to the **Assets** section.

1. Download the file called `grafana-agent-installer.exe.zip`.

1. Unzip the downloaded file.

1. Run the following command in PowerShell or Command Prompt:

   ```shell
   PATH_TO_INSTALLER/grafana-agent-installer.exe /S /ExpandEnv true
   ```

## Verify the installation

1. Make sure you can access `http://localhost:12345/-/healthy` and `http://localhost:12345/agent/api/v1/metrics/targets`.

1. Optional: You can adjust `C:\Program Files\Grafana Agent\agent-config.yaml` to meet your specific needs. After changing the configuration file, restart the Grafana Agent service to load changes to the configuration.

Existing configuration files are kept when re-installing or upgrading the Grafana Agent.

## Security

A configuration file for Grafana Agent is provided by default at `C:\Program Files\Grafana Agent`. Depending on your configuration, you can modify the default permissions of the file or move it to another directory.

If you change the location of the configuration file, do the following step.

1. Update the Grafana Agent service to load the new path.

1. Run the following with Administrator privileges in PowerShell or Command Prompt:

   ```shell
   sc config "Grafana Agent" binpath= "INSTALLED_DIRECTORY\agent-windows-amd64.exe -config.file=\"PATH_TO_CONFIG\agent-config.yaml\""
   ```

   Replace `PATH_TO_CONFIG` with the full path to your Grafana Agent configuratiuon file.

## Uninstall Grafana Agent

You can uninstall Grafana Agent with Windows Remove Programs or `C:\Program Files\Grafana Agent\uninstaller.exe`.
Uninstalling Grafana Agent will stop the service and remove it from disk. This includes any configuration files in the installation directory.

Grafana Agent can also be silently uninstalled by running `uninstall.exe /S` as Administrator.

## Push Windows logs to Grafana Loki

Grafana Agent can use the embedded [promtail](/docs/loki/latest/clients/promtail/) to push Windows Event Logs to [Grafana Loki](https://github.com/grafana/loki). Example configuration below:

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

Refer to [windows_events](/docs/loki/latest/clients/promtail/configuration/#windows_events) for additional configuration details.

## Next steps

- [Start Grafana Agent][start]
- [Configure Grafana Agent][configure]

{{% docs/reference %}}
[flags]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/flags"
[windows_exporter_config]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/integrations/windows-exporter-config"
[start]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/set-up/start-agent"
[start]: "/docs/grafana-cloud/ -> ../start-agent"
[configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/create-config-file"
[configure]: "/docs/grafana-cloud/ -> ../../configuration/create-config-file"
{{% /docs/reference %}}
