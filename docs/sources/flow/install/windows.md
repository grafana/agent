---
title: Windows
weight: 400
---

# Install Grafana Agent Flow on Windows

You can install Grafana Agent Flow on Microsoft Windows.

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
