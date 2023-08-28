---
aliases:
- ../../install/linux/
canonical: https://grafana.com/docs/agent/latest/flow/setup/install/linux/
description: Learn how to install Grafana Agent in flow mode on Linux
menuTitle: Linux
title: Install Grafana Agent in flow mode on Linux
weight: 300
---

# Install Grafana Agent in flow mode on Linux

You can install Grafana Agent in flow mode as a systemd service on Linux.

## Install on Debian or Ubuntu

To install Grafana Agent in flow mode on Debian or Ubuntu, run the following commands in a terminal window.

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

1. Install Grafana Agent:

   ```shell
   sudo apt-get install grafana-agent-flow
   ```

### Uninstall on Debian or Ubuntu

To uninstall Grafana Agent on Debian or Ubuntu, run the following commands in a terminal window.

1. Stop the systemd service for Grafana Agent:

   ```shell
   sudo systemctl stop grafana-agent-flow
   ```

1. Uninstall Grafana Agent:

   ```shell
   sudo apt-get remove grafana-agent-flow
   ```

1. Optional: Remove the Grafana repository:

   ```shell
   sudo rm -i /etc/apt/sources.list.d/grafana.list
   ```

## Install on RHEL or Fedora

To install Grafana Agent in flow mode on RHEL or Fedora, run the following commands in a terminal window.

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

1. Optional: Verify the Grafana repository configuration:

   ```shell
   cat /etc/yum.repos.d/grafana.repo
   ```

1. Install Grafana Agent:

   ```shell
   sudo dnf install grafana-agent-flow
   ```

### Uninstall on RHEL or Fedora

To uninstall Grafana Agent on RHEL or Fedora, run the following commands in a terminal window:

1. Stop the systemd service for Grafana Agent:

   ```shell
   sudo systemctl stop grafana-agent-flow
   ```

1. Uninstall Grafana Agent:

   ```shell
   sudo dnf remove grafana-agent-flow
   ```

1. Optional: Remove the Grafana repository:

   ```shell
   sudo rm -i /etc/yum.repos.d/rpm.grafana.repo
   ```

## Install on SUSE or openSUSE

To install Grafana Agent in flow mode on SUSE or openSUSE, run the following commands in a terminal window.

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

1. Install Grafana Agent:

   ```shell
   sudo zypper install grafana-agent-flow
   ```

### Uninstall on SUSE or openSUSE

To uninstall Grafana Agent on SUSE or openSUSE, run the following commands in a terminal:

1. Stop the systemd service for Grafana Agent:

   ```shell
   sudo systemctl stop grafana-agent-flow
   ```

1. Uninstall Grafana Agent:

   ```shell
   sudo zypper remove grafana-agent-flow
   ```

1. Optional: Remove the Grafana repository:

   ```shell
   sudo zypper removerepo grafana
   ```

## Next steps

- [Start Grafana Agent]({{< relref "../start-agent#linux" >}})
- [Configure Grafana Agent]({{< relref "../configure/configure-linux" >}})
