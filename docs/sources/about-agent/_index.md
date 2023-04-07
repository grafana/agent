---
title: About Grafana Agent
weight: 50
---

# About Grafana Agent

Grafana Agent is a telemetry collector for sending metrics, logs, and traces to the opinionated Grafana observability stack. It works with:

- Grafana Cloud
- Grafana Enterprise Stack
- OSS deployments of Grafana Loki, Grafana Mimir, Grafana Tempo, and Prometheus

This topic helps you to think about what you're trying to accomplish and how to use Grafana Agent to meet your goals.

You can [set up]({{< relref "../static/set-up/" >}}) and [configure Grafana]({{< relref "../static/configuration/" >}}) manually or you can follow the common workflows described in this topic.

## Grafana Agent for Grafana Cloud integrations

There are different ways for you to set up Grafana Agent to scrape data&mdash;through Grafana's integration platform or directly.  Select a guide to get started:

| Topic | Description |
|---|---|
| [Get started with monitoring using an integration](/docs/grafana-cloud/data-configuration/get-started-integration/) | Walk through installing a Linux integration using Grafana Agent in the Grafana Cloud interface. |
| [Install and manage integrations](/docs/grafana-cloud/data-configuration/integrations/install-and-manage-integrations/)  | View general steps for using Grafana Cloud integrations to install Grafana Agent to collect data. See [supported integrations](/docs/grafana-cloud/data-configuration/integrations/integration-reference/).
| [Ship your metrics to Grafana Cloud without an integration](/docs/grafana-cloud/data-configuration/metrics/agent-config-exporter/) | If you want to ship your Prometheus metrics to Grafana Cloud but there isn’t an integration available, you can use a Prometheus exporter and deploy Grafana Agent to scrape your local machine or service. |
| [Change your metrics scrape interval](/docs/grafana-cloud/billing-and-usage/control-prometheus-metrics-usage/changing-scrape-interval/) | Learn about reducing your total data points per minute (DPM) by adjusting your scrape interval. |

## Grafana Agent for Kubernetes Monitoring

Grafana Kubernetes Monitoring provides a simplified approach to monitoring your Kubernetes fleet by deploying Grafana Agent with useful defaults for collecting metrics. Select a guide to get started monitoring Kubernetes:

| Topic | Description |
|---|---|
| [Configure Kubernetes Monitoring using Agent](/docs/grafana-cloud/kubernetes-monitoring/configuration/config-k8s-agent-guide/) | Use the Kubernetes Monitoring solution to set up monitoring of your Kubernetes cluster and to install preconfigured dashboards and alerts. |
| [Ship Kubernetes metrics using Grafana Agent directly](/docs/grafana-cloud/kubernetes-monitoring/other-methods/k8s-agent-metrics/) |  Take a more hands-on approach and directly deploy Grafana Agent into a Kubernetes cluster without using the Kubernetes Monitoring interface. Use this guide to configure Agent to scrape the kubelet and cadvisor endpoints on your cluster Nodes. If you use this method, you still have access to the Kubernetes Monitoring preconfigured dashboards and alerts. |
| [Ship Kubernetes logs using Grafana Agent directly](/docs/grafana-cloud/kubernetes-monitoring/other-methods/k8s-agent-logs/) | Deploy Grafana Agent into your Kubernetes cluster as a DaemonSet and configure it to collect logs for your Kubernetes workloads.  |
| [Ship Kubernetes traces using Grafana Agent directly](/docs/grafana-cloud/kubernetes-monitoring/other-methods/k8s-agent-traces/) | Deploy Grafana Agent into your Kubernetes cluster as a deployment and configure it to collect traces for your Kubernetes workloads.  |

## Grafana Agent Operator for Kubernetes Monitoring

