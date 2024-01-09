---
aliases:
- /docs/grafana-cloud/agent/operator/architecture/
- /docs/grafana-cloud/monitor-infrastructure/agent/operator/architecture/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/operator/architecture/
- /docs/grafana-cloud/send-data/agent/operator/architecture/
canonical: https://grafana.com/docs/agent/latest/operator/architecture/
description: Learn about Grafana Agent architecture
title: Architecture
weight: 300
---

# Architecture

Grafana Agent Operator works by watching for Kubernetes [custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) that specify how to collect telemetry data from your Kubernetes cluster and where to send it. Agent Operator manages corresponding Grafana Agent deployments in your cluster by watching for changes against the custom resources.

Grafana Agent Operator works in two phases&mdash;it discovers a hierarchy of custom resources and it reconciles that hierarchy into a Grafana Agent deployment.

## Custom resource hierarchy

The root of the custom resource hierarchy is the `GrafanaAgent` resource&mdash;the primary resource Agent Operator looks for. `GrafanaAgent` is called the _root_ because it
discovers other sub-resources, `MetricsInstance` and `LogsInstance`. The `GrafanaAgent` resource endows them with Pod attributes defined in the GrafanaAgent specification, for example, Pod requests, limits, affinities, and tolerations, and defines the Grafana Agent image. You can only define Pod attributes at the `GrafanaAgent` level. They are propagated to MetricsInstance and LogsInstance Pods.

The full hierarchy of custom resources is as follows:

- `GrafanaAgent`
    - `MetricsInstance`
        - `PodMonitor`
        - `Probe`
        - `ServiceMonitor`
    - `LogsInstance`
        - `PodLogs`

The following table describes these custom resources:

