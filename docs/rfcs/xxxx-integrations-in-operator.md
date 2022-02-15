---
title: Integrations in Grafana Agent Operator
weight: xxxx
---

# Integrations in Grafana Agent Operator

* Date: 2022-01-04
* Author: Robert Fratto (@rfratto)

## Background

At the time of writing, Grafana Agent Operator can be used to collect metrics
and logs from a Kubernetes cluster by deploying and managing Grafana Agent
pods. However, logs and metrics are just a subset of the total functionality
available in Grafana Agent.

Grafana Agent includes support for integrations, which are intended as
"batteries-included" features to assist with collecting telemetry data. With
the `integrations-next` feature enabled, there are multiple types of integrations:

* Integrations that generate metrics (i.e., `node_exporter`)
* Integrations that generate logs (i.e., `eventhandler`)
* In the future, integrations that generate traces (i.e., a planned `app_o11y_receiver`)

This document proposes adding a way to add support for all current and future
integrations into the Grafana Agent Operator.

This proposal supersedes [#883][], which was the first attempt at designing the
feature. This proposal takes advantage of the lessons we've learned and
minimizes the implementation effort.

## Goals

* Allow Grafana Agent Operator to deploy integrations
* Allow deployed integrations to self-generate telemetry data
* Support integrations that depend on node-level information (i.e., `node_exporter`)
* Minimize development effort for developing new integrations by using a
  generic integrations CRD

## Non-Goals

* Allow deployed integrations to have telemetry collected externally
* Minimize number of created pods used for metrics / logging / integrations
* Discussion of deployment strategies for multiple GrafanaAgents

## Proposal

At a high level, the proposal is to:

* Define a new `Integration` CRD which specifies a single instance of an
  integration to run.
* Update `GrafanaAgent` to discover `Integration`s and run integrations.

## Architecture

### Running integrations

The new CRD, Integration, will be used for supporting all current integrations.
The spec of MetricsIntegration primarily revolves around three fields:

* `name`: The name of the integration (e.g., `node_exporter`, `mysqld_exporter`)
* `type`: Information about the integration beyind deployed
* `config`: YAML configuration block for the integration

The `type` field is an object with the following fields:

* `allNodes`: True when the `name` integration should run on all Kubernetes
  Nodes.
* `unique`: True when the `name` integration must be unique across a
  GrafanaAgent resource hierarchy.

The values of the `type` field are crticical for managing the integration
properly. Documentation will inform users what settings for `type` are needed
for the integration they're configuring. Incorrect values may lead to undefined
behaviors when running the integration.

Future versions of Grafana Agent Operator may include knowledge of the
appropriate `type` field values for known integrations, and only require the
user to specify values for integrations which are unknown to the operator.

> Example of a valid integration:
>
> ```yaml
> apiVersion: monitoring.grafana.com/v1alpha1
> kind: Integration
> metadata:
>   name: mysql
>   namespace: default
> spec:
>   name: mysqld_exporter
>   type:
>     allNodes: false # optional; false is default
>     unique:   false # optional; false is default
>   config:
>     data_source_name: root@(mysql.default:3306)/
>     disable_collectors: [slave_status]
> ```

GrafanaAgent will be updated to discover Integrations as part of its resource
hierarchy, alongside MetricsInstances and LogsInstances. During reconcile, the
following Kubernetes objects will be deployed:

* One DaemonSet and Secret if there is at least one integration in the resource
  hierarchy where `type.allNodes` is set to true.

* One Deployment and Secret if there is at least one integration in the
  resource hierarchy where `type.allNodes` is set to false.

Secrets are used for the Grafana Agent config as integration configs may
contain credentials.

**NOTE**: As this functionality depends on [#1198][], integration pods will
always be deployed with the experimental feature flag
`-enable-feature=integrations-next` enabled. This also means that operator
support for integrations requires a release of the agent where that
experimental feature is available.

### Self-collecting telemetry from integrations

The generated Grafana Agent config for integrations includes the definitions of
the MetricsInstances and LogsInstances that exist within the hierarchy. This
allows integrations to push telemetry data to the defined instances:

> ```yaml
> apiVersion: monitoring.grafana.com/v1alpha1
> kind: MetricsInstance
> metadata:
>   name: primary
>   namespace: default
> spec:
>   remoteWrite:
>   - url: http://prometheus:9090/api/v1/write
> ---
> apiVersion: monitoring.grafana.com/v1alpha1
> kind: Integration
> metadata:
>   name: mysql
>   namespace: default
> spec:
>   name: mysqld_exporter
>   config:
>     autoscrape:
>       enable: true
>       # MetricsInstance <namespace>/<name> to send metrics to
>       metrics_instance: default/primary
>     data_source_name: root@(mysql.default:3306)/
>     disable_collectors: [slave_status]
> ```

The `autoscrape` configuration is part of the unchecked `config` block and a
user must ensure that they specify a `metrics_instance` that actually exists
within the hierarchy, otherwise the integration will fail to send metrics
anywhere.

An officially supported way for telemetry from an integration to be collected
by an external pod is out of scope of this proposal.

### Additional settings for the Integration CRD

Some integrations may require changes to the deployed Pods to function
properly. Integrations will additionally support declaring `volumes`,
`volumeMounts`, `secrets` and `configMaps`. These fields will be merged with
the fields of the same name from the root GrafanaAgent resource when creating
integration pods:

> ```yaml
> apiVersion: monitoring.grafana.com/v1alpha1
> kind: Integration
> metadata:
>   name: kafka
>   namespace: default
> spec:
>   name: kafka_exporter
>   type:
>     hasMetrics: true
>   config: |
>     ca_file: /etc/grafana-agent/secrets/kafka-ca-file
>     # ...
>   # Same "secrets" field present in GrafanaAgent.spec, where each secret
>   # is loaded from the same namespace and gets exposed at
>   # /etc/grafana-agent/secrets/<secret name>
>   secrets: [kafka-ca-file]
> ```

## Alternatives considered

[#883]: https://github.com/grafana/agent/issues/883
[#1198]: https://github.com/grafana/agent/pull/1198
