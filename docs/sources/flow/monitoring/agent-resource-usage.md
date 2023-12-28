---
aliases:
- /docs/agent/flow/monitoring/resource-usage/
- /docs/grafana-cloud/agent/flow/monitoring/resource-usage/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/monitoring/resource-usage/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/monitoring/resource-usage/
- /docs/grafana-cloud/send-data/agent/flow/monitoring/resource-usage/
canonical: https://grafana.com/docs/agent/latest/flow/monitoring/resource-usage/
description: Guidance for expected Agent resource usage
headless: true
title: Resource usage
---

# {{% param "PRODUCT_NAME" %}} resource usage
This page provides guidance for expected resource usage of {{% param "PRODUCT_NAME" %}}
for each telemetry type. The information on this page is based on operational
experience of some of the {{% param "PRODUCT_NAME" %}} maintainers.

{{% admonition type="note" %}}

The resource usage depends on the workload, hardware and the configuration used.
The information provided on this page is intended to be a starting point for 
most users and your actual resource usage may be different.

{{% /admonition %}}

## Prometheus metrics

The Prometheus metrics resource usage depends mainly on the number of active series
that need to be scraped.

As a rule of thumb, **per each 1 million active series** with 15s scrape interval,
you can expect to use approximately:
* 1.1 CPU cores 
* 10 GiB of memory
* 4.5 MiB/s of total network bandwidth, send and receive

The recommendations above are based on deployments
that use [clustering][], but they will broadly apply to other deployment modes.
For more information on how to deploy {{% param "PRODUCT_NAME" %}}, see
[deploying grafana agent][].

[deploying grafana agent]: {{< relref "../setup/deploy-agent.md" >}}
[clustering]: {{< relref "../concepts/clustering.md" >}}

## Loki logs

Loki logs resource usage depends mainly on the volume of logs ingested.

As a rule of thumb, **per each 1 MiB/second of logs ingested**, you can expect to use approximately:
* 1 CPU core
* 120 MiB of memory

The recommendations above are based on Kubernetes DaemonSet deployments on clusters
with relatively small number of large nodes and high logs volume. The resource usage
can be higher per each 1 MiB/second of logs if you have a large number of small nodes due
to the overhead of running the {{% param "PRODUCT_NAME" %}} on each node.

Additionally, factors such as number of labels, number of files and average log line length
may all play a role in the resource usage.


## Pyroscope profiles

Pyroscope profiles resource usage depends mainly on the volume of profiles.

As a rule of thumb, **per each 100 profiles/second**, you can expect to use approximately:
* 1 CPU core
* 10 GiB of memory

Factors such as size of each profile and frequency of fetching them also play
a role in the overall resource usage.
