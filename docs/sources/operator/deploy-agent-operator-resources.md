---
aliases:
- /docs/grafana-cloud/agent/operator/deploy-agent-operator-resources/
- /docs/grafana-cloud/monitor-infrastructure/agent/operator/deploy-agent-operator-resources/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/operator/deploy-agent-operator-resources/
- /docs/grafana-cloud/send-data/agent/operator/deploy-agent-operator-resources/
- custom-resource-quickstart/
canonical: https://grafana.com/docs/agent/latest/operator/deploy-agent-operator-resources/
description: Learn how to deploy Operator resources
title: Deploy Operator resources
weight: 120
---
# Deploy Operator resources

To start collecting telemetry data, you need to roll out Grafana Agent Operator custom resources into your Kubernetes cluster. Before you can create the custom resources, you must first apply the Agent Custom Resource Definitions (CRDs) and install Agent Operator, with or without Helm. If you haven't yet taken these steps, follow the instructions in one of the following topics:

- [Install Agent Operator]({{< relref "./getting-started" >}})
- [Install Agent Operator with Helm]({{< relref "./helm-getting-started" >}})

Follow the steps in this guide to roll out the Grafana Agent Operator custom resources to:

- Scrape and ship cAdvisor and kubelet metrics to a Prometheus-compatible metrics endpoint.
- Collect and ship your Pods’ container logs to a Loki-compatible logs endpoint.

The hierarchy of custom resources is as follows:

- `GrafanaAgent`
  - `MetricsInstance`
    - `PodMonitor`
    - `Probe`
    - `ServiceMonitor`
  - `LogsInstance`
    - `PodLogs`

To learn more about the custom resources Agent Operator provides and their hierarchy, see [Grafana Agent Operator architecture]({{< relref "./architecture" >}}).

{{< admonition type="note" >}}
Agent Operator is currently in [beta]({{< relref "../stability.md#beta" >}}) and its custom resources are subject to change.
{{< /admonition >}}

## Before you begin

Before you begin, make sure that you have deployed the Grafana Agent Operator CRDs and installed Agent Operator into your cluster. See [Install Grafana Agent Operator with Helm]({{< relref "./helm-getting-started" >}}) or [Install Grafana Agent Operator]({{< relref "./getting-started" >}}) for instructions.

## Deploy the GrafanaAgent resource

In this section, you'll roll out a `GrafanaAgent` resource. See [Grafana Agent Operator architecture]({{< relref "./architecture" >}}) for a discussion of the resources in the `GrafanaAgent` resource hierarchy.

{{< admonition type="note" >}}
Due to the variety of possible deployment architectures, the official Agent Operator Helm chart does not provide built-in templates for the custom resources described in this guide. You must configure and deploy these manually as described in this section. We recommend templating and adding the following manifests to your own in-house Helm charts and GitOps flows.
{{< /admonition >}}

To deploy the `GrafanaAgent` resource:

