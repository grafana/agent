---
aliases:
- /docs/agent/latest/flow/tutorials/filtering-metrics
title: Filtering metrics
weight: 300
---

The goal of this tutorial is to add a new component [prometheus.relabel]({{< ref "prometheus.relabel.md" >}}) using [relabel.flow](../assets/flow_configs/relabel.flow) that allows the filtering of metrics. This builds upon [Collecting Prometheus Metrics]({{< relref "./collecting-prometheus-metrics.md">}}) and requires all the other previous prerequisites.

# Prometheus Relabel Component

The `prometheus.relabel` component is used to drop, add or filter metrics.  To quickly spin up an example environment run the following: `docker-compose -f ./assets/adding-new-component.yaml up` from the tutorials directory.

Wait two minutes to ensure the scrape has been completed then visit the [Grafana](http://localhost:3000/explore?orgId=1&left=%5B%22now-1h%22,%22now%22,%22Cortex%22,%7B%22refId%22:%22A%22,%22instant%22:true,%22range%22:true,%22exemplar%22:false,%22expr%22:%22rate(process_cpu_seconds_total%5B5m%5D)%22%7D%5D) page and the `cool_label` will be there.

![](../assets/filter.png)

# What's Happening

The scraper is sending the metrics to the filter which is then sending metrics to the remote_write. 