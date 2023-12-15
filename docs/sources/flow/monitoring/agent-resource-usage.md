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

When deploying {{% param "PRODUCT_NAME" %}} as a host daemon or a container
sidecar to collect logs, the resource usage depends mainly on the number of 
agents running.
Loki logs resource usage depends mainly on the volume of logs ingested.

As a rule of thumb, **per each 1 million log lines**, you can expect to use approximately:



TODO:
- Metrics resources:
  - CPU - # of cores per 1m series
  - Memory - # of GB per 1m series
  - Disk - # of GB for 1 minute of 1m series / or max disk usage approx 
  - Network - # MB/s for 1m of series
- Logs resources:
  - CPU - # of cores per 1m log lines? or CPU that is sufficient on our busiest node?
  - Memory - # of GB per 1m log lines? or memory that is sufficient on our busiest node?
  - Disk - just positions file or WAL - depepnding on logs volume and delay
  - Network - # MB/s for 1m of log lines?
- Traces resources
  - CPU - # of cores per 1m traces?
  - Memory - # of GB per 1m traces?
  - Disk - # of GB for 1 minute of 1m traces / or max disk usage approx
  - Network - # MB/s for 1m of traces
- Profiles resources
  - CPU - # of cores per 1m profiles?
  - Memory - # of GB per 1m profiles?
  - Disk - # of GB for 1 minute of 1m profiles / or max disk usage approx
  - Network - # MB/s for 1m of profiles
- Link to deployment page
- Choosing the topology
  - When single node is okay
  - When we should consider a cluster
- It's a starting point, your mileage may vary