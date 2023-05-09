---
title: Windows
weight: 400
---

# Install Grafana Agent Flow on Windows

Grafana Agent Flow can be installed on AMD64-based Windows machines.

## Graphical install

1. Navigate to the [latest release][latest].
2. Scroll down to the **Assets** section.
3. Download the file called `grafana-agent-flow-installer.exe.zip`.
4. Unzip the downloaded file.
5. Double-click on the unzipped installer to run it.

[latest]: https://github.com/grafana/agent/releases/latest

## Silent install

1. Navigate to the [latest release][latest].

2. Scroll down to the **Assets** section.

3. Download the file called `grafana-agent-flow-installer.exe.zip`.

4. Unzip the downloaded file.

5. Run the following command in PowerShell or Command Prompt:

   ```shell
   PATH_TO_INSTALLER /S
   ```

   1. Replace `PATH_TO_INSTALLER` with the path where the unzipped installer
      executable is located, such as
      `C:\Users\Alexis\Downloads\grafana-agent-flow-installer.exe`.

[latest]: https://github.com/grafana/agent/releases/latest

## Operation guide

After installing Grafana Agent Flow on Windows, it will be exposed as a Windows
Service, where it automatically runs on startup.

### Configuring Grafana Agent Flow

To configure Grafana Agent Flow when installed on Windows, perform the following
steps:

1. Edit the default configuration file at `C:\Program Files\Grafana Agent
   Flow\config.river`.

2. Restart the Grafana Agent Flow service:

   1. Open the Windows Services manager (`services.msc`):

      1. Right click on the Start Menu icon.

      2. Click on **Run**.

      3. In the resulting dialog box, type `services.msc`.

      4. Click **OK**.

   2. Right click on the service called "Grafana Agent Flow".

   3. In the resulting dialog menu, click on All Tasks > Restart.

### Change command-line arguments

By default, the Grafana Agent Flow service will launch with passing the
following arguments to the Grafana Agent Flow binary:

* `run`
* `C:\Program Files\Grafana Agent Flow\config.river`
* `--storage.path=C:\ProgramData\Grafana Agent Flow\data`

To change the set of command-line arguments passed to the Grafana Agent Flow
binary, perform the following steps:

1. Open the Registry Editor:

   1. Right click on the Start Menu icon.

   2. Click on **Run**.

   3. In the resulting dialog box, type `regedit`.

   4. Click **OK**.

2. Navigate to the key at the path `HKEY_LOCAL_MACHINE\SOFTWARE\Grafana\Grafana
   Agent Flow`.

3. Double-click on the value called "Arguments".

4. In the resulting dialog box, enter the new set of arguments to pass to the
   Grafana Agent Flow binary.

5. Restart the Grafana Agent Flow service:

   1. Open the Windows Services manager (`services.msc`):

      1. Right click on the Start Menu icon.

      2. Click on **Run**.

      3. In the resulting dialog box, type `services.msc`.

      4. Click **OK**.

   2. Right click on the service called "Grafana Agent Flow".

   3. In the resulting dialog menu, click on All Tasks > Restart.

### Exposing the UI to other machines

By default, Grafana Agent Flow listens on the local network for its HTTP
server. This prevents other machines on the network from being able to access
the [UI for debugging][UI].

To expose the UI to other machines, complete the following steps:

1. Follow [Change command-line arguments](#change-command-line-arguments)
   to edit command line flags passed to Grafana Agent Flow, including the
   following customizations:

    1. Add the following command line argument:

       ```
       --server.http.listen-addr=LISTEN_ADDR:12345
       ```

       Replace `LISTEN_ADDR` with an address which other machines on the
       network have access to, like the network IP address of the machine
       Grafana Agent Flow is running on.

       To listen on all interfaces, replace `LISTEN_ADDR` with `0.0.0.0`.

[UI]: {{< relref "../monitoring/debugging.md#grafana-agent-flow-ui" >}}

### Viewing Grafana Agent Flow logs

When running on Windows, Grafana Agent Flow writes its logs to Windows Event
Logs with an event source name of "Grafana Agent Flow".

To view the logs, perform the following steps:

1. Open the Event Viewer:

   1. Right click on the Start Menu icon.

   2. Click on **Run**.

   3. In the resulting dialog box, type `eventvwr`.

   4. Click **OK**.

2. In the Event Viewer, click on Windows Logs > Application.

3. Search for events with the source called "Grafana Agent Flow."
