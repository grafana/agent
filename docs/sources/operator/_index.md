---
aliases:
- /docs/grafana-cloud/agent/operator/
- /docs/grafana-cloud/monitor-infrastructure/agent/operator/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/operator/
- /docs/grafana-cloud/send-data/agent/operator/
canonical: https://grafana.com/docs/agent/latest/operator/
description: Learn about the static mode Kubernetes operator
menuTitle: Static mode Kubernetes operator
title: Static mode Kubernetes operator (Beta)
weight: 300
---

# Static mode Kubernetes operator (Beta)

Grafana Agent Operator is a [Kubernetes operator][] for the [static mode][] of
Grafana Agent. It makes it easier to deploy and configure static mode to
collect telemetry data from Kubernetes resources.

Grafana Agent Operator supports consuming various [custom resources][] for
telemetry collection:

* Prometheus Operator [ServiceMonitor][] resources for collecting metrics from Kubernetes [Services][].
* Prometheus Operator [PodMonitor][] resources for collecting metrics from Kubernetes [Pods][].
* Prometheus Operator [Probe][] resources for collecting metrics from Kubernetes [Ingresses][].
* Custom [PodLogs][] resources for collecting logs.

{{< admonition type="note" >}}
Grafana Agent Operator does not collect traces.
{{< /admonition >}}

Grafana Agent Operator is currently in [Beta][], and is subject to change or
being removed with functionality which covers the same use case.

{{< admonition type="note" >}}
If you are shipping your data to Grafana Cloud, use [Kubernetes Monitoring](/docs/grafana-cloud/kubernetes-monitoring/) to set up Agent Operator.
Kubernetes Monitoring provides a simplified approach and preconfigured dashboards and alerts.
{{< /admonition >}}

Grafana Agent Operator uses additional custom resources to manage the deployment
and configuration of Grafana Agents running in static mode. In addition to the
supported custom resources, you can also provide your own Service Discovery
(SD) configurations to collect metrics from other types of sources.

Grafana Agent Operator is particularly useful for Helm users, where manually
writing generic service discovery to match all of your chart installations can
be difficult, or where manually writing a specific SD for each chart
installation can be tedious.

The following sections describe how to use Grafana Agent Operator:

| Topic | Describes |
|---|---|
| [Configure Kubernetes Monitoring using Agent Operator](/docs/grafana-cloud/monitor-infrastructure/kubernetes-monitoring/configuration/configure-infrastructure-manually/k8s-agent-operator/) | Use the Kubernetes Monitoring solution to set up monitoring of your Kubernetes cluster and to install preconfigured dashboards and alerts. |
| [Install Grafana Agent Operator with Helm]({{< relref "./helm-getting-started" >}}) | How to deploy the Grafana Agent Operator into your Kubernetes cluster using the grafana-agent-operator Helm chart. |
| [Install Grafana Agent Operator]({{< relref "./getting-started" >}}) | How to deploy the Grafana Agent Operator into your Kubernetes cluster without using Helm. |
| [Deploy the Grafana Agent Operator resources]({{< relref "./deploy-agent-operator-resources" >}}) | How to roll out the Grafana Agent Operator custom resources, needed to begin monitoring your cluster. Complete this procedure *after* installing Grafana Agent Operator&mdash;either with or without Helm. |
| [Grafana Agent Operator architecture]({{< relref "./architecture" >}}) | Learn about the resources used by Agent Operator to collect telemetry data and how it discovers the hierarchy of custom resources, continually reconciling the hierarchy.  |
| [Set up Agent Operator integrations]({{< relref "./operator-integrations" >}}) | Learn how to set up node-exporter and mysqld-exporter integrations. |

[Kubernetes operator]: https://www.cncf.io/blog/2022/06/15/kubernetes-operators-what-are-they-some-examples/
[static mode]: {{< relref "../static/" >}}
[Services]: https://kubernetes.io/docs/concepts/services-networking/service/
[Pods]: https://kubernetes.io/docs/concepts/workloads/pods/
[Ingresses]: https://kubernetes.io/docs/concepts/services-networking/ingress/
[custom resources]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/
[Beta]: {{< relref "../stability.md#beta" >}}
[ServiceMonitor]: https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.ServiceMonitor
[PodMonitor]: https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.PodMonitor
[Probe]: https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.Probe
[PodLogs]: {{< relref "./api.md#podlogs-a-namemonitoringgrafanacomv1alpha1podlogsa">}}
[Prometheus Operator]: https://github.com/prometheus-operator/prometheus-operator
