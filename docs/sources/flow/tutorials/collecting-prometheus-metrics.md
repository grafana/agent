---
canonical: https://grafana.com/docs/agent/latest/flow/tutorials/collecting-prometheus-metrics/
title: Collecting Prometheus metrics
weight: 200
---

# Collecting Prometheus metrics

Grafana Agent is a telemetry collector with the primary goal of moving telemetry data from one location to another. In this tutorial, you'll set up a Grafana Agent in Flow mode.

## Prerequisites

* [Docker](https://www.docker.com/products/docker-desktop)

## Run the example

Run the following `curl https://raw.githubusercontent.com/grafana/agent/main/docs/sources/flow/tutorials/assets/runt.sh -O && bash ./runt.sh agent.river`.

The `runt.sh` script does:

1. Downloads the configs necessary for Mimir, Grafana and the Grafana Agent.
2. Downloads the docker image for Grafana Agent explicitly.
3. Runs the docker-compose up command to bring all the services up.

Allow the Grafana Agent to run for two minutes, then navigate to [Grafana](http://localhost:3000/explore?orgId=1&left=%5B%22now-1h%22,%22now%22,%22Mimir%22,%7B%22refId%22:%22A%22,%22instant%22:true,%22range%22:true,%22exemplar%22:true,%22expr%22:%22agent_build_info%7B%7D%22%7D%5D).

![](../assets/agent_build_info.png)

This example scrapes the Grafana Agent's `http://localhost:12345/metrics` endpoint and pushes those metrics to the Mimir instance.

Navigate to `http://localhost:12345/graph` to view the Grafana Agent Flow UI.

![](../assets/graph.png)

The Agent displays the component pipeline in a dependency graph.  See [Scraping component](#scraping-component) and [Remote Write component](#remote-write-component) for details about the components used in this configuration.
Click the nodes to navigate to the associated component page. There, you can view the state, health information, and, if applicable, the debug information.

![](../assets/comp_info.png)

## Scraping component

The [`prometheus.scrape`]({{< relref "../reference/components/prometheus.scrape.md" >}}) component is responsible for scraping the metrics of a particular endpoint and passing them on to another component.

```river
// prometheus.scrape is the name of the component and "default" is its label.
prometheus.scrape "default" {
    // Tell the scraper to scrape at http://localhost:12345/metrics.
    // The http:// and metrics are implied but able to be overwritten.
    targets = [{"__address__" = "localhost:12345"}]
    // Forward the scrape results to the receiver. In general,
    // Flow uses forward_to to tell which receiver to send results to.
    // The forward_to is an argument of prometheus.scrape.default and
    // the receiver is an exported field of prometheus.remote_write.prom.
    forward_to = [prometheus.remote_write.prom.receiver]
}
```

The `prometheus.scrape "default"` annotation indicates the name of the component, `prometheus.scrape`, and its label, `default`. All components must have a unique combination of name and if applicable label.

The `targets` [attribute]({{< relref "../concepts/configuration_language/#attributes" >}}) is an [argument]({{< relref "../concepts/components.md">}}). `targets` is a list of labels that specify the target via the special key `__address__`. The scraper is targeting the Agent's `/metrics` endpoint. Both `http` and `/metrics` are implied but can be overridden.

The `forward_to` attribute is an argument that references the [export]({{< relref "../concepts/components.md">}}) of the `prometheus.remote_write.prom` component. This is where the scraper will send the metrics for further processing.

## Remote Write component

The [`prometheus.remote_write`]({{< relref "../reference/components/prometheus.remote_write.md" >}}) component is responsible for writing the metrics to a Prometheus-compatible endpoint (Mimir).

```river
prometheus.remote_write "prom" {
    endpoint {
        url = "http://mimir:9009/api/v1/push"
    }
}
```

## Running without Docker

To try out the Grafana Agent without using Docker:
1. Download the Grafana Agent.
1. Set the environment variable `AGENT_MODE=flow`.
1. Run the agent with `grafana-agent run <path_to_flow_config>`.
