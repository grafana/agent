---
aliases:
- ../../set-up/install-agent-linux/
- ../set-up/install-agent-linux/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/set-up/install/install-agent-linux/
- /docs/grafana-cloud/send-data/agent/static/set-up/install/install-agent-linux/
canonical: https://grafana.com/docs/agent/latest/static/set-up/install/install-agent-linux/
description: Learn how to install Grafana Agent in static mode on Linux
menuTitle: Linux
title: Install Grafana Agent in static mode on Linux
weight: 400
---

# Install Grafana Agent in static mode on Linux

You can install Grafana Agent in static mode on Linux.

## Install on Debian or Ubuntu

To install Grafana Agent in static mode on Debian or Ubuntu, run the following commands in a terminal window.

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
   sudo apt-get install grafana-agent
   ```

### Uninstall on Debian or Ubuntu

To uninstall Grafana Agent on Debian or Ubuntu, run the following commands in a terminal window.

1. Stop the systemd service for Grafana Agent:

   ```shell
   sudo systemctl stop grafana-agent
   ```

1. Uninstall Grafana Agent:

   ```shell
   sudo apt-get remove grafana-agent
   ```

1. Optional: Remove the Grafana repository:

   ```shell
   sudo rm -i /etc/apt/sources.list.d/grafana.list
   ```

## Install on RHEL or Fedora

To install Grafana Agent in static mode on RHEL or Fedora, run the following commands in a terminal window.

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
   sudo dnf install grafana-agent
   ```

### Uninstall on RHEL or Fedora

To uninstall Grafana Agent on RHEL or Fedora, run the following commands in a terminal window:

1. Stop the systemd service for Grafana Agent:

   ```shell
   sudo systemctl stop grafana-agent
   ```

1. Uninstall Grafana Agent:

   ```shell
   sudo dnf remove grafana-agent
   ```

1. Optional: Remove the Grafana repository:

   ```shell
   sudo rm -i /etc/yum.repos.d/rpm.grafana.repo
   ```

## Install on SUSE or openSUSE

To install Grafana Agent in static mode on SUSE or openSUSE, run the following commands in a terminal window.

1. Import the GPG key and add the Grafana package repository:

   ```shell
   wget -q -O gpg.key https://apt.grafana.com/gpg.key
   sudo rpm --import gpg.key
   sudo zypper addrepo https://rpm.grafana.com grafana
   ```

1. Update the repositories:

   ```shell
   sudo zypper update
   ```

1. Install Grafana Agent:

   ```shell
   sudo zypper install grafana-agent
   ```

### Uninstall on SUSE or openSUSE

To uninstall Grafana Agent on SUSE or openSUSE, run the following commands in a terminal:

1. Stop the systemd service for Grafana Agent:

   ```shell
   sudo systemctl stop grafana-agent
   ````

1. Uninstall Grafana Agent:

   ```shell
   sudo zypper remove grafana-agent
   ```

1. Optional: Remove the Grafana repository:

   ```shell
   sudo zypper removerepo grafana
   ```

## Operation guide

The Grafana Agent is configured as a [systemd](https://systemd.io/) service.

### Start the Agent

To run Grafana Agent, run the following in a terminal:

   ```shell
   sudo systemctl start grafana-agent
   ```

To check the status of Grafana Agent, run the following command in a terminal:

   ```shell
   sudo systemctl status grafana-agent
   ```

### Run Grafana Agent on startup

To automatically run Grafana Agent when the system starts, run the following command in a terminal:

   ```shell
   sudo systemctl enable grafana-agent.service
   ```

### Configuring Grafana Agent

To configure Grafana Agent when installed on Linux, perform the following steps:

1. Edit the default configuration file at `/etc/grafana-agent.yaml`. 

1. Run the following command in a terminal to reload the configuration file:

   ```shell
   sudo systemctl reload grafana-agent
   ```

### View Grafana Agent logs

Logs of Grafana Agent can be found by running the following command in a terminal:

   ```shell
   sudo journalctl -u grafana-agent
   ```

## Next steps

- [Start Grafana Agent][start]
- [Configure Grafana Agent][configure]

{{% docs/reference %}}
[start]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/set-up/start-agent"
[start]: "/docs/grafana-cloud/ -> ../start-agent"
[configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/create-config-file"
[configure]: "/docs/grafana-cloud/ -> ../../configuration/create-config-file"
{{% /docs/reference %}}
