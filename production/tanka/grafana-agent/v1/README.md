# Tanka Configs

**STATUS**: Work in progress, use of these configs is not recommended for production.

This directory contains the Tanka configs that we use to deploy the Grafana
Cloud Agent. It is marked as `v1` and is incompatible with the `v0` configs
found in the [parent directory](../).

This library is currently a work in progress and backwards-incompatible changes
may occur. Once the library is considered complete, no further backwards
incompatible changes will be made.

## Capabilities

This library is significantly more flexible than its `v0` counterpart. It tries
to allow to deploy and configure the Agent in a feature matrix:

| Mechanism        | Prometheus Metrics | Loki Logs | Traces | Integrations |
| ---------------- | ------------------ | --------- | ------ | ------------ |
| DaemonSet        | Yes                | Yes       | Yes    | Yes          |
| Deployment       | Yes                | No        | No     | No           |
| Scraping Service | Yes                | No        | No     | No           |

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

## Configure Prometheus

- `withPrometheusConfig(config)`: Creates a Prometheus config block.
- `defaultPrometheusConfig`: Default Prometheus config block.
- `withPrometheusInstances(instances)`: Creates a Prometheus instance config to
  tell the Agent what to scrape.
- `withRemoteWrite(remote_writes)`: Configures locations to remote write metrics
   to. Controlls remote writes for all instances.
- `scrapeInstanceKubernetes`: Default Prometheus instance config to scrape from
  Kubernetes.

## Configure Loki

- `withLokiConfig(config)`: Creates a Loki config block to pass to the Agent.
- `newLokiClient(client_config)`: Creates a new client configuration to pass
  to `withLokiClients`.
- `withLokiClients(clients)`: Add a set of clients to a Loki config block.
- `scrapeKubernetesLogs`: Default Loki config that collects logs from Kubernetes
  pods.

## Configure Tempo

- `withTempoConfig(config)`: Creates a Tempo config block to pass to the Agent.
- `withTempoPushConfig(push_config)`: Configures a location to push spans to.
- `withTempoSamplingStrategies(strategies)`: Configures strategies for trace collection.
- `withTempoScrapeConfigs(scrape_configs)`: Configures scrape configs to attach
   labels to incoming spans.
- `tempoScrapeKubernetes`: Default scrape configs to collect meta information
   from pods. Aligns with the labels from `scrapeInstanceKubernetes` and
   `scrapeKubernetesLogs` so logs, metrics, and traces all use the same set of
   labels.

## General

- `withImages(images)`: Use custom images.
- `withConfigHash(include=true)`: Whether to include a config hash annotation.
- `withPortsMixin(ports)`: Mixin ports from `k.core.v1.containerPort` against
   the container and service.
