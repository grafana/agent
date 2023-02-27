---
title: Agent Operator integrations
weight: 100
---
# Agent Operator integrations

This topic provides examples of Agent Operator integrations, including [node_exporter](#agent-operator-nodeexporter-integration) and [mysqld_exporter](#agent-operator-mysqldexporter-integration). 

## Agent Operator node_exporter integration

The Agent Operator node_exporter integration lets you monitor your hardware and OS metrics from Unix-based machines, including Linux machines.

The following YAML file causes Agent Operator to create an instance of a grafana-agent-integrations-deploy resource that exports Node metrics.

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
 
## Agent Operator mysqld_exporter integration

The Agent Operator mysqld_exporter integration is an embedded version of mysqld_exporter that lets you collect metrics from MySQL servers.

The following YAML file causes Agent Operator to create an instance of a grafana-agent-integrations-deploy resource that exports MySQL metrics.

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
