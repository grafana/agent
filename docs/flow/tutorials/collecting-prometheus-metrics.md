---
aliases:
- /docs/agent/latest/flow/tutorials/collecting-prometheus-metrics
title: Collecting Prometheus metrics
weight: 200
---

Grafana Agent is a telemetry collector with the primary goal of moving telemetry data from one location to another. In this tutorial, you'll set up a Grafana Agent in Flow mode.  

# Prerequisites

* [Docker](https://www.docker.com/products/docker-desktop)
* Clone the [Agent Repository](https://github.com/grafana/agent) `git clone git@github.com:grafana/agent.git`

# Scraping the Agent

To quickly spin up an example environment, run the following: `docker-compose -f ./assets/scraping-the-agent.yaml` with the [flow file](../assets/flow_configs/agent.flow). Allow the service to run for a few minutes, then navigate to the [metrics browser](http://localhost:3000/explore?orgId=1&left=%5B%22now-1h%22,%22now%22,%22Cortex%22,%7B%22refId%22:%22A%22,%22instant%22:true,%22range%22:true,%22exemplar%22:false,%22expr%22:%22agent_build_info%22%7D%5D). 

![](./assets/agent_build_info.png)

This example scrapes the Grafana Agent's `http://localhost:12345/metrics` endpoint and pushes those metrics to the cortex instance. 


## Scraping component

The [`prometheus.scrape`]({{< relref "prometheus.scrape.md" >}}) component is responsible for scraping the metrics of a particular endpoint and passing them on to another component.

```river
// prometheus.scrape is the name of component and "default" is the label of a component
prometheus.scrape "default" {
    // Tell the scraper to scrap at http://localhost:12345/metrics
    // the http:// and metrics are implied but able to be overwritten 
    targets = [{"__address__" = "localhost:12345"}]
    // Forward the scrape results to the receiver, in general 
    // Flow uses forward_to to tell which receiver to send results to. 
    // The forward_to is an argument of prometheus.scrape.default  and 
    // the receiver is an export prometheus.remote_write.prom
    forward_to = [prometheus.remote_write.prom.receiver]
}
```

The `prometheus.scrape "default"` annotation tells the name `prometheus.scrape` and the label `default`. All components must have a unique combination of name and label.

The `targets` [attribute]({{< relref "configuration_language.md#Attributes" >}}) is an [argument]({{< relref "../concepts/components.md">}}). Targets is a list of labels that specify the target via the special key `__address__`. The scraper is targeting the Agent's `/metrics` endpoint. Both the `http` and `/metrics` are implied but can be overridden.

The `forward_to` attribute is an argument that references the [export]({{< relref "../concepts/components.md">}}) of the `prometheus.remote_write.prom` component. This is where the scraper will send the metrics for further processing.

## Remote Write Component

The [`prometheus.remote_write`]({{< relref "prometheus.remote_write.md" >}}) is responsible for writing the metrics to a prometheus compatible endpoint (cortex).

```river
prometheus.remote_write "prom" {
    remote_write {
        url = "http://cortex:9009/api/prom/push"
    }
}
```

## Running without Docker

To try out the Grafana Agent without using Docker:
1. Download the Grafana Agent 
1. Set the environment variable `EXPERIMENTAL_ENABLE_FLOW=true`
1. Run the agent with `agent run <path_to_flow_config>`