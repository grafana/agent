---
aliases:
  - ../../install/windows/
  - /docs/grafana-cloud/agent/flow/setup/install/windows/
  - /docs/grafana-cloud/monitor-infrastructure/agent/flow/setup/install/windows/
  - /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/setup/install/windows/
canonical: https://grafana.com/docs/agent/latest/flow/setup/install/windows/
description: Learn how to install Grafana Agent in flow mode on Windows
menuTitle: Windows
title: Install Grafana Agent in flow mode on Windows
weight: 500
refs:
  configure-grafana-agent:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/setup/configure/configure-windows/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/setup/configure/configure-windows/
  start-grafana-agent:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/setup/start-agent/#windows
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/setup/start-agent/#windows
  data-collection:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/data-collection/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/data-collection/
---

# Install Grafana Agent in flow mode on Windows

You can install Grafana Agent in flow mode on Windows as a standard graphical install, or as a silent install.

## Standard graphical install

To do a standard graphical install of Grafana Agent on Windows, perform the following steps.

1. Navigate to the [latest release][latest] on GitHub.

1. Scroll down to the **Assets** section.

1. Download the file called `grafana-agent-flow-installer.exe.zip`.

1. Unzip the downloaded file.

1. Double-click on `grafana-agent-installer.exe` to install Grafana Agent.

Grafana Agent is installed into the default directory `C:\Program Files\Grafana Agent Flow`.

## Silent install

To do a silent install of Grafana Agent on Windows, perform the following steps.

1. Navigate to the [latest release][latest] on GitHub.

1. Scroll down to the **Assets** section.

1. Download the file called `grafana-agent-flow-installer.exe.zip`.

1. Unzip the downloaded file.

1. Run the following command in PowerShell or Command Prompt:

   ```shell
   PATH_TO_INSTALLER /S
   ```

   Replace `PATH_TO_INSTALLER` with the path where the unzipped installer executable is located.

## Uninstall

You can uninstall Grafana Agent with Windows Remove Programs or `C:\Program Files\Grafana Agent\uninstaller.exe`. Uninstalling Grafana Agent stops the service and removes it from disk. This includes any configuration files in the installation directory.

Grafana Agent can also be silently uninstalled by running `uninstall.exe /S` as Administrator.

## Next steps

- [Start Grafana Agent](ref:start-grafana-agent)
- [Configure Grafana Agent](ref:configure-grafana-agent)

[latest]: https://github.com/grafana/agent/releases/latest