You can use Kubernetes Monitoring with Grafana Agent or with Grafana Agent Operator. Use Grafana Agent Operator if you prefer to use a Kubernetes-style operator rather than an agent. See [Grafana Agent versus Grafana Agent Operator](/docs/grafana-cloud/kubernetes-monitoring/#grafana-agent-versus-grafana-agent-operator) to understand the differences.

| Topic | Description |
|---|---|
| [Configure Kubernetes Monitoring using Agent Operator](/docs/grafana-cloud/kubernetes-monitoring/configuration/config-k8s-agent-operator-guide/) | Use the Kubernetes Monitoring solution to set up monitoring of your Kubernetes cluster and to install preconfigured dashboards and alerts. |
| [Ship Kubernetes metrics directly using Agent Operator]({{< relref "../operator/getting-started/" >}}) |  Take a hands-on approach and directly deploy Grafana Agent Operator into your Kubernetes cluster. After you deploy Agent Operator, you then [deploy the Agent Operator resources]({{< relref "../operator/deploy-agent-operator-resources/" >}}) to begin collecting telemetry data.|
| [Ship Kubernetes metrics directly using Agent Operator with Helm]({{< relref "../operator/helm-getting-started/" >}}) |  Take a hands-on approach and directly deploy Grafana Agent Operator into your Kubernetes cluster using the [grafana-agent-operator Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/agent-operator). After you deploy Agent Operator, you then [deploy the Agent Operator resources]({{< relref "../operator/deploy-agent-operator-resources/" >}}) and begin collecting telemetry data. |

## Use Grafana Agent directly to scrape telemetry data

Grafana Cloud integration workflows and the Kubernetes Monitoring solution are the easiest ways to get started collecting telemetry data, but sometimes you might want to use a manual approach to set your configuration options.

| Topic | Description |
|---|---|
| [Install Grafana Agent](/docs/grafana-cloud/data-configuration/agent/install_agent/) | Install Grafana Agent using a script for Debian- and Red Hat-based systems. |
| [Manage Grafana Agent with systemd](/docs/grafana-cloud/data-configuration/agent/agent_as_service/) |  Run Grafana Agent as a [systemd](https://www.freedesktop.org/wiki/Software/systemd/) service to create a long-living process that can automatically restart when killed or when the host is rebooted. |
| [Monitor Grafana Agent](/docs/grafana-cloud/data-configuration/agent/agent_monitoring/) |  Grafana Agent lets you monitor services but you can also monitor Grafana Agent itself. Learn how to use PromQL to set up an alert for an Agent integration, as well as other methods to monitor Agent. |
| [Uninstall Grafana Agent](/docs/grafana-cloud/data-configuration/agent/install_agent/#uninstall-grafana-agent) | Uninstalling an integration doesn't automatically stop Agent from scraping data. Learn how to uninstall Agent. |
| [Troubleshoot Grafana Agent](/docs/grafana-cloud/data-configuration/agent/troubleshooting/) | Learn what to check when you are having trouble collecting data using Grafana Agent, and find solutions to common issues.  |

## Use Grafana Agent to send logs to Grafana Loki

Logs are included when you [set up a Cloud integration](/docs/grafana-cloud/data-configuration/integrations/install-and-manage-integrations) but you can take a more hands-on approach with the following guide.

| Topic | Description |
|---|---|
| [Collect logs with Grafana Agent](/docs/grafana-cloud/data-configuration/logs/collect-logs-with-agent/) |  Install Grafana Agent to collect logs for use with Grafana Loki, included with your [Grafana Cloud account](/docs/grafana-cloud/account-management/cloud-portal/). |

## Use Grafana Agent to send traces to Grafana Tempo

| Topic | Description |
|---|---|
| [Set up and use tracing](/docs/grafana-cloud/data-configuration/traces/set-up-and-use-tempo/) |  Install Grafana Agent to collect traces for use with Grafana Tempo, included with your [Grafana Cloud account](/docs/grafana-cloud/account-management/cloud-portal/). |
| [Use Grafana Agent as a tracing pipeline](/docs/tempo/latest/grafana-agent/) | Grafana Agent can be configured to run a set of tracing pipelines to collect data from your applications and write it to Grafana Tempo. Pipelines are built using OpenTelemetry, and consist of receivers, processors, and exporters. |
