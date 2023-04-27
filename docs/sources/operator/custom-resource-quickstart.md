---
title: Custom Resource Quickstart
weight: 120
---
# Grafana Agent Operator Custom Resource Quickstart

In this guide you'll learn how to deploy [Agent Operator]({{< relref "./_index.md" >}})'s custom resources into your Kubernetes cluster.

You'll roll out the following custom resources (CRs):

- A `GrafanaAgent` resource, which discovers one or more `MetricsInstance` and `LogsInstances` resources.
- A `MetricsInstance` resource that defines where to ship collected metrics. Under the hood, this rolls out a Grafana Agent StatefulSet that will scrape and ship metrics to a `remote_write` endpoint.
- A `ServiceMonitor` resource to collect cAdvisor and kubelet metrics. Under the hood, this configures the `MetricsInstance` / Agent StatefulSet.
- A `LogsInstance` resource that defines where to ship collected logs. Under the hood, this rolls out a Grafana Agent DaemonSet that will tail log files on your cluster nodes.
- A `PodLogs` resource to collect container logs from Kubernetes Pods. Under the hood, this configures the`LogsInstance` / Agent DaemonSet.

To learn more about the custom resources Operator provides and their hierarchy, please consult [Operator architecture]({{< relref "./architecture.md" >}}).

> **Note:** Agent Operator is currently in beta and its custom resources are subject to change as the project evolves. It currently supports the metrics and logs subsystems of Grafana Agent. Integrations and traces support is coming soon.

By the end of this guide, you will be scraping and shipping cAdvisor and Kubelet metrics to a Prometheus-compatible metrics endpoint. You'll also be collecting and shipping your Pods' container logs to a Loki-compatible logs endpoint.

## Prerequisites

Before you begin, make sure that you have installed Agent Operator into your cluster. You can learn how to do this in:
- [Installing Grafana Agent Operator with Helm]({{< relref "./helm-getting-started.md" >}})
- [Installing Grafana Agent Operator]({{< relref "./getting-started.md" >}})

## Step 1: Deploy GrafanaAgent resource

In this step you'll roll out a `GrafanaAgent` resource. A `GrafanaAgent` resource discovers `MetricsInstance` and `LogsInstance` resources and defines the Grafana Agent image, Pod requests, limits, affinities, and tolerations. Pod attributes can only be defined at the GrafanaAgent level and are propagated to `MetricsInstance` and `LogsInstance` Pods. To learn more, please see the GrafanaAgent [Custom Resource Definition](https://github.com/grafana/agent/blob/main/production/operator/crds/monitoring.grafana.com_grafanaagents.yaml).

> **Note:** Due to the variety of possible deployment architectures, the official Agent Operator Helm chart does not provide built-in templates for the custom resources described in this quickstart. These must be configured and deployed manually. However, you are encouraged to template and add the following manifests to your own in-house Helm charts and GitOps flows.

Roll out the following manifests in your cluster:

