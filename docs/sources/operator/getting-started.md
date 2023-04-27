---
title: Installing Grafana Agent Operator
weight: 100
---

# Installing Grafana Agent Operator

In this guide you'll learn how to deploy the [Grafana Agent Operator]({{< relref "./_index.md" >}}) into your Kubernetes cluster. This guide does *not* use Helm. To learn how to deploy Agent Operator using the [grafana-agent-operator Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/agent-operator), please see [Installing Grafana Agent Operator with Helm]({{< relref "./helm-getting-started.md" >}}).

> **Note:** Agent Operator is currently in beta and its custom resources are subject to change as the project evolves. It currently supports the metrics and logs subsystems of Grafana Agent. Integrations and traces support is coming soon.

By the end of this guide, you'll have deloyed Agent Operator into your cluster.

## Prerequisites

Before you begin, make sure that you have the following available to you:

- A Kubernetes cluster
- The `kubectl` command-line client installed and configured on your machine

## Step 1: Deploy CustomResourceDefinitions

Before you can write custom resources to describe a Grafana Agent deployment,
you _must_ deploy the
[CustomResourceDefinitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
to the cluster first. These definitions describe the schema that the custom
resources will conform to. This is also required for the operator to run; it
will fail if it can't find the custom resource definitions of objects it is
looking to use.

The current set of CustomResourceDefinitions can be found in
[production/operator/crds](https://github.com/grafana/agent/tree/main/production/operator/crds). Apply them from the
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

## Step 2: Install Agent Operator

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
        image: grafana/agent-operator:v0.26.1
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
  - integrations
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
  - deployments
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

### Run Operator locally

Before running locally, _make sure your kubectl context is correct!_
Running locally uses your current kubectl context, and you probably don't want
to accidentally deploy a new Grafana Agent to prod.

CRDs should be installed on the cluster prior to running locally. If you haven't
done this yet, follow [deploying CustomResourceDefinitions](#step-1-deploy-customresourcedefinitions)
first.

Afterwards, you can run the operator using `go run`:

```
go run ./cmd/agent-operator
```

## Conclusion

With Agent Operator up and running, you can move on to setting up a `GrafanaAgent` custom resource. This will discover `MetricsInstance` and `LogsInstance` custom resources and endow them with Pod attributes (like requests and limits) defined in the `GrafanaAgent` spec. To learn how to do this, please see [Custom Resource Quickstart]({{< relref "./custom-resource-quickstart.md" >}}).
