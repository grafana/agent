# Tanka Configs

**STATUS**: Work in progress, use of these configs is not recommended for production.

This directory contains the Tanka configs that we use to deploy the Grafana
Agent. It is marked as `v2` and is incompatible previous versions of the library
located in other directories.

This library is currently a work in progress and backwards-incompatible changes
may occur. Once the library is considered complete, no further backwards
incompatible changes will be made.

## Capabilities

This library is significantly simplified over the `v0` and `v1` counterparts.
Since there are many ways to combine the various functionalities of the Grafana
Agent, the `v2` library aims to stay out of your way and provide optional composible
helpers that may be useful for some people.

Users of the library will pick a controller for their deployment. They are
expected to know what feature are compatible with which controller:

| Controller       | Metrics              | Logs      | Traces | Integrations |
| ---------------- | -------------------  | --------- | ------ | ------------ |
| DaemonSet        | If host filtering    | Yes       | Yes    | No           |
| Deployment       | Yes                  | No        | No     | Yes          |
| StatefulSet      | Yes                  | No        | No     | Yes          |

Creating an incompatible deployment will cause runtime issues when running the
Agent (for example, if configuring Logs with a StatefulSet, you will only get
logs from the node the pods are running on).

To get full coverage of features, you must create multiple deployments of the
library. You may wish to combine a StatefulSet for metrics and integrations, a
Deployment for Traces, and a DaemonSet for logs.

## API

## Generate Agent Deployment

- `new(name='grafana-agent', namespace='')`: Create a new Agent without a
   controller.
- `withDeploymentController(replicas=1)`: Attach a Deployment as the Agent
  controller. Number of replicas may optionally be given.
- `withDaemonSetController()`: Attach a DaemonSet as the Agent controller.
- `withStatefulSetController(replicas=1, volumeClaims=[])`: Attach a StatefulSet
  as the Agent controller. Number of replicas and a set of volume claim
  templates may be given.

## Generate Scraping Service Syncer

The Scraping Service Syncer is used to sync metrics instance configs against the
scraping service config management API.

- `newSyncer(name='grafana-agent-sycner', namespace='', config={})`

## General

- `withAgentConfig(config)`: Provide a custom Agent config.
- `withArgsMixin(config)`: Pass a map of additional flags to set.
- `withMetricsPort(port)`: Value for the `http-metrics` port (default 80)
- `withImagesMixin(images)`: Use custom images instead of the defaults.
- `withConfigHash(include=true)`: Whether to include a config hash annotation.
- `withPortsMixin(ports=[])`: Mixin ports from `k.core.v1.containerPort` against
   the container and service.
- `withVolumesMixin(volumes=[])`: Volume to attach to the pod.
- `withVolumeMountsMixin(mounts=[])`: Volume mounts to attach to the container.

## Helpers

- `newKubernetesMetrics(config={})`: Creates a set of metrics scrape_configs for
  collecting metrics from Kubernetes pods.
- `newKubernetesLogs(config={})`: Creates a set of logs scrape_configs for
  collecting logs from Kubernetes pods.
- `newKubernetesTraces(config={})`: Creates a set of traces scrape_configs for
  associating spans with metadata from discovered Kubernetes pods.
- `withLogVolumeMounts(config={})`: Adds volume mounts to the controller for collecting
  logs.
- `withLogPermissions(config={})`: Runs the container as privileged and as the root user
  so logs can be collected properly.
- `withService(config)`: Add a service for the deployment, statefulset, or daemonset.
  Note that this must be called after any ports are added via `withPortsMixin`.


