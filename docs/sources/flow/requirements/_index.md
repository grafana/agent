---
aliases:
- /docs/grafana-cloud/agent/flow/sizing
- /docs/grafana-cloud/send-data/agent/flow/sizing
canonical: https://grafana.com/docs/agent/latest/flow/sizing/
description: Sizing and resource requirements
menuTitle: Sizing
title: Sizing and resource requirements
weight: 100
---

# Sizing and resource requirements

Capacity planning for {{< param "PRODUCT_NAME" >}} is a very important step in the deployment process. The sizing and capacity requirements for {{< param "PRODUCT_NAME" >}} depends on your specific environment, and the workload you have. Understanding the minimum hardware requirements for Grafana Agent Flow XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx

## Baseline

The following values define the baseline for estimating the resource requirements for {{< param "PRODUCT_NAME" >}}. These values provide a starting point for your deployment requirements.

| Data point | Size    | Throughput |
|------------|---------|------------|
| Metrics    | XXXXXXX | XXXXXXX    |
| Events     | XXXXXXX | XXXXXXX    |
| Logs       | XXXXXXX | XXXXXXX    |
| Traces     | XXXXXXX | XXXXXXX    |

A metrics series is active if it continuously has new data points appended to it. A metric series become inactive if they haven't received any new samples since the last WAL truncation interval.

## Test environment

Info about Agent version, software/OS, hardware etc. Definition of what a vCPU is?

Does it make any difference running on Linux, Windows, OSX, etc.?

### Test results for 14,000,000 to 16,000,000 metric series

| vCPU cores | Memory         | Disk I/O                                        | Disk space       | Network |
|------------|----------------|-------------------------------------------------|------------------| ------- |
| 1.5 to 2.2 | 81 GB to 88 GB | 0.7%, with periodic spikes to 6%-8% for the WAL | 2.7 GB to 7.4 GB | XXXXXX  |

### Test results for 650,000 to 800,000 metric series

| vCPU cores | Memory           | Disk I/O                                       | Disk space       | Network |
|------------|------------------|------------------------------------------------|------------------| ------- |
| 0.1 to 0.2 | 3.9 GB to 5.7 GB | 0.3%, with periodic spikes to 0.4% for the WAL | 230 MB to 410 MB | XXXXXX  |

## Recommendations

Intro covering general recommendations.

### CPUs

Info about vCPUs for common use cases. For example X vCPUs when scraping some measurable number of metrics and sending over remote_write.

### Memory

Info about memory/RAM for common use cases. For example X GB RAM when scraping some measurable number of metrics and sending over remote_write.

### Disk

Info about disk space for common use cases. For example X GB of disk space when scraping some measurable number of metrics and sending over remote_write.

### Network

Info about network bandwidth for common use cases. For example X MBps when scraping some measurable number of metrics and sending over remote_write.

## Clustering requirements

Text
