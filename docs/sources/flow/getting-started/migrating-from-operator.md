---
aliases:
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/getting-started/migrating-from-operator/
- /docs/grafana-cloud/send-data/agent/flow/getting-started/migrating-from-operator/
canonical: https://grafana.com/docs/agent/latest/flow/getting-started/migrating-from-operator/
description: Migrating from Grafana Agent Operator to Grafana Agent Flow
menuTitle: Migrate from Operator
title: Migrating from Grafana Agent Operator to Grafana Agent Flow
weight: 320
---

# Migrating from Grafana Agent Operator to Grafana Agent Flow

With the release of Flow, Grafana Agent Operator is no longer the recommended way to deploy Grafana Agent in Kubernetes. Some of the Operator functionality has been moved into Grafana Agent
itself, and the remaining functionality has been replaced by our Helm Chart.

- The Monitor types (`PodMonitor`, `ServiceMonitor`, `Probe`, and `LogsInstance`) are all supported natively by Grafana Agent in Flow mode. You are no longer
required to use the Operator to consume those CRDs for dynamic monitoring in your cluster.
- The parts of the Operator that deploy the Agent itself (`GrafanaAgent`, `MetricsInstance`, and `LogsInstance` CRDs) are deprecated. We now recommend
operator users use the [Grafana Agent Helm Chart](https://grafana.com/docs/agent/latest/flow/setup/install/kubernetes/) to deploy the Agent directly to your clusters.

This guide will provide some steps to get started with Grafana Agent for users coming from Grafana Agent Operator.

## Deploy Grafana Agent with Helm

1. You will need to create a `values.yaml` file, which contains options for deploying your Agent. You may start with the [default values](https://github.com/grafana/agent/blob/main/operations/helm/charts/grafana-agent/values.yaml) and customize as you see fit, or start with this snippet, which should be a good starting point for what the Operator does:

    ```yaml
    agent:
      mode: 'flow'
      configMap:
        create: true
      clustering:
        enabled: true
    controller:
      type: 'statefulset'
      replicas: 2
    crds:
      create: false
    ```

    This configuration will deploy Grafana Agent as a `StatefulSet` using the built-in [clustering](https://grafana.com/docs/agent/latest/flow/concepts/clustering/) functionality to allow distributing scrapes across all Agent Pods. 
    
    This is not the only deployment mode possible. For example, you may want to use a `DaemonSet` to collect host-level logs or metrics. See [the Agent deployment guide](https://grafana.com/docs/agent/latest/flow/setup/deploy-agent/) for more details about different topologies.

2. Create a Flow config file, `agent.river`.

    We will be adding to this config in the next step as we convert `MetricsInstances`. You can add any additional configuration to this file as you desire.

3. Install the grafana helm repository:

    ```
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
    ```

4. Create a Helm release. You may name the release anything you like. Here we are installing a release named `grafana-agent-metrics` in the `monitoring` namespace.

    ```shell
    helm upgrade grafana-agent-metrics grafana/grafana-agent -i -n monitoring -f values.yaml --set-file agent.configMap.content=agent.river
    ```

    This command uses the `--set-file` flag to pass the configuration file as a Helm value, so that we can continue to edit it as a regular River file.

## Convert `MetricsIntances` to Flow components

A `MetricsInstance` resource primarily defines:

- The remote endpoint(s) Grafana Agent should send metrics to.
- Which `PodMonitor`, `ServiceMonitor`, and `Probe` resources this Agent should discover.

These functions can be done in Grafana Agent Flow with the `prometheus.remote_write`, `prometheus.operator.podmonitors`, `prometheus.operator.servicemonitors`, and `prometheus.operator.probes` components respectively.

This is a River sample that is equivalent to the `MetricsInstance` from our [operator guide](https://grafana.com/docs/agent/latest/operator/deploy-agent-operator-resources/#deploy-a-metricsinstance-resource):

```river

// read the credentials secret for remote_write authorization
remote.kubernetes.secret "credentials" {
  namespace = "monitoring"
  name = "primary-credentials-metrics"
}

prometheus.remote_write "primary" {
    endpoint {
        url = "https://PROMETHEUS_URL/api/v1/push"
        basic_auth {
            username = nonsensitive(remote.kubernetes.secret.credentials.data["username"])
            password = remote.kubernetes.secret.credentials.data["password"]
        }
    }
}

prometheus.operator.podmonitors "primary" {
    forward_to = [prometheus.remote_write.primary.receiver]
    // leave out selector to find all podmonitors in the entire cluster
    selector {
        match_labels = {instance = "primary"}
    }
}

prometheus.operator.servicemonitors "primary" {
    forward_to = [prometheus.remote_write.primary.receiver]
    // leave out selector to find all servicemonitors in the entire cluster
    selector {
        match_labels = {instance = "primary"}
    }
}

```

You will need to replace `PROMETHEUS_URL` with the actual endpoint you want to send metrics to.

This configuration will discover all `PodMonitor`, `ServiceMonitor`, and `Probe` resources in your cluster that match our label selector `instance=primary`. It will then scrape metrics from their targets and forward them to your remote write endpoint.

You may need to customize this configuration further if you use additional features in your `MetricsInstance` resources.  Refer to the documentation for the relevant components for additional information:

- [remote.kubernetes.secret](https://grafana.com/docs/agent/latest/flow/reference/components/remote.kubernetes.secret)
- [prometheus.remote_write](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.remote_write)
- [prometheus.operator.podmonitors](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.operator.podmonitors)
- [prometheus.operator.servicemonitors](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.operator.servicemonitors)
- [prometheus.operator.probes](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.operator.probes)
- [prometheus.scrape](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.scrape)

## Collecting Logs

Our current recommendation is to create an additional DaemonSet deployment of Grafana Agents to scrape logs.

> We have components that can scrape pod logs directly from the Kubernetes API without needing a DaemonSet deployment. These are 
> still considered experimental, but if you would like to try them, see the documentation for [loki.source.kubernetes](https://grafana.com/docs/agent/latest/flow/reference/components/loki.source.kubernetes/) and 
> [loki.source.podlogs](https://grafana.com/docs/agent/latest/flow/reference/components/loki.source.podlogs/).

These values are close to what the Operator currently deploys for logs:

```yaml
agent:
  mode: 'flow'
  configMap:
    create: true
  clustering:
    enabled: false
  controller:
    type: 'daemonset'
  mounts:
    # -- Mount /var/log from the host into the container for log collection.
    varlog: true
```

This command will install a release named `grafana-agent-logs` in the `monitoring` namespace:

```
helm upgrade grafana-agent-logs grafana/grafana-agent -i -n monitoring -f values-logs.yaml --set-file agent.configMap.content=agent-logs.river
```

This simple configuration will scrape logs for every pod on each node:

```river
// read the credentials secret for remote_write authorization
remote.kubernetes.secret "credentials" {
  namespace = "monitoring"
  name      = "primary-credentials-logs"
}

discovery.kubernetes "pods" {
  role = "pod"
  // limit to pods on this node to reduce the amount we need to filter
  selectors {
    role  = "pod"
    field = "spec.nodeName=" + env("HOSTNAME")
  }
}

discovery.relabel "pod_logs" {
  targets = discovery.kubernetes.pods.targets
  rule {
    source_labels = ["__meta_kubernetes_namespace"]
    target_label  = "namespace"
  }
  rule {
    source_labels = ["__meta_kubernetes_pod_name"]
    target_label  = "pod"
  }
  rule {
    source_labels = ["__meta_kubernetes_pod_container_name"]
    target_label  = "container"
  }
  rule {
    source_labels = ["__meta_kubernetes_namespace", "__meta_kubernetes_pod_name"]
    separator     = "/"
    target_label  = "job"
  }
  rule {
    source_labels = ["__meta_kubernetes_pod_uid", "__meta_kubernetes_pod_container_name"]
    separator     = "/"
    action        = "replace"
    replacement   = "/var/log/pods/*$1/*.log"
    target_label  = "__path__"
  }
}

local.file_match "pod_logs" {
  path_targets = discovery.relabel.pod_logs.output
}

loki.source.file "pod_logs" {
  targets    = local.file_match.pod_logs.targets
  forward_to = [loki.process.pod_logs.receiver]
}

// basic processing to parse the container format. You can add additional processing stages
// to match your application logs.
loki.process "pod_logs" {
  stage.match {
    selector = "{tmp_container_runtime=\"containerd\"}"
    // the cri processing stage extracts the following k/v pairs: log, stream, time, flags
    stage.cri {}
    // Set the extract flags and stream values as labels
    stage.labels {
      values = {
        flags   = "",
        stream  = "",
      }
    }
  }

  // if the label tmp_container_runtime from above is docker parse using docker
  stage.match {
    selector = "{tmp_container_runtime=\"docker\"}"
    // the docker processing stage extracts the following k/v pairs: log, stream, time
    stage.docker {}

    // Set the extract stream value as a label
    stage.labels {
      values = {
        stream  = "",
      }
    }
  }

  // drop the temporary container runtime label as it is no longer needed
  stage.label_drop {
    values = ["tmp_container_runtime"]
  }

  forward_to = [loki.write.loki.receiver]
}

loki.write "loki" {
  endpoint {
    url = "https://LOKI_URL/loki/api/v1/push"
    basic_auth {
      username = nonsensitive(remote.kubernetes.secret.credentials.data["username"])
      password = remote.kubernetes.secret.credentials.data["password"]
    }
}
}
```

You will need to replace `LOKI_URL` with the actual endpoint of your Loki instance. The logging subsytem is very powerful
and has many options for processing logs. For further details see the [component documentation](https://grafana.com/docs/agent/latest/flow/reference/components/).


## Integrations

The `Integration` CRD is not supported with Grafana Agent Flow, however all static mode integrations have an equivalent component in the [`prometheus.exporter`](https://grafana.com/docs/agent/latest/flow/reference/components) namespace. The reference docs should help convert those integrations to their Flow equivalent.