1. Copy the following manifests to a file:

    ```yaml
    apiVersion: monitoring.grafana.com/v1alpha1
    kind: GrafanaAgent
    metadata:
      name: grafana-agent
      namespace: default
      labels:
        app: grafana-agent
    spec:
      image: grafana/agent:{{< param "AGENT_RELEASE" >}}
      integrations:
        selector:
          matchLabels:
              agent: grafana-agent-integrations
      logLevel: info
      serviceAccountName: grafana-agent
      metrics:
        instanceSelector:
          matchLabels:
            agent: grafana-agent-metrics
        externalLabels:
          cluster: cloud

      logs:
        instanceSelector:
          matchLabels:
            agent: grafana-agent-logs

    ---

    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: grafana-agent
      namespace: default

    ---

    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: grafana-agent
    rules:
    - apiGroups:
      - ""
      resources:
      - nodes
      - nodes/proxy
      - nodes/metrics
      - services
      - endpoints
      - pods
      - events
      verbs:
      - get
      - list
      - watch
    - apiGroups:
      - networking.k8s.io
      resources:
      - ingresses
      verbs:
      - get
      - list
      - watch
    - nonResourceURLs:
      - /metrics
      - /metrics/cadvisor
      verbs:
      - get

    ---

    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: grafana-agent
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: grafana-agent
    subjects:
    - kind: ServiceAccount
      name: grafana-agent
      namespace: default
    ```

    In the first manifest, the `GrafanaAgent` resource:

    - Specifies an Agent image version.
    - Specifies `MetricsInstance` and `LogsInstance` selectors. These search for `MetricsInstances` and `LogsInstances` in the same namespace with labels matching `agent: grafana-agent-metrics` and `agent: grafana-agent-logs`, respectively.
    - Sets a `cluster: cloud` label for all metrics shipped to your Prometheus-compatible endpoint. Change this label to your cluster name. To search for `MetricsInstances` or `LogsInstances` in a *different* namespace, use the `instanceNamespaceSelector` field. To learn more about this field, see the `GrafanaAgent` [CRD specification](https://github.com/grafana/agent/tree/main/operations/agent-static-operator/crds/monitoring.grafana.com_grafanaagents.yaml).

1. Customize the manifests as needed and roll them out to your cluster using `kubectl apply -f` followed by the filename.

    This step creates a `ServiceAccount`, `ClusterRole`, and `ClusterRoleBinding` for the `GrafanaAgent` resource.

    Deploying a `GrafanaAgent` resource on its own does not spin up Agent Pods. Agent Operator creates Agent Pods once `MetricsInstance` and `LogsIntance` resources have been created. Follow the instructions in the [Deploy a MetricsInstance resource](#deploy-a-metricsinstance-resource) and [Deploy LogsInstance and PodLogs resources](#deploy-logsinstance-and-podlogs-resources) sections to create these resources.

### Disable feature flags reporting

To disable the [reporting]({{< relref "../static/configuration/flags.md#report-information-usage" >}}) usage of feature flags to Grafana, set `disableReporting` field to `true`.

### Disable support bundle generation

To disable the [support bundles functionality]({{< relref "../static/configuration/flags.md#support-bundles" >}}), set the `disableSupportBundle` field to `true`.

## Deploy a MetricsInstance resource

Next, you'll roll out a `MetricsInstance` resource. `MetricsInstance` resources define a `remote_write` sink for metrics and configure one or more selectors to watch for creation and updates to `*Monitor` objects. These objects allow you to define Agent scrape targets via Kubernetes manifests:

- [ServiceMonitors](https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/api.md#servicemonitor)
- [PodMonitors](https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/api.md#podmonitor)
- [Probes](https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/api.md#probe)

To deploy a `MetricsInstance` resource:

1. Copy the following manifest to a file:

    ```yaml
    apiVersion: monitoring.grafana.com/v1alpha1
    kind: MetricsInstance
    metadata:
      name: primary
      namespace: default
      labels:
        agent: grafana-agent-metrics
    spec:
      remoteWrite:
      - url: your_remote_write_URL
        basicAuth:
          username:
            name: primary-credentials-metrics
            key: username
          password:
            name: primary-credentials-metrics
            key: password

      # Supply an empty namespace selector to look in all namespaces. Remove
      # this to only look in the same namespace as the MetricsInstance CR
      serviceMonitorNamespaceSelector: {}
      serviceMonitorSelector:
        matchLabels:
          instance: primary

      # Supply an empty namespace selector to look in all namespaces. Remove
      # this to only look in the same namespace as the MetricsInstance CR.
      podMonitorNamespaceSelector: {}
      podMonitorSelector:
        matchLabels:
          instance: primary

      # Supply an empty namespace selector to look in all namespaces. Remove
      # this to only look in the same namespace as the MetricsInstance CR.
      probeNamespaceSelector: {}
      probeSelector:
        matchLabels:
          instance: primary
    ```

1. Replace the `remote_write` URL and customize the namespace and label configuration as necessary.

    This step associates the `MetricsInstance` resource with the `agent: grafana-agent` `GrafanaAgent` resource deployed in the previous step. The `MetricsInstance` resource watches for creation and updates to `*Monitors` with the `instance: primary` label.

1. Once you've rolled out the manifest, create the `basicAuth` credentials [using a Kubernetes Secret](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-config-file/):

    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: primary-credentials-metrics
      namespace: default
    stringData:
      username: 'your_cloud_prometheus_username'
      password: 'your_cloud_prometheus_API_key'
    ```

If you're using Grafana Cloud, you can find your hosted Loki endpoint username and password by clicking **Details** on the Loki tile on the [Grafana Cloud Portal](/profile/org). If you want to base64-encode these values yourself, use `data` instead of `stringData`.

Once you've rolled out the `MetricsInstance` and its Secret, you can confirm that the `MetricsInstance` Agent is up and running using `kubectl get pod`. Since you haven't defined any monitors yet, this Agent doesn't have any scrape targets defined. In the next section, you'll create scrape targets for the cAdvisor and kubelet endpoints exposed by the `kubelet` service in the cluster.

## Create ServiceMonitors for kubelet and cAdvisor endpoints

Next, you'll create ServiceMonitors for kubelet and cAdvisor metrics exposed by the `kubelet` service. Every Node in your cluster exposes kubelet and cAdvisor metrics at `/metrics` and `/metrics/cadvisor`, respectively. Agent Operator creates a `kubelet` service that exposes these Node endpoints so that they can be scraped using ServiceMonitors.

To scrape the kubelet and cAdvisor endpoints:

1. Copy the following kubelet ServiceMonitor manifest to a file, then roll it out in your cluster using `kubectl apply -f` followed by the filename.

    ```yaml
    apiVersion: monitoring.coreos.com/v1
    kind: ServiceMonitor
    metadata:
      labels:
        instance: primary
      name: kubelet-monitor
      namespace: default
    spec:
      endpoints:
      - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
        honorLabels: true
        interval: 60s
        metricRelabelings:
        - action: keep
          regex: kubelet_cgroup_manager_duration_seconds_count|go_goroutines|kubelet_pod_start_duration_seconds_count|kubelet_runtime_operations_total|kubelet_pleg_relist_duration_seconds_bucket|volume_manager_total_volumes|kubelet_volume_stats_capacity_bytes|container_cpu_usage_seconds_total|container_network_transmit_bytes_total|kubelet_runtime_operations_errors_total|container_network_receive_bytes_total|container_memory_swap|container_network_receive_packets_total|container_cpu_cfs_periods_total|container_cpu_cfs_throttled_periods_total|kubelet_running_pod_count|node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate|container_memory_working_set_bytes|storage_operation_errors_total|kubelet_pleg_relist_duration_seconds_count|kubelet_running_pods|rest_client_request_duration_seconds_bucket|process_resident_memory_bytes|storage_operation_duration_seconds_count|kubelet_running_containers|kubelet_runtime_operations_duration_seconds_bucket|kubelet_node_config_error|kubelet_cgroup_manager_duration_seconds_bucket|kubelet_running_container_count|kubelet_volume_stats_available_bytes|kubelet_volume_stats_inodes|container_memory_rss|kubelet_pod_worker_duration_seconds_count|kubelet_node_name|kubelet_pleg_relist_interval_seconds_bucket|container_network_receive_packets_dropped_total|kubelet_pod_worker_duration_seconds_bucket|container_start_time_seconds|container_network_transmit_packets_dropped_total|process_cpu_seconds_total|storage_operation_duration_seconds_bucket|container_memory_cache|container_network_transmit_packets_total|kubelet_volume_stats_inodes_used|up|rest_client_requests_total
          sourceLabels:
          - __name__
        port: https-metrics
        relabelings:
        - sourceLabels:
          - __metrics_path__
          targetLabel: metrics_path
        - action: replace
          targetLabel: job
          replacement: integrations/kubernetes/kubelet
        scheme: https
        tlsConfig:
          insecureSkipVerify: true
      namespaceSelector:
        matchNames:
        - default
      selector:
        matchLabels:
          app.kubernetes.io/name: kubelet
    ```

1. Copy the following cAdvisor ServiceMonitor manifest to a file, then roll it out in your cluster using `kubectl apply -f` followed by the filename.

    ```yaml
    apiVersion: monitoring.coreos.com/v1
    kind: ServiceMonitor
    metadata:
      labels:
        instance: primary
      name: cadvisor-monitor
      namespace: default
    spec:
      endpoints:
      - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
        honorLabels: true
        honorTimestamps: false
        interval: 60s
        metricRelabelings:
        - action: keep
          regex: kubelet_cgroup_manager_duration_seconds_count|go_goroutines|kubelet_pod_start_duration_seconds_count|kubelet_runtime_operations_total|kubelet_pleg_relist_duration_seconds_bucket|volume_manager_total_volumes|kubelet_volume_stats_capacity_bytes|container_cpu_usage_seconds_total|container_network_transmit_bytes_total|kubelet_runtime_operations_errors_total|container_network_receive_bytes_total|container_memory_swap|container_network_receive_packets_total|container_cpu_cfs_periods_total|container_cpu_cfs_throttled_periods_total|kubelet_running_pod_count|node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate|container_memory_working_set_bytes|storage_operation_errors_total|kubelet_pleg_relist_duration_seconds_count|kubelet_running_pods|rest_client_request_duration_seconds_bucket|process_resident_memory_bytes|storage_operation_duration_seconds_count|kubelet_running_containers|kubelet_runtime_operations_duration_seconds_bucket|kubelet_node_config_error|kubelet_cgroup_manager_duration_seconds_bucket|kubelet_running_container_count|kubelet_volume_stats_available_bytes|kubelet_volume_stats_inodes|container_memory_rss|kubelet_pod_worker_duration_seconds_count|kubelet_node_name|kubelet_pleg_relist_interval_seconds_bucket|container_network_receive_packets_dropped_total|kubelet_pod_worker_duration_seconds_bucket|container_start_time_seconds|container_network_transmit_packets_dropped_total|process_cpu_seconds_total|storage_operation_duration_seconds_bucket|container_memory_cache|container_network_transmit_packets_total|kubelet_volume_stats_inodes_used|up|rest_client_requests_total
          sourceLabels:
          - __name__
        path: /metrics/cadvisor
        port: https-metrics
        relabelings:
        - sourceLabels:
          - __metrics_path__
          targetLabel: metrics_path
        - action: replace
          targetLabel: job
          replacement: integrations/kubernetes/cadvisor
        scheme: https
        tlsConfig:
          insecureSkipVerify: true
      namespaceSelector:
        matchNames:
        - default
      selector:
        matchLabels:
          app.kubernetes.io/name: kubelet
    ```

These two ServiceMonitors configure Agent to scrape all the kubelet and cAdvisor endpoints in your Kubernetes cluster (one of each per Node). In addition, it defines a `job` label which you can update (it is preset here for compatibility with Grafana Cloud's Kubernetes integration). It also provides an allowlist containing a core set of Kubernetes metrics to reduce remote metrics usage. If you don't need this allowlist, you can omit it, however, your metrics usage will increase significantly.

 When you're done, Agent should now be shipping kubelet and cAdvisor metrics to your remote Prometheus endpoint. To check this in Grafana Cloud, go to your dashboards, select **Integration - Kubernetes**, then select **Kubernetes / Kubelet**.

## Deploy LogsInstance and PodLogs resources

Next, you'll deploy a `LogsInstance` resource to collect logs from your cluster Nodes and ship these to your remote Loki endpoint. Agent Operator deploys a DaemonSet of Agents in your cluster that will tail log files defined in `PodLogs` resources.

To deploy the `LogsInstance` resource into your cluster:

1. Copy the following manifest to a file, then roll it out in your cluster using `kubectl apply -f` followed by the filename.

    ```yaml
    apiVersion: monitoring.grafana.com/v1alpha1
    kind: LogsInstance
    metadata:
      name: primary
      namespace: default
      labels:
        agent: grafana-agent-logs
    spec:
      clients:
      - url: your_remote_logs_URL
        basicAuth:
          username:
            name: primary-credentials-logs
            key: username
          password:
            name: primary-credentials-logs
            key: password

      # Supply an empty namespace selector to look in all namespaces. Remove
      # this to only look in the same namespace as the LogsInstance CR
      podLogsNamespaceSelector: {}
      podLogsSelector:
        matchLabels:
          instance: primary
    ```

    This `LogsInstance` picks up `PodLogs` resources with the `instance: primary` label. Be sure to set the Loki URL to the correct push endpoint. For Grafana Cloud, this will look similar to `logs-prod-us-central1.grafana.net/loki/api/v1/push`, however check the [Grafana Cloud Portal](/profile/org) to confirm by clicking **Details** on the Loki tile.

    Also note that this example uses the `agent: grafana-agent-logs` label, which associates this `LogsInstance` with the `GrafanaAgent` resource defined earlier. This means that it will inherit requests, limits, affinities and other properties defined in the `GrafanaAgent` custom resource.

1. To create the Secret for the `LogsInstance` resource, copy the following Secret manifest to a file, then roll it out in your cluster using `kubectl apply -f` followed by the filename.

    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: primary-credentials-logs
      namespace: default
    stringData:
      username: 'your_username_here'
      password: 'your_password_here'
    ```

    If you're using Grafana Cloud, you can find your hosted Loki endpoint username and password by clicking **Details** on the Loki tile on the [Grafana Cloud Portal](/profile/org). If you want to base64-encode these values yourself, use `data` instead of `stringData`.

1. Copy the following `PodLogs` manifest to a file, then roll it to your cluster using `kubectl apply -f` followed by the filename. The manifest defines your logging targets. Agent Operator  turns this into Agent configuration for the logs subsystem, and rolls it out to the DaemonSet of logging Agents.

    {{< admonition type="note" >}}
    The following is a minimal working example which you should adapt to your production needs.
    {{< /admonition >}}

    ```yaml
    apiVersion: monitoring.grafana.com/v1alpha1
    kind: PodLogs
    metadata:
      labels:
        instance: primary
      name: kubernetes-pods
      namespace: default
    spec:
      pipelineStages:
        - docker: {}
      namespaceSelector:
        matchNames:
        - default
      selector:
        matchLabels: {}
    ```

    This example tails container logs for all Pods in the `default` namespace. You can restrict the set of matched Pods by using the `matchLabels` selector. You can also set additional `pipelineStages` and create `relabelings` to add or modify log line labels. To learn more about the `PodLogs` specification and available resource fields, see the [PodLogs CRD](https://github.com/grafana/agent/tree/main/operations/agent-static-operator/crds/monitoring.grafana.com_podlogs.yaml).

    The above `PodLogs` resource adds the following labels to log lines:

    - `namespace`
    - `service`
    - `pod`
    - `container`
    - `job` (set to `PodLogs_namespace/PodLogs_name`)
    - `__path__` (the path to log files, set to `/var/log/pods/*$1/*.log` where `$1` is `__meta_kubernetes_pod_uid/__meta_kubernetes_pod_container_name`)

    To learn more about this configuration format and other available labels, see the [Promtail Scraping](/docs/loki/latest/clients/promtail/scraping/#promtail-scraping-service-discovery) documentation. Agent Operator loads this configuration into the `LogsInstance` agents automatically.

The DaemonSet of logging agents should be tailing your container logs, applying  default labels to the log lines, and shipping them to your remote Loki endpoint.

## Summary

You've now rolled out the following into your cluster:

- A `GrafanaAgent` resource that discovers one or more `MetricsInstance` and `LogsInstances` resources.
- A `MetricsInstance` resource that defines where to ship collected metrics.
- A `ServiceMonitor` resource to collect cAdvisor and kubelet metrics.
- A `LogsInstance` resource that defines where to ship collected logs.
- A `PodLogs` resource to collect container logs from Kubernetes Pods.

## What's next

You can verify that everything is working correctly by navigating to your Grafana instance and querying your Loki and Prometheus data sources.

> Tip: You can deploy multiple GrafanaAgent resources to isolate allocated resources to the agent pods. By default, the GrafanaAgent resource determines the resources of all deployed agent containers. However, you might want different memory limits for metrics versus logs.
