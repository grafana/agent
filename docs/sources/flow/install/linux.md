---
title: Linux
weight: 300
---

# Install Grafana Agent Flow on Linux systems

You can install Grafana Agent Flow as a systemd service on Linux.

## Install on Debian or Ubuntu

To install Grafana Agent Flow on Debian or Ubuntu, run the following commands in a terminal window.

1. Import the GPG key and add the Grafana package repository:

   ```shell
    sudo mkdir -p /etc/apt/keyrings/
    wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg > /dev/null
    echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list
   ```

1. Update the repositories:

   ```shell
   sudo apt-get update
   ```

1. Install Grafana Agent Flow:

   ```shell
   sudo apt-get install grafana-agent-flow
   ```

### Uninstall on Debian or Ubuntu

To uninstall Grafana Agent Flow on Debian or Ubuntu, run the following commands in a terminal window.

1. Stop the systemd service for Grafana Agent Flow:
   
   ```shell
   sudo systemctl stop grafana-agent-flow
   ```

1. Uninstall Grafana Agent Flow:

   ```shell
   sudo apt remove grafana-agent-flow
   ```

1. Optional. Remove the Grafana repository:

   ```shell
   sudo rm -i /etc/apt/sources.list.d/grafana.list
   ```

## Install on RHEL or Fedora

To install Grafana Agent Flow on RHEL or Fedora, run the following commands in a terminal window.

1. Import the GPG key:

   ```shell
   wget -q -O gpg.key https://rpm.grafana.com/gpg.key
   sudo rpm --import gpg.key
   ```

1. Create `/etc/yum.repos.d/grafana.repo` with the following content:

   ```shell
   [grafana]
   name=grafana
   baseurl=https://rpm.grafana.com
   repo_gpgcheck=1
   enabled=1
   gpgcheck=1
   gpgkey=https://rpm.grafana.com/gpg.key
   sslverify=1
   sslcacert=/etc/pki/tls/certs/ca-bundle.crt
   ```

1. Optional. Verify the Grafana repository configuration:

   ```shell
   cat /etc/yum.repos.d/grafana.repo
   ```

1. Install Grafana Agent Flow:

   ```shell
   sudo dnf install grafana-agent-flow
   ```

### Uninstall on RHEL or Fedora

To uninstall Grafana Agent Flow on RHEL or Fedora, run the following commands in a terminal window:

1. Stop the systemd service for Grafana Agent Flow:
   
   ```shell
   sudo systemctl stop grafana-agent-flow
   ```

1. Uninstall Grafana Agent Flow:

   ```shell
   sudo dnf remove grafana-agent-flow
   ```

1. Optional. Remove the Grafana repository:

   ```shell
   sudo rm -i /etc/yum.repos.d/rpm.grafana.repo
   ```

## Install on SUSE or openSUSE

To install Grafana Agent Flow on SUSE or openSUSE, run the following commands in a terminal window.

1. Import the GPG key and add the Grafana package repository:

   ```shell
   wget -q -O gpg.key https://rpm.grafana.com/gpg.key
   sudo rpm --import gpg.key
   sudo zypper addrepo https://rpm.grafana.com grafana
   ```

1. Update the repositories:

   ```shell
   sudo zypper update
   ```

1. Install Grafana Agent Flow:

   ```shell
   sudo zypper install grafana-agent-flow
   ```

### Uninstall on SUSE or openSUSE

To uninstall Grafana Agent Flow on SUSE or openSUSE, run the following commands in a terminal:

1. Stop the systemd service for Grafana Agent Flow:
   
   ```shell
   sudo systemctl stop grafana-agent-flow
   ```

1. Uninstall Grafana Agent Flow:

   ```shell
   sudo zypper remove grafana-agent-flow
   ```

1. Optional. Remove the Grafana repository:

   ```shell
   sudo zypper removerepo grafana
   ```

## Operation guide

Grafana Agent Flow is configured as a [systemd][] service.

[systemd]: https://systemd.io/

### Start Grafana Agent Flow

To start Grafana Agent Flow, run the following command in a terminal:

```shell
sudo systemctl start grafana-agent-flow
```

To check the status of Grafana Agent Flow, run the following command in a terminal:

```shell
sudo systemctl status grafana-agent-flow
```

### Run Grafana Agent Flow on startup

To automatically run Grafana Agent Flow when the system starts, run the following command in a terminal:

```shell
sudo systemctl enable grafana-agent-flow.service
```

### Configuring Grafana Agent Flow

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

### Passing additional command-line flags

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

### Exposing the UI to other machines

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

### Viewing Grafana Agent Flow logs

Logs of Grafana Agent Flow can be found by running the following command in a
terminal:

```shell
sudo journalctl -u grafana-agent-flow
```
