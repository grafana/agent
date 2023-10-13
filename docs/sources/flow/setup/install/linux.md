---
aliases:
- ../../install/linux/
- /docs/grafana-cloud/agent/flow/setup/install/linux/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/setup/install/linux/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/setup/install/linux/
canonical: https://grafana.com/docs/agent/latest/flow/setup/install/linux/
description: Learn how to install Grafana Agent in flow mode on Linux
menuTitle: Linux
title: Install or uninstall Grafana Agent in flow mode on Linux
weight: 300
---

# Install or uninstall Grafana Agent in flow mode on Linux

You can install Grafana Agent in flow mode as a systemd service on Linux.

## Install

To install Grafana Agent in flow mode on Linux, run the following commands in a terminal window.

1. Import the GPG key and add the Grafana package repository.

   {{< code >}}
   ```debian-ubuntu
   sudo mkdir -p /etc/apt/keyrings/
   wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg > /dev/null
   echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list
   ```

   ```rhel-fedora
   wget -q -O gpg.key https://rpm.grafana.com/gpg.key
   sudo rpm --import gpg.key
   sudo echo '[grafana]\nname=grafana\nbaseurl=https://rpm.grafana.com\nrepo_gpgcheck=1\nenabled=1\ngpgcheck=1\ngpgkey=https://rpm.grafana.com/gpg.key\nsslverify=1
sslcacert=/etc/pki/tls/certs/ca-bundle.crt' > /etc/yum.repos.d/grafana.repo
   ```

   ```suse-opensuse
   wget -q -O gpg.key https://rpm.grafana.com/gpg.key
   sudo rpm --import gpg.key
   sudo zypper addrepo https://rpm.grafana.com grafana
   ```
   {{< /code >}}

1. Update the repositories.

   {{< code >}}
   ```debian-ubuntu
   sudo apt-get update
   ```

   ```rhel-fedora
   yum update
   ```

   ```suse-opensuse
   sudo zypper update
   ```
   {{< /code >}}

1. Install Grafana Agent.

   {{< code >}}
   ```debian-ubuntu
   sudo apt-get install grafana-agent-flow
   ```

   ```rhel-fedora
   sudo dnf install grafana-agent-flow
   ```

   ```suse-opensuse
   sudo zypper install grafana-agent-flow
   ```
   {{< /code >}}

## Uninstall

To uninstall Grafana Agent on Linux, run the following commands in a terminal window.

1. Stop the systemd service for Grafana Agent.

   ```All-distros
   sudo systemctl stop grafana-agent-flow
   ```

1. Uninstall Grafana Agent.

   {{< code >}}
   ```debian-ubuntu
   sudo apt-get remove grafana-agent-flow
   ```

   ```rhel-fedora
   sudo dnf remove grafana-agent-flow
   ```

   ```suse-opensuse
   sudo zypper remove grafana-agent-flow
   ```
   {{< /code >}}

1. Optional: Remove the Grafana repository.

   {{< code >}}
   ```debian-ubuntu
   sudo rm -i /etc/apt/sources.list.d/grafana.list
   ```

   ```rhel-fedora
   sudo rm -i /etc/yum.repos.d/rpm.grafana.repo
   ```

   ```suse-opensuse
   sudo zypper removerepo grafana
   ```
   {{< /code >}}

## Next steps

- [Start Grafana Agent]({{< relref "../start-agent#linux" >}})
- [Configure Grafana Agent]({{< relref "../configure/configure-linux" >}})
