---
title: Install Grafana Agent on Linux
weight: 115
---

## Install Grafana Agent on Linux

Install Grafana Agent and get it up and running on Linux.

### Install on Debian or Ubuntu

1.  If your distribution supports the signed-by option, open a terminal and enter:
```shell
$ mkdir -p /etc/apt/keyrings/
$ wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg
$ echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list
```
Otherwise, with the deprecated apt-key command:
```shell
$ echo "deb https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list
$ wget -q -O - https://apt.grafana.com/gpg.key | apt-key add -
```
2. After you add the repository, update package list:
```shell
sudo apt-get update
```
3. Install Grafana Agent:
```shell
sudo apt-get install grafana-agent
```
### Install on RPM-based Linux (CentOS, Fedora, OpenSuse, Red Hat)

1. Manually create a new `.repo` file inside `/etc/yum.repos.d` using a text editor:
```shell
$ sudo nano /etc/yum.repos.d/grafana.repo
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
2. Verify that the repository is properly configured using `yum-config-manager`:
```shell
yum-config-manager grafana
```
3. Install Grafana Agent:
```shell
sudo yum install grafana-agent
```

### Operation guide

The Grafana Agent will be configured a [systemd](https://systemd.io/) service after using the installation methods
explained in the previous sections.

#### Start the Agent

To run the service you just need to type:
```shell
sudo systemctl start grafana-agent
```

You can check the status of the running agent typing:
```shell
sudo systemctl status grafana-agent
```

Alternately, you can configure the Grafana Agent to restart at boot:
```shell
sudo systemctl enable grafana-agent.service
```

#### Editing the Agent's config file

By default, the config file is located in `/etc/grafana-agent.yaml`. After editing the file
with the desired config, you need to restart the agent running:
```shell
sudo systemctl restart grafana-agent
```

#### Check the logs of running Agent

You can check the logs of running agent typing:

```shell
sudo journalctl -u grafana-agent

Sep 02 14:33:28 grafana systemd[1]: Started Monitoring system and forwarder.
Sep 02 14:33:28 grafana grafana-agent[1633180]: ts=2022-09-02T12:33:28Z level=info caller=traces/traces.go:143 msg="Traces Logger I>
Sep 02 14:37:45 grafana systemd[1]: Stopping Monitoring system and forwarder...
```