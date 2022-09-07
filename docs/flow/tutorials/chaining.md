---
aliases:
- /docs/agent/latest/flow/tutorials/chaining
title: Chaining components
weight: 400
---

The goal of this tutorial is to show the ability to have data go to several different locations using [multiple-inputs.flow](../assets/flow_configs/multiple-inputs.flow). This builds upon [Relabel Component]({{< relref "relabel-component.md">}}) and requires all the other previous prerequisites.

A new concept in Flow is chaining components together. 

## Example

To run the example execute `docker-compose -f ./assets/multiple-inputs.yaml up` from the tutorials directory. 

Allow the Agent to run for two minutes then go to [Grafana](http://localhost:3000/explore?orgId=1&left=%5B%22now-1m%22,%22now%22,%22Cortex%22,%7B%22refId%22:%22A%22,%22instant%22:true,%22range%22:true,%22exemplar%22:false,%22expr%22:%22agent_build_info%22%7D%5D).

![](./assets/multiple.png)

There are 4 series: Two scrapers each sending metrics to both filters so you end up with 4 series. They differ in `job` and `cool_label/not_cool_label`.

## Multiple Outputs

```river
prometheus.scrape "first" {
    targets = [{"__address__" = "localhost:12345"}]
    forward_to = [prometheus.relabel.filter.receiver,prometheus.relabel.not_cool.receiver]
    scrape_config {
        job_name = "first"
    }
}
```

In the above flow block, the `forward_to` accepts an array of `receivers`. In previous examples, a single receiver was used but a basic building block of Flow is multiple inputs and outputs. In the above example `prometheus.scrape.first` is sending to both `prometheus.relabel.filter` and `prometheus.relabel.not_cool`. 

## Multiple Inputs

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
        url = "http://cortex:9009/api/prom/push"
    }
}
```

In the above flow blocks the `prometheus.remote_write.prom` receives input from both the `prometheus.relabel.cool` and `prometheus.relabel.not_cool`. 


