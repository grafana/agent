---
aliases:
- /docs/grafana-cloud/agent/operator/operator-integrations/
- /docs/grafana-cloud/monitor-infrastructure/agent/operator/operator-integrations/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/operator/operator-integrations/
- /docs/grafana-cloud/send-data/agent/operator/operator-integrations/
canonical: https://grafana.com/docs/agent/latest/operator/operator-integrations/
description: Learn how to set up integrations
title: Set up integrations
weight: 350
---
# Set up integrations

This topic provides examples of setting up Grafana Agent Operator integrations, including [node_exporter](#set-up-an-agent-operator-node_exporter-integration) and [mysqld_exporter](#set-up-an-agent-operator-mysqld_exporter-integration).

## Before you begin

Before you begin, make sure that you have deployed the Grafana Agent Operator CRDs and installed Agent Operator into your cluster. See [Install Grafana Agent Operator with Helm]({{< relref "./helm-getting-started.md" >}}) or [Install Grafana Agent Operator]({{< relref "./getting-started.md" >}}) for instructions.

Also, make sure that you [deploy the GrafanaAgent resource]({{< relref "./deploy-agent-operator-resources.md" >}}) and the `yaml` you use has the `integrations` definition under `spec`.

**Important:** The field `name` under the `spec` section of the manifest must contain the name of the integration to be installed according to the list of integrations defined [here]({{< relref "../static/configuration/integrations/integrations-next/_index.md#config-changes" >}}).

**Important:** The value of the `metrics_instance` field needs to be in the format `<namespace>/<name>`, with namespace and name matching the values defined in the `metadata` section from the `MetricsInstance` resource as explained in [deploy a MetricsInstance resource]({{< relref "./deploy-agent-operator-resources.md#deploy-a-metricsinstance-resource" >}})

## Set up an Agent Operator node_exporter integration

The Agent Operator node_exporter integration lets you monitor your hardware and OS metrics from Unix-based machines, including Linux machines.

To set up a node_exporter integration:

1. Copy the following manifest to a file:

    ```yaml
    apiVersion: monitoring.grafana.com/v1alpha1
    kind: Integration
    metadata:
     name: node-exporter
     namespace: default
     labels:
       agent: grafana-agent-integrations
    spec:
     name: node_exporter
     type:
       allNodes: true
       unique: true
     config:
       autoscrape:
         enable: true
         metrics_instance: default/primary
       rootfs_path: /default/node_exporter/rootfs
       sysfs_path: /default/node_exporter/sys
       procfs_path: /default/node_exporter/proc
     volumeMounts:
       - mountPath: /default/node_exporter/proc
         name: proc
       - mountPath: /default/node_exporter/sys
         name: sys
       - mountPath: /default/node_exporter/rootfs
         name: root
     volumes:
       - name: proc
         hostPath:
           path: /proc
       - name: sys
         hostPath:
           path: /sys
       - name: root
         hostPath:
           path: /root
    ```

2. Customize the manifest as needed and roll it out to your cluster using `kubectl apply -f` followed by the filename.

    The manifest causes Agent Operator to create an instance of a grafana-agent-integrations-deploy resource that exports Node metrics.

## Set up an Agent Operator mysqld_exporter integration

The Agent Operator mysqld_exporter integration is an embedded version of mysqld_exporter that lets you collect metrics from MySQL servers.

To set up a mysqld_exporter integration:

1. Copy the following manifest to a file:

    ```yaml
    apiVersion: monitoring.grafana.com/v1alpha1
    kind: Integration
    metadata:
     name: mysqld-exporter
     namespace: default
     labels:
       agent: grafana-agent-integrations
    spec:
     name: mysql
     type:
       allNodes: true
       unique: true
     config:
       autoscrape:
         enable: true
         metrics_instance: default/primary
       data_source_name: root@(server-a:3306)/
    ```

2. Customize the manifest as needed and roll it out to your cluster using `kubectl apply -f` followed by the filename.

    The manifest causes Agent Operator to create an instance of a grafana-agent-integrations-deploy resource that exports MySQL metrics.
