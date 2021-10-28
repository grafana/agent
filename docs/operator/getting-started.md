+++
title = "Get started with Grafana Agent Operator"
weight = 100
+++

# Get started with Grafana Agent Operator

An official Helm chart is planned to make it really easy to deploy the Grafana Agent
Operator on Kubernetes. For now, things must be done a little manually.

## Deploy CustomResourceDefinitions

Before you can write custom resources to describe a Grafana Agent deployment,
you _must_ deploy the
[CustomResourceDefinitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
to the cluster first. These definitions describe the schema that the custom
resources will conform to. This is also required for the operator to run; it
will fail if it can't find the custom resource definitions of objects it is
looking to use.

The current set of CustomResourceDefinitions can be found in
[production/operator/crds](../../production/operator/crds). Apply them from the
root of this repository using:

```
kubectl apply -f production/operator/crds
```

This step _must_ be done before installing the Operator, as the Operator will
fail to start if the CRDs do not exist.

### Find information on the supported values for the CustomResourceDefinitions

Once you've deployed the CustomResourceDefinitions
to your Kubernetes cluster, use `kubectl explain <resource>` to get access to
the documentation for each resource. For example, `kubectl explain GrafanaAgent`
will describe the GrafanaAgent CRD, and `kubectl explain GrafanaAgent.spec` will
give you information on its spec field.

## Install Agent Operator on Kubernetes

Use the following deployment to run the Operator, changing values as desired:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana-agent-operator
  namespace: default
  labels:
    app: grafana-agent-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana-agent-operator
  template:
    metadata:
      labels:
        app: grafana-agent-operator
    spec:
      serviceAccountName: grafana-agent-operator
      containers:
      - name: operator
        image: grafana/agent-operator:v0.20.0
        args:
        - --kubelet-service=default/kubelet
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: grafana-agent-operator
  namespace: default

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: grafana-agent-operator
rules:
- apiGroups: [monitoring.grafana.com]
  resources:
  - grafanaagents
  - metricsinstances
  - logsinstances
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
  - nodes
  verbs: [get, list, watch]
- apiGroups: [""]
  resources:
  - secrets
  - services
  - configmaps
  - endpoints
  verbs: [get, list, watch, create, update, patch, delete]
- apiGroups: ["apps"]
  resources:
  - statefulsets
  - daemonsets
  verbs: [get, list, watch, create, update, patch, delete]

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: grafana-agent-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: grafana-agent-operator
subjects:
- kind: ServiceAccount
  name: grafana-agent-operator
  namespace: default
```

## Run Operator locally

Before running locally, _make sure your kubectl context is correct!_
Running locally uses your current kubectl context, and you probably don't want
to accidentally deploy a new Grafana Agent to prod.

CRDs should be installed on the cluster prior to running locally. If you haven't
done this yet, follow [deploying CustomResourceDefinitions](#deploying-customresourcedefinitions)
first.

Afterwards, you can run the operator using `go run`:

```
go run ./cmd/agent-operator
```

## Deploy GrafanaAgent

Now that the Operator is running, you can create a deployment of the
Grafana Agent. The first step is to create a GrafanaAgent resource. This
resource will discover a set of MetricsInstance resources. You can use
this example, which creates a GrafanaAgent and the appropriate ServiceAccount
for you:

```yaml
apiVersion: monitoring.grafana.com/v1alpha1
kind: GrafanaAgent
metadata:
  name: grafana-agent
  namespace: default
  labels:
    app: grafana-agent
spec:
  image: grafana/agent:v0.20.0
  logLevel: info
  serviceAccountName: grafana-agent
  metrics:
    instanceSelector:
      matchLabels:
        agent: grafana-agent

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

Note that this searches for MetricsInstances in the same namespace with the
label matching `agent: grafana-agent`. A MetricsInstance is a custom resource
that describes where to write collected metrics. Use this one as an example:

```yaml
apiVersion: monitoring.grafana.com/v1alpha1
kind: MetricsInstance
metadata:
  name: primary
  namespace: default
  labels:
    agent: grafana-agent
spec:
  remoteWrite:
  - url: https://prometheus-us-central1.grafana.net/api/prom/push
    basicAuth:
      username:
        name: primary-credentials
        key: username
      password:
        name: primary-credentials
        key: password

  # Supply an empty namespace selector to look in all namespaces. Remove
  # this to only look in the same namespace.
  serviceMonitorNamespaceSelector: {}
  serviceMonitorSelector:
    matchLabels:
      instance: primary

  # Supply an empty namespace selector to look in all namespaces. Remove
  # this to only look in the same namespace.
  podMonitorNamespaceSelector: {}
  podMonitorSelector:
    matchLabels:
      instance: primary

  # Supply an empty namespace selector to look in all namespaces. Remove
  # this to only look in the same namespace.
  probeNamespaceSelector: {}
  probeSelector:
    matchLabels:
      instance: primary
```

Replace the remoteWrite URL to match your vendor. If your vendor doesn't need
credentials, you may remove the `basicAuth` section. Otherwise, configure a
secret with the base64-encoded values of the username and password:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: primary-credentials
  namespace: default
data:
  username: BASE64_ENCODED_USERNAME
  password: BASE64_ENCODED_PASSWORD
```

The above configuration of MetricsInstance will discover all
[PodMonitors](https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/api.md#podmonitor),
[Probes](https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/api.md#probe),
and [ServiceMonitors](https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/api.md#servicemonitor)
with a label matching `instance: primary`. Create resources as appropriate for
your environment.

As an example, here is a ServiceMonitor that can collect metrics from `kube-dns`:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kube-dns
  namespace: kube-system
  labels:
    instance: primary
spec:
  selector:
    matchLabels:
      k8s-app: kube-dns
  endpoints:
  - port: metrics
```

## Monitor Kubelets

The `--kubelet-service=default/kubelet` argument passed to the Grafana Agent
Operator container tells the Operator to manage a Service called `kubelet` in
the `default` namespace. The Service will have one endpoint created per Node in
the cluster, allowing a ServiceMonitor to scrape both Kubelet and cAdvisor
metrics.

Use the following ServiceMonitor as a base to collect Kubelet and cAdvisor
metrics:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app.kubernetes.io/name: kubelet
    instance: primary
  name: kubelet
  namespace: default
spec:
  endpoints:
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    honorLabels: true
    interval: 30s
    metricRelabelings:
    - action: drop
      regex: kubelet_(pod_worker_latency_microseconds|pod_start_latency_microseconds|cgroup_manager_latency_microseconds|pod_worker_start_latency_microseconds|pleg_relist_latency_microseconds|pleg_relist_interval_microseconds|runtime_operations|runtime_operations_latency_microseconds|runtime_operations_errors|eviction_stats_age_microseconds|device_plugin_registration_count|device_plugin_alloc_latency_microseconds|network_plugin_operations_latency_microseconds)
      sourceLabels:
      - __name__
    - action: drop
      regex: scheduler_(e2e_scheduling_latency_microseconds|scheduling_algorithm_predicate_evaluation|scheduling_algorithm_priority_evaluation|scheduling_algorithm_preemption_evaluation|scheduling_algorithm_latency_microseconds|binding_latency_microseconds|scheduling_latency_seconds)
      sourceLabels:
      - __name__
    - action: drop
      regex: apiserver_(request_count|request_latencies|request_latencies_summary|dropped_requests|storage_data_key_generation_latencies_microseconds|storage_transformation_failures_total|storage_transformation_latencies_microseconds|proxy_tunnel_sync_latency_secs)
      sourceLabels:
      - __name__
    - action: drop
      regex: kubelet_docker_(operations|operations_latency_microseconds|operations_errors|operations_timeout)
      sourceLabels:
      - __name__
    - action: drop
      regex: reflector_(items_per_list|items_per_watch|list_duration_seconds|lists_total|short_watches_total|watch_duration_seconds|watches_total)
      sourceLabels:
      - __name__
    - action: drop
      regex: etcd_(helper_cache_hit_count|helper_cache_miss_count|helper_cache_entry_count|object_counts|request_cache_get_latencies_summary|request_cache_add_latencies_summary|request_latencies_summary)
      sourceLabels:
      - __name__
    - action: drop
      regex: transformation_(transformation_latencies_microseconds|failures_total)
      sourceLabels:
      - __name__
    - action: drop
      regex: (admission_quota_controller_adds|admission_quota_controller_depth|admission_quota_controller_longest_running_processor_microseconds|admission_quota_controller_queue_latency|admission_quota_controller_unfinished_work_seconds|admission_quota_controller_work_duration|APIServiceOpenAPIAggregationControllerQueue1_adds|APIServiceOpenAPIAggregationControllerQueue1_depth|APIServiceOpenAPIAggregationControllerQueue1_longest_running_processor_microseconds|APIServiceOpenAPIAggregationControllerQueue1_queue_latency|APIServiceOpenAPIAggregationControllerQueue1_retries|APIServiceOpenAPIAggregationControllerQueue1_unfinished_work_seconds|APIServiceOpenAPIAggregationControllerQueue1_work_duration|APIServiceRegistrationController_adds|APIServiceRegistrationController_depth|APIServiceRegistrationController_longest_running_processor_microseconds|APIServiceRegistrationController_queue_latency|APIServiceRegistrationController_retries|APIServiceRegistrationController_unfinished_work_seconds|APIServiceRegistrationController_work_duration|autoregister_adds|autoregister_depth|autoregister_longest_running_processor_microseconds|autoregister_queue_latency|autoregister_retries|autoregister_unfinished_work_seconds|autoregister_work_duration|AvailableConditionController_adds|AvailableConditionController_depth|AvailableConditionController_longest_running_processor_microseconds|AvailableConditionController_queue_latency|AvailableConditionController_retries|AvailableConditionController_unfinished_work_seconds|AvailableConditionController_work_duration|crd_autoregistration_controller_adds|crd_autoregistration_controller_depth|crd_autoregistration_controller_longest_running_processor_microseconds|crd_autoregistration_controller_queue_latency|crd_autoregistration_controller_retries|crd_autoregistration_controller_unfinished_work_seconds|crd_autoregistration_controller_work_duration|crdEstablishing_adds|crdEstablishing_depth|crdEstablishing_longest_running_processor_microseconds|crdEstablishing_queue_latency|crdEstablishing_retries|crdEstablishing_unfinished_work_seconds|crdEstablishing_work_duration|crd_finalizer_adds|crd_finalizer_depth|crd_finalizer_longest_running_processor_microseconds|crd_finalizer_queue_latency|crd_finalizer_retries|crd_finalizer_unfinished_work_seconds|crd_finalizer_work_duration|crd_naming_condition_controller_adds|crd_naming_condition_controller_depth|crd_naming_condition_controller_longest_running_processor_microseconds|crd_naming_condition_controller_queue_latency|crd_naming_condition_controller_retries|crd_naming_condition_controller_unfinished_work_seconds|crd_naming_condition_controller_work_duration|crd_openapi_controller_adds|crd_openapi_controller_depth|crd_openapi_controller_longest_running_processor_microseconds|crd_openapi_controller_queue_latency|crd_openapi_controller_retries|crd_openapi_controller_unfinished_work_seconds|crd_openapi_controller_work_duration|DiscoveryController_adds|DiscoveryController_depth|DiscoveryController_longest_running_processor_microseconds|DiscoveryController_queue_latency|DiscoveryController_retries|DiscoveryController_unfinished_work_seconds|DiscoveryController_work_duration|kubeproxy_sync_proxy_rules_latency_microseconds|non_structural_schema_condition_controller_adds|non_structural_schema_condition_controller_depth|non_structural_schema_condition_controller_longest_running_processor_microseconds|non_structural_schema_condition_controller_queue_latency|non_structural_schema_condition_controller_retries|non_structural_schema_condition_controller_unfinished_work_seconds|non_structural_schema_condition_controller_work_duration|rest_client_request_latency_seconds|storage_operation_errors_total|storage_operation_status_count)
      sourceLabels:
      - __name__
    port: https-metrics
    relabelings:
    - sourceLabels:
      - __metrics_path__
      targetLabel: metrics_path
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    honorLabels: true
    honorTimestamps: false
    interval: 30s
    metricRelabelings:
    - action: drop
      regex: container_(network_tcp_usage_total|network_udp_usage_total|tasks_state|cpu_load_average_10s)
      sourceLabels:
      - __name__
    - action: drop
      regex: (container_fs_.*|container_spec_.*|container_blkio_device_usage_total|container_file_descriptors|container_sockets|container_threads_max|container_threads|container_start_time_seconds|container_last_seen);;
      sourceLabels:
      - __name__
      - pod
      - namespace
    path: /metrics/cadvisor
    port: https-metrics
    relabelings:
    - sourceLabels:
      - __metrics_path__
      targetLabel: metrics_path
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
  jobLabel: app.kubernetes.io/name
  namespaceSelector:
    matchNames:
    - default
  selector:
    matchLabels:
      app.kubernetes.io/name: kubelet
```
