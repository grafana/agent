---
aliases:
- /docs/agent/latest/flow/tutorials/chaining
title: Chaining Prometheus components
weight: 400
---

This tutorial shows how to use [multiple-inputs.flow](../assets/flow_configs/multiple-inputs.flow) to send data to several different locations. This tutorial uses the same base as [Filtering metrics]({{< relref "filtering-metrics.md">}}). 

A new concept introduced in Flow is chaining components together in a composable pipeline. This promotes the reusability of components while offering flexibility. 

# Prerequisites

* [Docker](https://www.docker.com/products/docker-desktop)
* Clone the [Agent Repository](https://github.com/grafana/agent) 
    * `git clone git@github.com:grafana/agent.git`

## Example

To run the example, execute `CONFIG_FILE=multiple-inputs.flow docker-compose -f ./assets/docker-compose.yaml up` from the tutorials directory. 

Allow the Agent to run for two minutes, then go to [Grafana](http://localhost:3000/explore?orgId=1&left=%5B%22now-1h%22,%22now%22,%22Mimir%22,%7B%22refId%22:%22A%22,%22instant%22:true,%22range%22:true,%22exemplar%22:true,%22expr%22:%22agent_build_info%7B%7D%22%7D%5D).

![](../assets/multiple.png)

There are two scrapes each sending two metrics to both filters, so you end up with a total of four metrics. They differ in `job` and `cool_label/not_cool_label`.

## Multiple outputs

```river
prometheus.scrape "first" {
	targets    = [{"__address__" = "localhost:12345"}]
	forward_to = [prometheus.relabel.cool.receiver, prometheus.relabel.not_cool.receiver]
}
```

In the above Flow block, `forward_to` accepts an array of receivers. In previous examples, a single receiver was used, but the use of multiple inputs and outputs is a basic building block of Flow. In the above example, `prometheus.scrape.first` is sending to both `prometheus.relabel.filter` and `prometheus.relabel.not_cool`. 

## Multiple inputs

```river
prometheus.relabel "cool" {
    metric_relabel_config {
        source_labels = ["__name__"]
        regex = "(.+)"
        replacement = "${1}_cool"
        target_label = "cool_label"
    }
    forward_to = [prometheus.remote_write.prom.receiver]
}

prometheus.relabel "not_cool" {
    metric_relabel_config {
        source_labels = ["__name__"]
        regex = "(.+)"
        replacement = "${1}_not_cool"
        target_label = "not_cool_label"
    }
    forward_to = [prometheus.remote_write.prom.receiver]
}

prometheus.remote_write "prom" {
    remote_write {
        url = "http://mimir:9009/api/v1/push"
    }
}
```

In the above Flow blocks, `prometheus.remote_write.prom` receives input from both `prometheus.relabel.cool` and `prometheus.relabel.not_cool`. 