---
aliases:
- ./kubernetes-logs/
- /docs/grafana-cloud/agent/flow/tutorials/kubernetes-logs/
canonical: https://grafana.com/docs/agent/latest/flow/tutorials/kubernetes-logs/
description: Learn how to collect Kubernetes Pod Logs
menuTitle: Collect Kubernetes Logs
title: Collecting Kubernetes Logs
weight: 400
---

## Discovering and Scraping Logs

### Option 1: Daemonset to scrape log files on each node

discovery.kubernetes (with node filter) -> discovery.relabel -> local.file -> loki.source.file

- explain discovery.kubelet as alternative (with tested example)

### Option 2: Clustered agents to scrape logs via k8a api

- only recommended in Agent v0.38+

- discovery.kubernetes (without filter) -> loki.source.kubernetes (clustered!)

- podlogs as alternative (or alongside)

## Writing logs to loki

- loki.write

## Processing

- Simple static example (just docker or cri)

- Some application specific processing with filters

- Label based process selection