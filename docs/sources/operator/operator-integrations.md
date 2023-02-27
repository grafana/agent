---
title: Set up Agent Operator integrations
weight: 350
---
# Set up Agent Operator integrations

This topic provides examples of setting up Agent Operator integrations, including [node_exporter](#set-up-an-agent-operator-node_exporter-integration) and [mysqld_exporter](#set-up-an-agent-operator-mysqld_exporter-integration). 

## Set up an Agent Operator node_exporter integration

The Agent Operator node_exporter integration lets you monitor your hardware and OS metrics from Unix-based machines, including Linux machines.

To set up a node_exporter integration:

1. Copy the following manifest to a file:  

    ```yaml
    # Collect node_exporter metrics
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

1. Customize the manifest as needed and roll it out to your cluster using `kubectl apply -f` followed by the filename.

    The manifest causes Agent Operator to create an instance of a grafana-agent-integrations-deploy resource that exports Node metrics.

## Set up an Agent Operator mysqld_exporter integration

The Agent Operator mysqld_exporter integration is an embedded version of mysqld_exporter that lets you collect metrics from MySQL servers.

To set up a mysqld_exporter integration:

1. Copy the following manifest to a file: 

    ```yaml
    # Collect mysqld_exporter metrics
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

1. Customize the manifest as needed and roll it out to your cluster using `kubectl apply -f` followed by the filename.

    The manifest causes Agent Operator to create an instance of a grafana-agent-integrations-deploy resource that exports MySQL metrics.