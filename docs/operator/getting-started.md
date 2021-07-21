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
        image: grafana/agent-operator:v0.16.1
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
  - grafana-agents
  - prometheus-instances
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
resource will discover a set of PrometheusInstance resources. You can use
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
  image: grafana/agent:v0.15.0
  logLevel: info
  serviceAccountName: grafana-agent
  prometheus:
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
  - services
  - endpoints
  - pods
  verbs:
  - get
  - list
  - watch
- nonResourceURLs:
  - /metrics
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

Note that this searches for PrometheusInstances in the same namespace with the
label matching `agent: grafana-agent`. A PrometheusInstance is a custom resource
that describes where to write collected metrics. Use this one as an example:

```yaml
apiVersion: monitoring.grafana.com/v1alpha1
kind: PrometheusInstance
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

The above configuration of PrometheusInstance will discover all
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
