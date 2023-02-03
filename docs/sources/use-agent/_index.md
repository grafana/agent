---
title: How to use Grafana Agent
weight: 350
---

# How to use Grafana Agent

Grafana Agent is a telemetry collector for sending metrics, logs, and trace data to the opinionated Grafana observability stack. It works with:

- Grafana Cloud
- Grafana Enterprise Stack
- OSS deployments of Grafana Loki, Grafana Mimir, Grafana Tempo, and Prometheus 

This topic helps you to think about what you're trying to accomplish and how to use Grafana Agent to meet your goals. 
## Use Grafana Agent with Grafana Cloud integrations to scrape telemetry data

There are different ways for you to set up Grafana Agent to scrape data&mdash;through Grafana's integration platform or directly. Select a guide to get started:

| Topic | Description |
|---|---|
| [Get started with monitoring using an integration](/docs/grafana-cloud/latest/data-configuration/get-started-integration/) | Walk through installing a Linux integration using Grafana Agent in the Grafana Cloud interface. |
| [Install and manage integrations](/docs/grafana-cloud/latest/data-configuration/integrations/install-and-manage-integrations/)  | View general steps for using Grafana Cloud integrations to install Grafana Agent to collect data. See [supported integrations](/docs/grafana-cloud/latest/data-configuration/integrations/integration-reference/).  
| [Ship your metrics to Grafana Cloud without an integration](/docs/grafana-cloud/latest/data-configuration/metrics/agent-config-exporter/) | If you want to ship your Prometheus metrics to Grafana Cloud but there isn’t an integration available, you can use a Prometheus exporter and deploy Grafana Agent to scrape your local machine or service. |

## Use Grafana Agent or Agent Operator for monitoring Kubernetes

| Topic | Description |
|---|---|
| Configure Kubernetes Monitoring using Agent |  |
| Ship Kubernetes metrics using Grafana Agent |   |

## Use Grafana Agent Operator for monitoring Kubernetes

## Use Grafana Agent for Kubernetes dire

## Use Grafana Agent directly to scrape telemetry data

Grafana Cloud integration workflows are the easiest way to get started collecting telemetry data, but sometimes you might want to use a manual approach to select your configuration options.

| Topic | Description |
|---|---|
| [Install Grafana Agent](/docs/grafana-cloud/latest/data-configuration/agent/install_agent/) | Install Grafana Agent using a script for Debian- and Red Hat-based systems. |
| [Manage Grafana Agent with systemd](/docs/grafana-cloud/latest/data-configuration/agent/agent_as_service/) |  Run Grafana Agent as a [systemd](https://www.freedesktop.org/wiki/Software/systemd/) service to create a long-living process that can automatically restart when killed or when the host is rebooted. |
| [Monitor Grafana Agent](/docs/grafana-cloud/latest/data-configuration/agent/agent_monitoring/) |  You use Grafana Agent to monitor services but you should also monitor Grafana Agent itself. Learn how to use PromQL to set up an alert for an Agent integration, as well as other methods to monitor Agent. |
| [Uninstall Grafana Agent](/docs/grafana-cloud/latest/data-configuration/agent/install_agent/#uninstall-grafana-agent) | Uninstalling an integration doesn't automatically stop Agent from scraping data. Learn how to uninstall Agent. |
| [Troubleshoot Grafana Agent](/docs/grafana-cloud/latest/data-configuration/agent/troubleshooting/) | Learn what to check when you are having trouble collecting data using Grafana Agent, and find solutions to common issues.  |

## Use Grafana Agent to send logs to Grafana Cloud

| Topic | Description |
|---|---|
| [Collect logs with Grafana Agent](/docs/grafana-cloud/latest/data-configuration/logs/collect-logs-with-agent/) |  Install Grafana Agent to collect logs for use with Grafana Loki, included with your [Grafana Cloud account](/docs/grafana-cloud/latest/account-management/cloud-portal/). Logs are also included when you [set up a Cloud integration]/docs/grafana-cloud/latest/data-configuration/integrations/install-and-manage-integrations/. |
| 