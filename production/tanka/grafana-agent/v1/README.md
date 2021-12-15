# Tanka Configs

**STATUS**: Abandoned. Use v0 (parent directory) or v2 instead.

This directory contains the Tanka configs that we use to deploy the Grafana
Agent. It is marked as `v1` and is incompatible with the `v0` configs
found in the [parent directory](../).

This library is currently a work in progress and backwards-incompatible changes
may occur. Once the library is considered complete, no further backwards
incompatible changes will be made.

## Capabilities

This library is significantly more flexible than its `v0` counterpart. It tries
to allow to deploy and configure the Agent in a feature matrix:

| Mechanism        | Metrics | Logs      | Traces | Integrations |
| ---------------- | ------- | --------- | ------ | ------------ |
| DaemonSet        | Yes     | Yes       | Yes    | Yes          |
| Deployment       | Yes     | No        | No     | No           |
| Scraping Service | Yes     | No        | No     | No           |

The library can be invoked multiple times to get full coverage. For example, you
may wish to deploy a scraping service for scalable metrics collection, and a
DaemonSet with just Loki Logs for log collection.

Trying to use the library in incompatible ways will generate errors. For
example, you may not deploy a scraping service with Loki logs collection.

## API

## Generate Agent Deployment

- `new(name, namespace)`: Create a new DaemonSet. This is the default mode to
  deploy the Agent.  Enables host filtering.
- `newDeployment(name, namespace)`: Create a new single-replica Deployment.
  Disables host filtering.
- `newScrapingService(name, namespace, replicas)`: (Not yet available). Create a
  scalable deployment of clustered Agents. Requires being given a KV store such as Redis or ETCD.

## Configure Metrics

- `withMetricsConfig(config)`: Creates a metrics config block.
- `defaultMetricsConfig`: Default metrics config block.
- `withMetricsInstances(instances)`: Creates a metrics instance config to
  tell the Agent what to scrape.
- `withRemoteWrite(remote_writes)`: Configures locations to remote write metrics
   to. Controls remote writes globally.
- `scrapeInstanceKubernetes`: Default metrics instance config to scrape from
  Kubernetes.

## Configure Logs

- `withLogsConfig(config)`: Creates a Logs config block to pass to the Agent.
- `newLogsClient(client_config)`: Creates a new client configuration to pass
  to `withLogsClients`.
- `withLogsClients(clients)`: Add a set of clients to a Logs config block.
- `scrapeKubernetesLogs`: Default Logs config that collects logs from Kubernetes
  pods.

## Configure Traces

- `withTracesConfig(config)`: Creates a Traces config block to pass to the Agent.
- `withTracesRemoteWrite(remote_write)`: Configures one or multiple locations to push spans to.
- `withTracesSamplingStrategies(strategies)`: Configures strategies for trace collection.
- `withTracesScrapeConfigs(scrape_configs)`: Configures scrape configs to attach
   labels to incoming spans.
- `tracesScrapeKubernetes`: Default scrape configs to collect meta information
   from pods. Aligns with the labels from `scrapeInstanceKubernetes` and
   `scrapeKubernetesLogs` so logs, metrics, and traces all use the same set of
   labels.

## General

- `withImages(images)`: Use custom images.
- `withConfigHash(include=true)`: Whether to include a config hash annotation.
- `withPortsMixin(ports)`: Mixin ports from `k.core.v1.containerPort` against
   the container and service.
