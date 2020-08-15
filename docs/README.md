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
    1. [Metrics](./overview.md#metrics)
    2. [Logs](./overview.md#logs)
    3. [Comparison to alternatives](./overview.md#comparison-to-alternatives)
    4. [Next Steps](./overview.md#next-steps)
2. [Getting Started](./getting-started.md)
    1. [Docker-Compose Example](./getting-started.md#docker-compose-example)
    2. [k3d Example](./getting-started.md#k3d-example)
    3. [Installing](./getting-started.md#installing)
    4. [Creating a Config File](./getting-started.md#creating-a-config-file)
        1. [Integrations](./getting-started.md#integrations)
        2. [Prometheus-like Config/Migrating from Prometheus](./getting-started.md#prometheus-like-configmigrating-from-prometheus)
        3. [Loki Config/Migrating from Promtail](./getting-started.md#loki-configmigrating-from-promtail)
    5. [Running](./getting-started.md#running)
3. [Configuration Reference](./configuration-reference.md)
4. [API](./api.md)
5. [Scraping Service Mode](./scraping-service.md)
6. [Operation Guide](./operation-guide.md)
7. [Maintainers Guide](./maintaining.md)
