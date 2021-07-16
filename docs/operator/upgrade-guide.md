+++
title = "Upgrade guide"
weight = 200
+++

# Upgrade guide

This guide describes all breaking changes that have happened in prior releases
and how to migrate to newer versions of the Grafana Agent Operator. For
upgrading the Grafana Agent, please refer to its
[upgrade guide]({{< relref "../upgrade-guide" >}}) instead.

## Unreleased

These changes will come in a future version.

### RBAC additions for logging support

Now that the Grafana Agent Operator supports logs, the RBAC rules used by
the operator must be extended to LogsInstances, PodLogs, and DaemonSets.

Example new ClusterRole for the Operator to use:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: grafana-agent-operator
rules:
- apiGroups: [monitoring.grafana.com]
  resources:
  - grafana-agents
  - prometheus-instances
  - logs-instances
  - podlogs
  verbs: [get, list, watch]
- apiGroups: [monitoring.coreos.com]
  resources:
  - podmonitors
  - probes
  - servicemonitors
  verbs: [get, list, watch]
- apiGroups: [""]
  resources:
  - namespaces
  verbs: [get, list, watch]
- apiGroups: [""]
  resources:
  - secrets
  - services
  verbs: [get, list, watch, create, update, patch, delete]
- apiGroups: ["apps"]
  resources:
  - statefulsets
  - daemonsets
  verbs: [get, list, watch, create, update, patch, delete]
```

These RBAC permissions do not need to be given to the GrafanaAgent resource
itself, just the operator. The RBAC permissions recommended for the GrafanaAgent
resource for metrics already cover logging support.