| Custom resource | description |
|---|---|
| `GrafanaAgent` | Discovers one or more `MetricsInstance` and `LogsInstance` resources. |
| `MetricsInstance` | Defines where to ship collected metrics. This rolls out a Grafana Agent StatefulSet that will scrape and ship metrics to a `remote_write` endpoint. |
| `ServiceMonitor` | Collects [cAdvisor](https://github.com/google/cadvisor) and [kubelet metrics](https://github.com/kubernetes/kube-state-metrics). This configures the `MetricsInstance` / Agent StatefulSet |
| `LogsInstance` | Defines where to ship collected logs. This rolls out a Grafana Agent [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) that will tail log files on your cluster nodes. |
| `PodLogs` | Collects container logs from Kubernetes Pods. This configures the `LogsInstance` / Agent DaemonSet. |

Most of the Grafana Agent Operator resources have the ability to reference a [ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/) or a
[Secret](https://kubernetes.io/docs/concepts/configuration/secret/). All referenced ConfigMaps or Secrets are added into the resource
hierarchy.

When a hierarchy is established, each item is watched for changes. Any changed
item causes a reconcile of the root `GrafanaAgent` resource, either
creating, modifying, or deleting the corresponding Grafana Agent deployment.

A single resource can belong to multiple hierarchies. For example, if two
`GrafanaAgents` use the same Probe, modifying that Probe causes both
`GrafanaAgents` to be reconciled.

To set up monitoring, Grafana Agent Operator works in the following two phases:

- Builds (discovers) a hierarchy of custom resources.
- Reconciles that hierarchy into a Grafana Agent deployment.

Agent Operator also performs [sharding and replication](#sharding-and-replication) and adds [labels](#added-labels) to every metric.

## How Agent Operator builds the custom resource hierarchy

Grafana Agent Operator builds the hierarchy using label matching on the custom resources. The following figure illustrates the matching. The `GrafanaAgent` picks up the `MetricsInstance`
and `LogsInstance` that match the label `instance: primary`. The instances pick up the resources the same way.

{{<figure class="float-right" src="../../assets/hierarchy.svg" >}}

### To validate the Secrets

The generated configurations are saved in Secrets. To download and
validate them manually, use the following commands:

```
$ kubectl get secrets <???>-logs-config -o json | jq -r '.data."agent.yml"' | base64 --decode
$ kubectl get secrets <???>-config -o json | jq -r '.data."agent.yml"' | base64 --decode
```

## How Agent Operator reconciles the custom resource hierarchy

When a resource hierarchy is created, updated, or deleted, a reconcile occurs.
When a `GrafanaAgent` resource is deleted, the corresponding Grafana Agent
deployment will also be deleted.

Reconciling creates the following cluster resources:

1. A Secret that holds the Grafana Agent
   [configuration]({{< relref "../static/configuration/_index.md" >}}) is generated.
2. A Secret that holds all referenced Secrets or ConfigMaps from
   the resource hierarchy is generated. This ensures that Secrets referenced from a custom
   resource in another namespace can still be read.
3. A Service is created to govern the StatefulSets that are generated.
4. One StatefulSet per Prometheus shard is created.

PodMonitors, Probes, and ServiceMonitors are turned into individual scrape jobs
which all use Kubernetes Service Discovery (SD).

## Sharding and replication

The GrafanaAgent resource can specify a number of shards. Each shard results in
the creation of a StatefulSet with a hashmod + keep relabel_config per job:

```yaml
- source_labels: [__address__]
  target_label: __tmp_hash
  modulus: NUM_SHARDS
  action: hashmod
- source_labels: [__tmp_hash]
  regex: CURRENT_STATEFULSET_SHARD
  action: keep
```

This allows for horizontal scaling capabilities, where each shard
will handle roughly 1/N of the total scrape load. Note that this does not use
consistent hashing, which means changing the number of shards will cause
anywhere between 1/N to N targets to reshuffle.

The sharding mechanism is borrowed from the Prometheus Operator.

The number of replicas can be defined, similarly to the number of shards. This
creates deduplicate shards. This must be paired with a `remote_write` system that
can perform HA deduplication. [Grafana Cloud](/docs/grafana-cloud/) and [Mimir](/docs/mimir/latest/) provide this out of the
box, and the Grafana Agent Operator defaults support these two systems.

The total number of created metrics pods will be the product of `numShards *
numReplicas`.

## Added labels

Two labels are added by default to every metric:

- `cluster`, representing the `GrafanaAgent` deployment. Holds the value of
  `<GrafanaAgent.metadata.namespace>/<GrafanaAgent.metadata.name>`.
- `__replica__`, representing the replica number of the Agent. This label works
   out of the box with Grafana Cloud and Cortex's [HA
   deduplication](https://cortexmetrics.io/docs/guides/ha-pair-handling/).

The shard number is not added as a label, as sharding is designed to be
transparent on the receiver end.

## Enable sharding and replication

To enable sharding and replication, you must set the `shards` and `replicas` properties in the Grafana Agent configuration file. For example, the following configuration file would shard the data into three shards and replicate each shard to two other Grafana Agent instances:

```
shards: 3
replicas: 2
```

You can also enable sharding and replication by setting the `shards` and `replicas` arguments when you start the Grafana Agent. 

### Examples

The following examples show you how to enable sharding and replication in a Kubernetes environment.

* To shard the data into three shards and replicate each shard to two other Grafana Agent instances, you would use the following deployment manifest:

  ```
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: grafana-agent
  spec:
    replicas: 3
    selector:
      matchLabels:
        app: grafana-agent
    template:
      metadata:
        labels:
          app: grafana-agent
      spec:
        containers:
        - name: grafana-agent
          image: grafana/agent:latest
          args:
          - "--shards=3"
          - "--replicas=2"
  ```

* To shard the data into 10 shards and replicate each shard to three other Grafana Agent instances, you would use the following deployment manifest:

  ```
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: grafana-agent
  spec:
    replicas: 10
    selector:
      matchLabels:
        app: grafana-agent
    template:
      metadata:
        labels:
          app: grafana-agent
      spec:
        containers:
        - name: grafana-agent
          image: grafana/agent:latest
          args:
          - "--shards=10"
          - "--replicas=3"
  ```

