---
aliases:
- /docs/grafana-cloud/agent/flow/get-started/install/ansible/
canonical: https://grafana.com/docs/agent/latest/flow/get-started/install/ansible/
description: Learn how to install Grafana Agent Flow with Ansible
menuTitle: Ansible
title: Install Grafana Agent Flow with Ansible
weight: 550
---

# Install or uninstall {{% param "PRODUCT_NAME" %}} using Ansible

You can use Ansible to install and manage {{< param "PRODUCT_NAME" >}}.

## Before you begin

- These steps assume you already have a working [Ansible](https://www.ansible.com/) setup,
and a pre-existing inventory.
- You can add the tasks below to any new or existing Role you choose.
- These tasks install {{< param "PRODUCT_NAME" >}} from the package repositories. They expect to target Linux systems using 
APT or YUM package managers.

## Steps

To add {{% param "PRODUCT_NAME" %}} to a host:

1. Add these tasks to your playbook to add the Grafana package repositories to your system:
    ```yaml
    - name: "Install DEB repo"
      when:
        - "ansible_pkg_mgr == 'apt'"
      block:
        - name: "Import Grafana APT GPG key"
          ansible.builtin.get_url:
            url: "https://apt.grafana.com/gpg.key"
            dest: /etc/apt/keyrings/grafana.gpg
            mode: "0644"

        - name: "Add Grafana APT repository"
          ansible.builtin.apt_repository:
            repo: "deb [signed-by=/usr/share/keyrings/grafana.key] https://apt.grafana.com/ stable main"
            state: present
            update_cache: true
    - name: "Install yum repo"
      when:
        - "ansible_pkg_mgr in ['yum', 'dnf']"
      block:
        - name: "Add Grafana YUM/DNF repository"
          ansible.builtin.yum_repository:
            name: grafana
            description: grafana
            baseurl: "https://packages.grafana.com/oss/rpm"
            enabled: true
            gpgkey: "https://packages.grafana.com/gpg.key"
            repo_gpgcheck: true
            gpgcheck: true
    ```
1. Add the following tasks to install and enable the `grafana-agent-flow` service:
    ```yaml
    - name: Install grafana-agent-flow
      ansible.builtin.package:
        name: grafana-agent-flow
        state: present

    - name: Enable grafana-agent-flow service
      ansible.builtin.service:
        name: grafana-agent-flow
        enabled: yes
        state: started
    ```

## Configuration

The `grafana-agent-flow` package installs a default configuration file that doesn't send telemetry anywhere.

The default configuration file location is `/etc/grafana-agent-flow.river`. You can replace this file with your own configuration, or create a new configuration file for the service to use. 

## Next steps

- [Configure {{< param "PRODUCT_NAME" >}}][Configure]

{{% docs/reference %}}

[Configure]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/tasks/configure/configure-linux.md"
[Configure]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/tasks/configure/configure-linux.md"
{{% /docs/reference %}}