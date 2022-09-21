---
aliases:
- /docs/agent/latest/flow/tutorials/filtering-metrics
title: Filtering Prometheus metrics
weight: 300
---

In this tutorial, you'll add a new component [prometheus.relabel]({{< ref "prometheus.relabel.md" >}}) using [relabel.flow](../assets/flow_configs/relabel.flow) to filter metrics. This tutorial uses the same base as [Collecting Prometheus metrics]({{< relref "./collecting-prometheus-metrics.md">}}).

# Prerequisites

* [Docker](https://www.docker.com/products/docker-desktop)
* Clone the [Agent Repository](https://github.com/grafana/agent) `git clone git@github.com:grafana/agent.git`

# Prometheus Relabel component

The `prometheus.relabel` component is used to drop, add, or filter metrics.  To quickly spin up an example environment, run the following: `CONFIG_FILE=relabel.flow docker-compose -f ./assets/docker-compose.yaml up` from the `tutorials` directory.

Wait two minutes to ensure the scrape is complete, then visit the [Grafana](http://localhost:3000/explore?orgId=1&left=%5B%22now-1h%22,%22now%22,%22Mimir%22,%7B%22refId%22:%22A%22,%22instant%22:true,%22range%22:true,%22exemplar%22:true,%22expr%22:%22agent_build_info%7B%7D%22%7D%5D) page and the `cool_label` will be there.

![](../assets/filter.png)

# What's happening?

The scraper is sending the metrics to the filter which is then sending metrics to the remote_write. 