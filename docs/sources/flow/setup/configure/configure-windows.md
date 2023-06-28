---
description: Learn how to configure Grafana Agent on Windows
title: Configure Grafana Agent on Windows
menuTitle: Windows
weight: 500
---

# Configure Grafana Agent on Windows

To configure Grafana Agent when installed on Windows, perform the following
steps:

1. Edit the default configuration file at `C:\Program Files\Grafana Agent Flow\config.river`.

1. Restart the Grafana Agent service:

   1. Open the Windows Services manager (`services.msc`):

      1. Right click on the Start Menu and select **Run**.

      1. Type `services.msc` and click **OK**.

   1. Right click on the service called **Grafana Agent Flow**.
   
   1. Click on **All Tasks > Restart**.

## Change command-line arguments

By default, the Grafana Agent service will launch with passing the
following arguments to the Grafana Agent binary:

* `run`
* `C:\Program Files\Grafana Agent Flow\config.river`
* `--storage.path=C:\ProgramData\Grafana Agent Flow\data`

To change the set of command-line arguments passed to the Grafana Agent
binary, perform the following steps:

1. Open the Registry Editor:

   1. Right click on the Start Menu and select **Run**.

   1. Type `regedit` and click **OK**.

1. Navigate to the key at the path `HKEY_LOCAL_MACHINE\SOFTWARE\Grafana\Grafana Agent Flow`.

1. Double-click on the value called **Arguments***.

1. In the dialog box, enter the new set of arguments to pass to the Grafana Agent binary.

1. Restart the Grafana Agent service:

   1. Open the Windows Services manager (`services.msc`):

      1. Right click on the Start Menu and select **Run**.

      1. Type `services.msc` and click **OK**.

   1. Right click on the service called **Grafana Agent Flow**.

   1. Click on **All Tasks > Restart**.

## Expose the UI to other machines

By default, Grafana Agent listens on the local network for its HTTP
server. This prevents other machines on the network from being able to access
the [UI for debugging][UI].

To expose the UI to other machines, complete the following steps:

1. Follow [Change command-line arguments](#change-command-line-arguments)
   to edit command line flags passed to Grafana Agent, including the
   following customizations:

    1. Add the following command line argument:

       ```shell
       --server.http.listen-addr=LISTEN_ADDR:12345
       ```

       Replace `LISTEN_ADDR` with an address which other machines on the
       network have access to, like the network IP address of the machine
       Grafana Agent is running on.

       To listen on all interfaces, replace `LISTEN_ADDR` with `0.0.0.0`.

[UI]: {{< relref "../../monitoring/debugging.md#grafana-agent-flow-ui" >}}
