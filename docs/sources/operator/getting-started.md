---
aliases:
- /docs/grafana-cloud/agent/operator/getting-started/
- /docs/grafana-cloud/monitor-infrastructure/agent/operator/getting-started/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/operator/getting-started/
- /docs/grafana-cloud/send-data/agent/operator/getting-started/
canonical: https://grafana.com/docs/agent/latest/operator/getting-started/
description: Learn how to install the Operator
title: Install the Operator
weight: 110
---

# Install the Operator

In this guide, you'll learn how to deploy [Grafana Agent Operator]({{< relref "./_index.md" >}}) into your Kubernetes cluster. This guide does not use Helm. To learn how to deploy Agent Operator using the [grafana-agent-operator Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/agent-operator), see [Install Grafana Agent Operator with Helm]({{< relref "./helm-getting-started.md" >}}).

> **Note**: If you are shipping your data to Grafana Cloud, use [Kubernetes Monitoring](/docs/grafana-cloud/kubernetes-monitoring/) to set up Agent Operator. Kubernetes Monitoring provides a simplified approach and preconfigured dashboards and alerts.
## Before you begin

To deploy Agent Operator, make sure that you have the following:

- A Kubernetes cluster
- The `kubectl` command-line client installed and configured on your machine

> **Note:** Agent Operator is currently in beta and its custom resources are subject to change.

## Deploy the Agent Operator Custom Resource Definitions (CRDs)

Before you can create the custom resources for a Grafana Agent deployment,
you need to deploy the
[Custom Resource Definitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
to the cluster. These definitions describe the schema that the custom
resources will conform to. This is also required for Grafana Agent Operator to run; it
will fail if it can't find the Custom Resource Definitions of objects it is
looking to use. To learn more about the custom resources Agent Operator provides and their hierarchy, see [Grafana Agent Operator architecture]({{< relref "./architecture" >}}).

You can find the set of Custom Resource Definitions for Grafana Agent Operator in the Grafana Agent repository under
[`operations/agent-static-operator/crds`](https://github.com/grafana/agent/tree/main/operations/agent-static-operator/crds).

To deploy the CRDs:

1. Clone the agent repo and then apply the CRDs from the root of the agent repository:
    ```
    kubectl apply -f production/operator/crds
    ```

    This step _must_ be completed before installing Agent Operator&mdash;it will
fail to start if the CRDs do not exist.

2. To check that the CRDs are deployed to your Kubernetes cluster and to access documentation for each resource, use `kubectl explain <resource>`.

    For example, `kubectl explain GrafanaAgent` describes the GrafanaAgent CRD, and `kubectl explain GrafanaAgent.spec` gives you information on its spec field.

## Install Grafana Agent Operator

Next, install Agent Operator by applying the Agent Operator deployment schema.

To install Agent Operator:

1. Copy the following deployment schema to a file, updating the namespace if needed:

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
            image: grafana/agent-operator:{{< param "AGENT_RELEASE" >}}
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

2. Roll out the deployment in your cluster using `kubectl apply -f` followed by your  deployment filename.

> **Note**: If you want to run Agent Operator locally, make sure your kubectl context is correct. Running locally uses your current kubectl context. If it is set to your production environment, you could accidentally deploy a new Grafana Agent to production. Install CRDs on the cluster prior to running locally. Afterwards, you can run Agent Operator using `go run ./cmd/grafana-agent-operator`.

## Deploy the Grafana Agent Operator resources

Agent Operator is now up and running. Next, you need to install a Grafana Agent for Agent Operator to run for you. To do so, follow the instructions in the [Deploy the Grafana Agent Operator resources]({{< relref "./deploy-agent-operator-resources" >}}) topic.
