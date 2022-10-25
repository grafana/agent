---
aliases:
- /docs/agent/latest/operator/architecture/
title: Operator architecture
weight: 300
---

# Operator architecture

This guide gives a high-level overview of how the Grafana Agent Operator
works.

The Grafana Agent Operator works in two phases:

1. Discover a hierarchy of custom resources
2. Reconcile that hierarchy into a Grafana Agent deployment

## Custom Resource Hierarchy

The root of the custom resource hierarchy is the `GrafanaAgent` resource. It is
primary resource the Operator looks for, and is called the "root" because it
discovers many other sub-resources.

The full hierarchy of custom resources is as follows:

- `GrafanaAgent`
    - `MetricsInstance`
        - `PodMonitor`
        - `Probe`
        - `ServiceMonitor`
    - `LogsInstance`
        - `PodLogs`

Most of the resources above have the ability to reference a ConfigMap or a
Secret. All referenced ConfigMaps or Secrets are added into the resource
hierarchy.

When a hierarchy is established, each item is watched for changes. Any changed
item will cause a reconcile of the root GrafanaAgent resource, either
creating, modifying, or deleting the corresponding Grafana Agent deployment.

A single resource can belong to multiple hierarchies. For example, if two
GrafanaAgents use the same Probe, modifying that Probe will cause both
GrafanaAgents to be reconciled.

### Build the Hierarchy

Grafana Agent Operator builds the hierarchy using label matching on the custom resources. The following figure illustrates the matching. The `GrafanaAgent` picks up the `MetricsInstance`
and `LogsInstance` that match the labels `app.kubernetes.io/name: loki` and `app.kubernetes.io/instance: release-name`, respectively. The instances pick up the resources the same way.

{{<figure class="float-right" src="../../assets/hierarchy.svg" >}}

### Debug the Hierarchy

The generated configurations are saved in secrets. To download and
validate them manually, use the following commands:

```
$ kubectl get secrets <???>-logs-config -o json | jq -r '.data."agent.yml"' | base64 --decode
$ kubectl get secrets <???>-config -o json | jq -r '.data."agent.yml"' | base64 --decode
```

## Reconcile

When a resource hierarchy is created, updated, or deleted, a reconcile occurs.
When a GrafanaAgent resource is deleted, the corresponding Grafana Agent
deployment will also be deleted.

Reconciling creates a few cluster resources:

1. A Secret is generated holding the
   [configuration]({{< relref "../configuration/_index.md" >}}) of the Grafana Agent.
2. Another Secret is created holding all referenced Secrets or ConfigMaps from
   the resource hierarchy. This ensures that Secrets referenced from a custom
   resource in another namespace can still be read.
3. A Service is created to govern the created StatefulSets.
4. One StatefulSet per Prometheus shard is created.

PodMonitors, Probes, and ServiceMonitors are turned into individual scrape jobs
which all use Kubernetes SD.

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

This allows for some decent horizontal scaling capabilities, where each shard
will handle roughly 1/N of the total scrape load. Note that this does not use
consistent hashing, which means changing the number of shards will cause
anywhere between 1/N to N targets to reshuffle.

The sharding mechanism is borrowed from the Prometheus Operator.

The number of replicas can be defined, similarly to the number of shards. This
creates duplicate shards. This must be paired with a remote_write system that
can perform HA duplication. Grafana Cloud and Cortex provide this out of the
box, and the Grafana Agent Operator defaults support these two systems.

The total number of created metrics pods will be product of `numShards *
numReplicas`.

## Labels

Two labels are added by default to every metric:

- `cluster`, representing the `GrafanaAgent` deployment. Holds the value of
  `<GrafanaAgent.metadata.namespace>/<GrafanaAgent.metadata.name>`.
- `__replica__`, representing the replica number of the Agent. This label works
   out of the box with Grafana Cloud and Cortex's [HA
   deduplication](https://cortexmetrics.io/docs/guides/ha-pair-handling/).

The shard number is not added as a label, as sharding is designed to be
transparent on the receiver end.
