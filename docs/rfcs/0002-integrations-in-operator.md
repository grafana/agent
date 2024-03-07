# Integrations in Grafana Agent Operator

* Date: 2022-01-04
* Author: Robert Fratto (@rfratto)
* PR: [grafana/agent#1224](https://github.com/grafana/agent/pull/1224)

## Background

Grafana Agent includes support for integrations, which are intended as
"batteries-included" features to assist with collecting telemetry data. With
the `integrations-next` feature enabled, there are multiple types of
integrations:

* Integrations that generate metrics (i.e., `node_exporter`)
* Integrations that generate logs (i.e., `eventhandler`)
* Integrations that generate other types of telemetry are planned (i.e., an
  upcoming `app_agent_receiver`)

Generically, an integration is a specialized telemetry collector for some
system under observation. For example, a `redis` integration collects telemetry
for Redis. Integrations can generate any combination of Prometheus metrics,
Grafana Loki logs, or Grafana Tempo traces.

This document proposes adding a way to add support for all current and future
integrations into the Grafana Agent Operator.

This proposal supersedes [#883][], which was the first attempt at designing the
feature. This proposal takes advantage of the lessons I've learned and
minimizes the implementation effort.

## Goals

* Allow Grafana Agent Operator to deploy integrations
* Allow deployed integrations to write telemetry data
* Support integrations that must exist on every machine (i.e., `node_exporter`)
* Minimize development effort for creating new integrations

## Non-Goals

* Support externally collecting metrics from integrations

## Proposal

At a high level, the proposal is to:

* Define a new `Integration` CRD which specifies a single instance of an
  integration to run.
* Update `GrafanaAgent` to discover `Integration`s and run integrations.

## Architecture

### Running integrations

The new CRD, Integration, will be used for supporting all current integrations.
The spec of Integration primarily revolves around three fields:

* `name`: The name of the integration (e.g., `node_exporter`, `mysqld_exporter`)
* `type`: Information about the integration being deployed
* `config`: YAML configuration block for the integration

The `type` field is an object with the following fields:

* `allNodes`: True when the `name` integration should run on all Kubernetes
  Nodes.
* `unique`: True when the `name` integration must be unique across a
  GrafanaAgent resource hierarchy.

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
hierarchy. During reconcile, the following Kubernetes objects will be deployed:

* One DaemonSet and Secret if there is at least one integration in the resource
  hierarchy where `type.allNodes` is true.

* One Deployment and Secret if there is at least one integration in the
  resource hierarchy where `type.allNodes` is false.

Secrets hold the generated Grafana Agent configuration; a Secret is used as
integration configs may contain credentials.

**NOTE**: As this functionality depends on [#1198][], integration pods will
always be deployed with the experimental feature flag
`-enable-feature=integrations-next` enabled. This also means that operator
support for integrations requires a release of the agent where that
experimental feature is available.

### Integration validation

The initial implementation of integrations support will have no knowledge of
what integrations exist. As a result, the `spec.type` and `spec.config` fields
for an Integration MUST be configured correctly for an integration to work.
Users must refer to documentation to discover how `type` should be configured
for their specific integration, and what settings are valid for the `config`
block. Configuration errors will only surface as runtime errors from the
deployed agent.

Future versions of the Operator may:

* Add knowledge for some integrations and validate `type` and `config`
  accordingly (though breaking changes to the config at the Agent level may
  introduce extra complexity to this).

* Update the `status` field of the root GrafanaAgent resource during reconcile
  to expose any reconcile or runtime errors.

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
>   config: |
>     ca_file: /etc/grafana-agent/secrets/kafka-ca-file
>     # ...
>   # Same "secrets" field present in GrafanaAgent.spec, where each secret
>   # is loaded from the same namespace and gets exposed at
>   # /etc/grafana-agent/secrets/<secret name>
>   secrets: [kafka-ca-file]
> ```

### Sending telemetry from integrations

Because the operator will not have any knowledge about individual integrations, it
also doesn't know how integrations generate telemetry data. Users must manually
configure an integration to send its data to the appropriate instance.

Users can refer to MetricsInstances and LogsInstance from the same resource
hierarchy by `<namespace>/<name>` in their integration configs. This
configuring `autoscrape` for collecting metrics from an exporter-based
integration.

Given the following resource hierarchy:

> ```yaml
> apiVersion: monitoring.grafana.com/v1alpha1
> kind: GrafanaAgent
> metadata:
>   name: grafana-agent-example
>   namespace: default
>   labels:
>     app: grafana-agent-example
> spec:
>   metrics:
>     instanceSelector:
>       matchLabels:
>         agent: grafana-agent-example
>   integrations:
>     instanceSelector:
>       matchLabels:
>         agent: grafana-agent-example
> ---
> apiVersion: monitoring.grafana.com/v1alpha1
> kind: MetricsInstance
> metadata:
>   name: primary
>   namespace: default
>   labels:
>     app: grafana-agent-example
> spec:
>   remoteWrite:
>   - url: http://prometheus:9090/api/v1/write
> ---
> apiVersion: monitoring.grafana.com/v1alpha1
> kind: Integration
> metadata:
>   name: mysql
>   namespace: default
>   labels:
>     app: grafana-agent-example
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

the Operator would generate the following agent config:

```yaml
metrics:
  configs:
  - name: default/primary
    remote_write:
    - url: http://prometheus:9090/api/v1/write
integrations:
  mysqld_exporter_configs:
  - autoscrape:
      enable: true
      metrics_instance: default/primary
    data_source_name: root@(mysql.default:3306)/
    disable_collectors: [slave_status]
```

All integrations support some way of self-collecting their telemetry data. In
the future, Integrations that support metrics could support being collected by
an external source (i.e., a MetricsInstance). This is out of scope of this
proposal, as we are focusing on lowest-common-denominator support for all
integrations first.

Note that the Integration config above is only contextually valid: it is only
valid if it is part of a resource hierarchy where a `default/primary`
MetricsInstance exists. This makes it impossible for an Integration to be fully
validated independently of the resource hierarchy where it is discovered.

## Pros/Cons

Despite its limitations, this specific implementation is proposed for its
simplicity. Its issues with validation can be resolved in the future without
needing to change the CRD or introduce new CRDs.

Pros:

* Works for all known integrations
* Supports future work for custom validation logic
* No changes needed to support future integrations
* You do not have to update the operator to use new integrations

Cons:

* Users must know use documentation to configure `type` and `config` properly.
* Without validation, configuration errors can be hard to debug.
* An Integration may be discovered as part of two resource hierarchies, but
  refer to a MetricsInstance that exists in one hierarchy but not the other.

## Alternatives considered

### Do nothing

Instead of adding support for integrations, users could be expected to deploy
exporters though custom means (i.e., a `node_exporter` Helm chart +
ServiceMonitor).

Pros:

* Requires no additional effort to implement
* Metrics can be scraped by any MetricsInstance
* Feels like a natural fit for Kubernetes' deployment model

Cons:

* Prevents non-exporter integrations from working (i.e., `eventhandler` has no
  separate container that can be run independently)
* Prevents us from making agent-specific changes on top of exporters
* Requires different documentation for people using the node_exporter
  integration vs deploying the actual node_exporter

### One CRD per integration

Instead of a generic CRD, we could have CRD per supported integration.

Pros:

* Allows creating Kubernetes-specific config schemas for integrations
* Can be validated at the CRD level

Cons:

* Operator must be updated whenever a new integration is added
* Adds extra development effort for creating new integrations
* Requires custom config mapping code for each integration
* Breaking changes to Grafana Agent can break the translation of the CRD to
  Agent config.
  * This is true for the current proposal, but in the current proposal you can
    fix the error in the Integration resource, while a custom CRD would need a
    new operator version to fix the translation.

[#883]: https://github.com/grafana/agent/issues/883
[#1198]: https://github.com/grafana/agent/pull/1198