```yaml
apiVersion: monitoring.grafana.com/v1alpha1
kind: GrafanaAgent
metadata:
  name: grafana-agent
  namespace: default
  labels:
    app: grafana-agent
spec:
  image: grafana/agent:v0.28.1
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

This creates a ServiceAccount, ClusterRole, and ClusterRoleBinding for the GrafanaAgent resource. It also creates a GrafanaAgent resource and specifies an Agent image version. Finally, the GrafanaAgent resource specifies `MetricsInstance` and `LogsInstance` selectors. These search for MetricsInstances and LogsInstances in the **same namespace** with labels matching `agent: grafana-agent-metrics` and `agent: grafana-agent-logs`, respectively. It also sets a `cluster: cloud` label for all metrics shipped your Prometheus-compatible endpoint. You should change this label to your desired cluster name. To search for MetricsInstances or LogsInstances in a *different* namespace, please use the `instanceNamespaceSelector` field. To learn more about this field, please consult the `GrafanaAgent` [CRD specification](https://github.com/grafana/agent/blob/main/production/operator/crds/monitoring.grafana.com_grafanaagents.yaml#L3789).

The full hierarchy of custom resources is as follows:

- `GrafanaAgent`
  - `MetricsInstance`
    - `PodMonitor`
    - `Probe`
    - `ServiceMonitor`
  - `LogsInstance`
    - `PodLogs`

Deploying a GrafanaAgent resource on its own will not spin up any Agent Pods. Agent Operator will create Agent Pods once MetricsInstance and LogsIntance resources have been created. In the next step, you'll roll out a `MetricsInstance` resource to scrape cAdvisor and Kubelet metrics and ship these to your Prometheus-compatible metrics endpoint.

### Disable feature flags reporting

If you would like to disable the [reporting]({{< relref "../configuration/flags.md/#report-information-usage" >}}) usage of feature flags to Grafana, set `disableReporting` field to `true`.

## Step 2: Deploy a MetricsInstance resource

In this step you'll roll out a MetricsInstance resource. MetricsInstance resources define a `remote_write` sink for metrics and configure one or more selectors to watch for creation and updates to `*Monitor` objects. These objects allow you to define Agent scrape targets via K8s manifests:

- [ServiceMonitors](https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/api.md#servicemonitor)
- [PodMonitors](https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/api.md#podmonitor)
- [Probes](https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/api.md#probe)

Roll out the following manifest into your cluster:

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

Be sure to replace the `remote_write` URL and customize the namespace and label configuration as necessary. This will associate itself with the `agent: grafana-agent` GrafanaAgent resource deployed in the previous step, and watch for creation and updates to `*Monitors` monitors with the the `instance: primary` label.

Once you've rolled out this manifest, create the `basicAuth` credentials using a Kubernetes Secret:

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

If you're using Grafana Cloud, you can find your hosted Prometheus endpoint username and password in the [Grafana Cloud Portal](https://grafana.com/profile/org ). You may wish to base64-encode these values yourself. In this case, please use `data` instead of `stringData`.

Once you've rolled out the `MetricsInstance` and its Secret, you can confirm that the MetricsInstance Agent is up and running with `kubectl get pod`. Since we haven't defined any monitors yet, this Agent will not have any scrape targets defined. In the next step, we'll create scrape targets for the cAdvisor and kubelet endpoints exposed by the `kubelet` service in the cluster.

## Step 3: Create ServiceMonitors for kubelet and cAdvisor endpoints

In this step, you'll create ServiceMonitors for kubelet and cAdvisor metrics exposed by the `kubelet` Service. Every node in your cluster exposes kubelet and cadvisor metrics at `/metrics` and `/metrics/cadvisor` respectively. Agent Operator creates a `kubelet` service that exposes these Node endpoints so that they can be scraped using ServiceMonitors.

To scrape these two endpoints, roll out the following two ServiceMonitors in your cluster:

- Kubelet ServiceMonitor

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

- cAdvsior ServiceMonitor

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

These two ServiceMonitors configure Agent to scrape all the Kubelet and cAdvisor endpoints in your Kubernetes cluster (one of each per Node). In addition, it defines a `job` label which you may change (it is preset here for compatibility with Grafana Cloud's Kubernetes integration), and allowlists a core set of Kubernetes metrics to reduce remote metrics usage. If you don't need this allowlist, you may omit it, however note that your metrics usage will increase significantly.

 When you're done, Agent should now be shipping Kubelet and cAdvisor metrics to your remote Prometheus endpoint.

## Step 4: Deploy LogsInstance and PodLogs resources

In this step, you'll deploy a LogsInstance resource to collect logs from your cluster nodes and ship these to your remote Loki endpoint. Under the hood, Agent Operator will deploy a DaemonSet of Agents in your cluster that will tail log files defined in PodLogs resources.

Deploy the LogsInstance into your cluster:

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

This LogsInstance will pick up PodLogs resources with the `instance: primary` label. Be sure to set the Loki URL to the correct push endpoint (for Grafana Cloud, this will be something like `logs-prod-us-central1.grafana.net/loki/api/v1/push`, however you should check the Cloud Portal to confirm).

Also note that we are using the `agent: grafana-agent-logs` label here, which will associate this LogsInstance with the GrafanaAgent resource defined in Step 1. This means that it will inherit requests, limits, affinities and other properties defined in the GrafanaAgent custom resource.

Create the Secret for the LogsInstance resource:

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

If you're using Grafana Cloud, you can find your hosted Loki endpoint username and password in the [Grafana Cloud Portal](https://grafana.com/profile/org). You may wish to base64-encode these values yourself. In this case, please use `data` instead of `stringData`.

Finally, we'll roll out a PodLogs resource to define our logging targets. Under the hood, Agent Operator will turn this into Agent config for the logs subsystem, and roll it out to the DaemonSet of logging agents.

The following is a minimal working example which you should adapt to your production needs:

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

This tails container logs for all Pods in the `default` Namespace. You can restrict the set of Pods matched by using the `matchLabels` selector. You can also set additional `pipelineStages` and create `relabelings` to add or modify log line labels. To learn more about the PodLogs spec and available resource fields, please see the [PodLogs CRD](https://github.com/grafana/agent/blob/main/production/operator/crds/monitoring.grafana.com_podlogs.yaml).

Under the hood, the above PodLogs resource will add the following labels to log lines:

- `namespace`
- `service`
- `pod`
- `container`
- `job`
  - Set to `PodLogs_namespace/PodLogs_name`
- `__path__` (the path to log files)
  - Set to `/var/log/pods/*$1/*.log` where `$1` is `__meta_kubernetes_pod_uid/__meta_kubernetes_pod_container_name`

To learn more about this config format and other available labels, please see the [Promtail Scraping](https://grafana.com/docs/loki/latest/clients/promtail/scraping/#promtail-scraping-service-discovery) reference documentation. Agent Operator will load this config into the LogsInstance agents automatically.

At this point the DaemonSet of logging agents should be tailing your container logs, applying some default labels to the log lines, and shipping them to your remote Loki endpoint.

## Conclusion

At this point you've rolled out the following into your cluster:

- A `GrafanaAgent` resource, which discovers one or more `MetricsInstance` and `LogsInstances` resources.
- A `MetricsInstance`  resource that defines where to ship collected metrics.
- A `ServiceMonitor` resource to collect cAdvisor and kubelet metrics.
- A `LogsInstance` resource that defines where to ship collected logs.
- A `PodLogs` resource to collect container logs from Kubernetes Pods.

You can verify that everything is working correctly by navigating to your Grafana instance and querying your Loki and Prometheus datasources. Operator support for Tempo and traces is coming soon.
