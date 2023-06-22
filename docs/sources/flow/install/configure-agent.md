---
title: Configure
weight: 800
---

## Configure Grafana Agent Flow on Linux

To configure Grafana Agent Flow when installed on Linux, perform the following steps:

1. Edit the default configuration file at `/etc/grafana-agent-flow.river`.

2. Run the following command in a terminal to reload the configuration file:

   ```shell
   sudo systemctl reload grafana-agent-flow
   ```

To change the configuration file used by the service, perform the following steps:

1. Edit the environment file for the service:

   * Debian-based systems: edit `/etc/default/grafana-agent-flow`
   * RedHat or SUSE-based systems: edit `/etc/sysconfig/grafana-agent-flow`

2. Change the contents of the `CONFIG_FILE` environment variable to point to
   the new configuration file to use.

3. Restart the Grafana Agent Flow service:

   ```shell
   sudo systemctl restart grafana-agent-flow
   ```

### Pass additional command-line flags

By default, the Grafana Agent Flow service launches with the [run][]
command, passing the following flags:

* `--storage.path=/var/lib/grafana-agent-flow`

To pass additional command-line flags to the Grafana Agent Flow binary, perform
the following steps:

1. Edit the environment file for the service:

   * Debian-based systems: edit `/etc/default/grafana-agent-flow`
   * RedHat or SUSE-based systems: edit `/etc/sysconfig/grafana-agent-flow`

2. Change the contents of the `CUSTOM_ARGS` environment variable to specify
   command-line flags to pass.

3. Restart the Grafana Agent Flow service:

   ```shell
   sudo systemctl restart grafana-agent-flow
   ```

To see the list of valid command-line flags that can be passed to the service,
refer to the documentation for the [run][] command.

[run]: {{< relref "../reference/cli/run.md" >}}

### Expose the UI to other machines

By default, Grafana Agent Flow listens on the local network for its HTTP
server. This prevents other machines on the network from being able to access
the [UI for debugging][UI].

To expose the UI to other machines, complete the following steps:

1. Follow [Passing additional command-line flags](#passing-additional-command-line-flags)
   to edit command line flags passed to Grafana Agent Flow, including the
   following customizations:

    1. Add the following command line argument to `CUSTOM_ARGS`:

       ```
       --server.http.listen-addr=LISTEN_ADDR:12345
       ```

       Replace `LISTEN_ADDR` with an address which other machines on the
       network have access to, like the network IP address of the machine
       Grafana Agent Flow is running on.

       To listen on all interfaces, replace `LISTEN_ADDR` with `0.0.0.0`.

[UI]: {{< relref "../monitoring/debugging.md#grafana-agent-flow-ui" >}}

## Configure Grafana Agent Flow on macOS

To configure Grafana Agent Flow on macOS, perform the following
steps:

1. Edit the default configuration file at
   `$(brew --prefix)/etc/grafana-agent-flow/config.river`.

2. Run the following command in a terminal to restart the Grafana Agent Flow
   service:

   ```shell
   brew services restart grafana-agent-flow
   ```

### Configure the Grafana Agent Flow service

> **NOTE**: Due to limitations in Homebrew, customizing the service used by
> Grafana Agent Flow on macOS requires changing the Homebrew formula and
> reinstalling Grafana Agent Flow.

To customize the Grafana Agent Flow service on macOS, perform the following
steps:

1. Run the following command in a terminal:

   ```shell
   brew edit grafana-agent-flow
   ```

   This will open the Grafana Agent Flow Homebrew Formula in an editor.

2. Modify the `service` section as desired to change things such as:

   * The configuration file used by Grafana Agent Flow.
   * Flags passed to the Grafana Agent Flow binary.
   * Location of log files.

   When done, save the resulting formula file.

3. Reinstall the Grafana Agent Flow Formula by running the following command in
   a terminal:

   ```shell
   brew reinstall grafana-agent-flow
   ```

4. Restart the Grafana Agent Flow service by running the command in a terminal:

   ```shell
   brew services restart grafana-agent-flow
   ```

### Expose the UI to other machines

By default, Grafana Agent Flow listens on the local network for its HTTP
server. This prevents other machines on the network from being able to access
the [UI for debugging][UI].

To expose the UI to other machines, complete the following steps:

1. Follow [Configuring the Grafana Agent Flow service](#configuring-the-grafana-agent-flow-service)
   to edit command line flags passed to Grafana Agent Flow, including the
   following customizations:

    1. Modify the line inside the `service` block containing
       `--server.http.listen-addr=127.0.0.1:12345`, replacing `127.0.0.1` with
       the address which other machines on the network have access to, like the
       network IP address of the machine Grafana Agent Flow is running on.

       To listen on all interfaces, replace `127.0.0.1` with `0.0.0.0`.

[UI]: {{< relref "../monitoring/debugging.md#grafana-agent-flow-ui" >}}

## Configure Grafana Agent Flow on Windows

To configure Grafana Agent Flow when installed on Windows, perform the following
steps:

1. Edit the default configuration file at `C:\Program Files\Grafana Agent
   Flow\config.river`.
1. Restart the Grafana Agent Flow service:
   1. Open the Windows Services manager (`services.msc`):
      1. Right click on the Start Menu icon.
      1. Click on **Run**.
      1. In the resulting dialog box, type `services.msc`.
      1. Click **OK**.
   1. Right click on the service called "Grafana Agent Flow".
   1. In the resulting dialog menu, click on All Tasks > Restart.

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
   1. Click on **Run**.
   1. In the resulting dialog box, type `regedit`.
   1. Click **OK**.

1. Navigate to the key at the path `HKEY_LOCAL_MACHINE\SOFTWARE\Grafana\Grafana
   Agent Flow`.
1. Double-click on the value called "Arguments".
1. In the resulting dialog box, enter the new set of arguments to pass to the
   Grafana Agent Flow binary.
1. Restart the Grafana Agent Flow service:
   1. Open the Windows Services manager (`services.msc`):
      1. Right click on the Start Menu icon.
      1. Click on **Run**.
      1. In the resulting dialog box, type `services.msc`.
      1. Click **OK**.
   1. Right click on the service called "Grafana Agent Flow".
   1. In the resulting dialog menu, click on All Tasks > Restart.

### Expose the UI to other machines

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
