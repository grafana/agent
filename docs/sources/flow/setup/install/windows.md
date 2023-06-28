---
description: Learn how to install Grafana Agent Flow on Windows
title: Install Grafana Agent Flow on Windows
menuTitle: Windows
weight: 400
aliases:
 - ../../install/windows/
---

# Install Grafana Agent Flow on Windows

You can install Grafana Agent Flow on Windows with the standard graphical installer, or as a silent install.

## Graphical install

1. Navigate to the [latest release][latest].

1. Scroll down to the **Assets** section.

1. Download the file called `grafana-agent-flow-installer.exe.zip`.

1. Unzip the downloaded file.

1. Double-click on the unzipped installer to run it.

[latest]: https://github.com/grafana/agent/releases/latest

## Silent install

1. Navigate to the [latest release][latest].

1. Scroll down to the **Assets** section.

1. Download the file called `grafana-agent-flow-installer.exe.zip`.

1. Unzip the downloaded file.

1. Run the following command in PowerShell or Command Prompt:

   ```shell
   PATH_TO_INSTALLER /S
   ```

   Replace `PATH_TO_INSTALLER` with the path where the unzipped installer
   executable is located, such as
   `C:\Users\Alexis\Downloads\grafana-agent-flow-installer.exe`.

[latest]: https://github.com/grafana/agent/releases/latest

## Uninstall

You can uninstall Grafana Agent Flow with Windows Remove Programs or `C:\Program Files\Grafana Agent\uninstaller.exe`. Uninstalling Grafana Agent Flow will stop the service and remove it from disk. This includes any configuration files in the installation directory. 

Grafana Agent Flow can also be silently uninstalled by running `uninstall.exe /S` as Administrator.
