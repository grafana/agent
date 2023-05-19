---
title: Linux
weight: 115
aliases:
- ../../set-up/install-agent-linux/
---

## Install static mode on Linux

Install Grafana Agent and get it up and running on Linux.

### Install on Debian or Ubuntu

1. Open a terminal and run the following command to install Grafana’s package repository:

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

### Install on RedHat, RHEL, or Fedora

1. Create `/etc/yum.repos.d/grafana.repo` with the following content:

   ```shell
   sudo nano /etc/yum.repos.d/grafana.repo
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

1. Verify that the repository is properly configured using `yum-config-manager`:

   ```shell
   yum-config-manager grafana
   ```

1. Install Grafana Agent:

   ```shell
   sudo yum install grafana-agent
   ```

### Install on SUSE or openSUSE

1. Open a terminal and run the following to install Grafana's package repository:

   ```shell
   wget -q -O gpg.key https://apt.grafana.com/gpg.key
   sudo rpm --import gpg.key
   sudo zypper addrepo https://rpm.grafana.com grafana
   ```

1. Update the repository and install Grafana Agent:

   ```shell
   sudo zypper update
   sudo zypper install grafana-agent
   ```

### Operation guide

The Grafana Agent is configured as a [systemd](https://systemd.io/) service.

#### Start the Agent

To run Grafana Agent, run the following in a terminal:

   ```shell
   sudo systemctl start grafana-agent
   ```

To check the status of Grafana Agent, run the following command in a terminal:

   ```shell
   sudo systemctl status grafana-agent
   ```

#### Run Grafana Agent on startup

To automatically run Grafana Agent Flow when the system starts, run the following command in a terminal:

   ```shell
   sudo systemctl enable grafana-agent.service
   ```

#### Configuring Grafana Agent

To configure Grafana Agent when installed on Linux, perform the following steps:

1. Edit the default configuration file at `/etc/grafana-agent.yaml`. 

1. Run the following command in a terminal to reload the configuration file:

   ```shell
   sudo systemctl reload grafana-agent
   ```

#### View Grafana Agent logs

Logs of Grafana Agent can be found by running the following command in a terminal:

   ```shell
   sudo journalctl -u grafana-agent
   ```
