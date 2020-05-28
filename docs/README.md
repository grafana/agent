# Grafana Cloud Agent

The Grafana Cloud Agent is an observability data collector optimized for sending
metrics and log data to [Grafana Cloud](https://grafana.com/products/cloud/).
Today, with its Prometheus metrics collection, it is designed to handle the
main problems faced by users of large deployments of Prometheus:

- Grafana Cloud Agent uses less memory on average than Prometheus – by doing less
  (only focusing on `remote_write`-related functionality).
- Grafana Cloud Agent allows for deploying multiple instances of the Agent in a
  cluster and only scraping metrics from targets that running at the same host.
  This allows distributing memory requirements across the cluster
  rather than pressurizing a single node.

## Table of Contents

1. [Overview](./overview.md)
    1. [Comparison to alternatives](./overview.md#comparison-to-alternatives)
    2. [Roadmap](./overview.md#roadmap)
2. [Getting Started](./getting-started.md)
    1. [Docker-Compose Example](./getting-started.md#docker-compose-example)
    1. [k3d Example](./getting-started.md#k3d-example)
    2. [Installing](./getting-started.md#installing)
    3. [Migrating from Prometheus](./getting-started.md#migrating-from-prometheus)
    4. [Running](./getting-started.md#running)
3. [Configuration Reference](./configuration-reference.md)
4. [API](./api.md)
5. [Scraping Service Mode](./scraping-service.md)
6. [Maintainers Guide](./maintaining.md)
