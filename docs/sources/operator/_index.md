---
title: Grafana Agent Operator
weight: 300
---

# Grafana Agent Operator

The Grafana Agent Operator is a [Kubernetes operator](https://www.cncf.io/blog/2022/06/15/kubernetes-operators-what-are-they-some-examples/) that makes it easier to
deploy [Grafana Agent]({{< relref "../_index.md" >}}) and collect telemetry data from your [Pods](https://kubernetes.io/docs/concepts/workloads/pods/).
Agent Operator is currently in **Beta**, and is subject to change.

> **Note**: If you are shipping your data to Grafana Cloud, use [Kubernetes Monitoring](https://grafana.com/docs/grafana-cloud/kubernetes-monitoring/) to set up Agent Operator. Kubernetes Monitoring provides a simplified approach and preconfigured dashboards and alerts.

Grafana Agent Operator uses Kubernetes [custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) to simplify the deployment and configuration of Grafana Agents. Agent Operator installs and manages Agents, and dynamically watches resources on your Kubernetes clusters, helping to discover Pods, [Services](https://kubernetes.io/docs/concepts/services-networking/service/), and [Ingresses](https://kubernetes.io/docs/concepts/services-networking/ingress/) to scrape. This dynamic, declarative approach works well for decentralized deployments. You can provide your own Service Discovery (SD) scrape configs to indicate how your Pods should be monitored without explicitly defining the entire monitoring configuration yourself.

Metric collection is based on the [Prometheus
Operator](https://github.com/prometheus-operator/prometheus-operator) and
supports the official v1 ServiceMonitor, PodMonitor, and Probe CRDs from the
project. These custom resources represent abstractions for monitoring Services,
Pods, and Ingresses. They are especially useful for Helm users, where manually
writing a generic SD to match all your charts can be difficult or where manually writing a specific SD for each chart can be tedious.

The following sections describe how to use Grafana Agent Operator.

| Topic | Describes |
|---|---|
| [Install Grafana Agent Operator with Helm]({{< relref "./helm-getting-started/" >}}) | How to deploy the Grafana Agent Operator into your Kubernetes cluster using the grafana-agent-operator Helm chart. |
| [Install Grafana Agent Operator]({{< relref "./getting-started/" >}}) | How to deploy the Grafana Agent Operator into your Kubernetes cluster without using Helm. |
| [Deploy the Grafana Agent Operator resources]({{< relref "./deploy-agent-operator-resources/" >}}) | How to roll out the Grafana Agent Operator custom resources, needed to begin monitoring your cluster. Complete this procedure *after* installing Grafana Agent Operator&mdash;either with or without Helm. |
| [Grafana Agent Operator architecture]({{< relref "./architecture/" >}}) | Learn about the resources used by Agent Operator to collect telemetry data and how it discovers the hierarchy of custom resources, continually reconciling the hierarchy.  |
| [Set up Agent Operator integrations]({{< relref "./operator-integrations/" >}}) | Learn how to set up node-exporter and mysqld-exporter integrations. |
