---
description: Learn how to install Grafana Agent in flow mode on Windows
title: Install Grafana Agent in flow mode on Windows
menuTitle: Windows
weight: 500
aliases:
 - ../../install/windows/
---

# Install Grafana Agent in flow mode on Windows

You can install Grafana Agent in flow mode on Windows as a standard install, or as a silent install.

## Standard install

To do a standard install of Grafana Agent on Windows, perform the following steps.

1. Navigate to the [latest release][latest] on GitHub.

1. Scroll down to the **Assets** section.

1. Download the file called `grafana-agent-flow-installer.exe.zip`.

1. Unzip the downloaded file.

1. Double-click on `grafana-agent-installer.exe` to install Grafana Agent.

Grafana Agent is installed into the default directory `C:\Program Files\Grafana Agent Flow`.

[latest]: https://github.com/grafana/agent/releases/latest

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

[latest]: https://github.com/grafana/agent/releases/latest

## Uninstall

You can uninstall Grafana Agent with Windows Remove Programs or `C:\Program Files\Grafana Agent\uninstaller.exe`. Uninstalling Grafana Agent will stop the service and remove it from disk. This includes any configuration files in the installation directory. 

Grafana Agent can also be silently uninstalled by running `uninstall.exe /S` as Administrator.

## Next steps

- [Start Grafana Agent]({{< relref "../start-agent#windows" >}})
- [Configure Grafana Agent]({{< relref "../configure/configure-windows" >}})
