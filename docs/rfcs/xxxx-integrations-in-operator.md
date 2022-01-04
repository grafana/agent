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
"batteries-included" features to assist with collecting telemetry data. Today,
all integrations are metrics-based, where each metrics-based integration is an
embeded Prometheus exporter. This document proposes adding support for
metrics-based integrations into the Grafana Agent Operator.

This proposal supersedes [#883][], which was the first attempt at designing the
feature. This proposal takes advantage of the lessons we've learned and
utilizes the upcoming [integrations revamp][#1198] work done in December 2021
so this functionality can be implemented more easily.

## Goals

* Allow Grafana Agent Operator to deploy integrations
* Allow Grafana Agent Operator to collect metrics from deployed integrations
* Support both per-node and "normal" integrations
* Use a generic CRD for all metrics-based integrations

The final goal prevents us from incurring any additional overhead when
developing new integrations. The operator will be able to configure and deploy
any integration without having to know what the underlying Grafana Agent
version supports.

## Non-Goals

* Minimize number of created pods used for metrics / logging / integrations
* Discussion of deployment strategies for multiple GrafanaAgents

## Proposal

At a high level, the proposal is to:

* Define two new CRDs:
  * `MetricsIntegration`: an integration to run
  * `IntegrationMonitor`: collection rules for integrations, similar to
    ServiceMonitor.
* Update `GrafanaAgent` to discover `MetricsIntegration`s and run integrations.
* Update `MetricsInstance` to discover `IntegrationMonitor`s and collect
  metrics from integrations.

## Architecture

### Running integrations

The new CRD, MetricsIntegration, will be used for supporting all metrics-based
integrations today. A metrics-based integration is an integration that exposes
metrics. These integrations have an always-available set of common
configuration fields. We define MetricsIntegration as a CRD specific for
metrics-based integrations to add guarantees that these common configuration
fields can be used when necessary for implementation details.

The spec of MetricsIntegration primarily revolves around three fields:

* `name`: The name of the integration (e.g., `node_exporter`, `mysqld_exporter`)
* `type`: The type of integration
* `config`: A YAML string declaring the configuration for the integration.

`type` is critical for determining how the integration should be deployed.
There are three supported values for `type`:

* `daemonset`: Declares that the `name` integration should be run on every Node in the
  Kubernetes cluster. It is invalid to have a GrafanaAgent CRD discover more
  than one `daemonset` integrations with the same `name` . Example integration:
  `node_exporter`.

* `singleton`: Declares that the `name` integration may only be run once per
  deployment. It is invalid to have a GrafanaAgent CRD discover more than one
  `singleton` integration with the same `name.` Example integration:
  `statsd_exporter`.

* `normal`: Declares that the integration may exist any number of times per
  deployment. This is the default. Example integration: `mysqld_exporter`.

Users must supply the appropriate `type` value for the integration they are
deploying. Integrations will be documented with their type to assist users in
defining `MetricsIntegration` resources. Failure to use the proper `type` value may lead
to failure to start Grafana Agent integration pods.

> Example of a valid integration:
>
> ```yaml
> apiVersion: monitoring.grafana.com/v1alpha1
> kind: MetricsIntegration
> metadata:
>   name: mysql
>   namespace: default
> spec:
>   name: mysqld_exporter
>   type: normal
>   config: |
>     data_source_name: root@(mysql.default:3306)/
>     disable_collectors: [slave_status]
> ```

GrafanaAgent will be updated to discover MetricsIntegrations as part of its
resource hierarchy, alongside MetricsInstances and LogsInstances. During
reconcile, the following Kubernetes objects will be deployed:

* One DaemonSet and Secret if there is at least one `daemonset` integration
  in the GrafanaAgent resource hierarchy.

* One Deployment and Secret if there is at least one non-`daemonset`
  integration in the GrafanaAgent resource hierarchy.

* One Service if there is at least one integration in the GrafanaAgent resource
  hierarchy of any kind.

Secrets are used for the Grafana Agent config as integration configs may
contain credentials. Deployed integrations will never self-scrape, and metrics
must be collected via an IntegrationMonitor.

The Service is used for exposing the endpoint for the individual pod's
integrations service discovery API introduced in [#1198][]. We will discuss
this in more detail later.

Some integrations may require changes to the deployed Pods to function
properly. MetricsIntegrations will additionally support declaring `volumes`,
`volumeMounts`, `secrets` and `configMaps`. These fields will be merged with
the fields of the same name from the root GrafanaAgent resource when creating
integration pods:

> ```yaml
> apiVersion: monitoring.grafana.com/v1alpha1
> kind: MetricsIntegration
> metadata:
>   name: kafka
>   namespace: default
> spec:
>   name: kafka_exporter
>   type: normal
>   config: |
>     ca_file: /etc/grafana-agent/secrets/kafka-ca-file
>     # ...
>   # Same "secrets" field present in GrafanaAgent.spec, where each secret
>   # is loaded from the same namespace and gets exposed at
>   # /etc/grafana-agent/secrets/<secret name>
>   secrets: [kafka-ca-file]
> ```

**NOTE**: As this functionality depends on [#1198][], integration pods will
always be deployed with the experimental feature flag
`-enable-feature=integrations-next` enabled. This also means that operator
support for integrations requires a release of the agent where that
experimental feature is available.

### Arbitrary integration labels

Metrics-based integrations will be extended with support for a `labels` field
to add arbitrary labels to those integrations.

Deployed integrations will always have the following labels set:

* `__meta_agentoperator_integration_type`: Integration type (`daemonset`, `singleton`, `normal`)
* `__meta_agentoperator_integration_cr_namespace`: Namespace of the MetricsIntegration CR
* `__meta_agentoperator_integration_cr_name`: Name of the MetricsIntegration CR
* `__meta_agentoperator_integration_cr_label_<labelname>`: Each label from the MetricsIntegration CR
* `__meta_agentoperator_integration_cr_labelpresent_<labelname>`: `true` for each label from the MetricsIntegration CR

### Discovering running integrations

[#1198][] introduces a new service discovery (SD) API to retrieve a list of
running integrations. The endpoint is available at
`/agent/api/v1/metrics/integrations/sd` and returns JSON which conforms to
Prometheus' [`http_sd_config`][http_sd_config] API.

Note that integration targets returned by this API always have the following
base labels:

* `instance`: The inferred `instance` key based on integration settings.
* `job`: `integrations/<integration name>`
* `agent_hostname`: Hostname of the agent running the integration. This will be
  the Node name for `daemonset` integrations, otherwise will be the Pod name.
* `__meta_agent_integration_name`: Name of the integration, e.g., `node_exporter`
* `__meta_agent_integration_instance`: The inferred `instance` key based on
  integration settings. Same as `instance`.
* `__meta_agent_integration_autoscrape`: `true` or `false` depending on whether
  autoscrape is enabled.

These labels are available on top of any custom labels supplied in the `labels`
field.

Pods created for MetricsInstances will use `http_sd_config` to find, filter,
and scrape metrics for integrations. However, there is a problem:
`http_sd_config` can only be configured with one known URL, but there may be
any number of pods which are running integrations.

The operator's existing HTTP server will be extended with a new endpoint which
aggregates the integrations targets for all managed integrations pods. This
endpoint will be available at `/operator/api/v1/metrics/integrations/sd`. For
this endpoint to work, the Operator will create a new controller that watches
for changes to Endpoints related to the integrations-specific Services created
by the operator. The list of relevant pod IPs will then be maintained
internally.

When the operator's integration SD endpoint is invoked, it will parallelize
over all pods, invoking their individual SD API endpoints, and aggregating the
responses into a final set. Failed API requests to pods will be logged but not
fail the operator's request.

Finally, on top of the target's labels, two additional labels will be added:

* `__meta_agentoperator_grafanaagent_name`: Name of the owning GrafanaAgent.
* `__meta_agentoperator_grafanaagent_namespace`: Namespace of the owning GrafanaAgent.

As a result, there is a centralized API to discover metrics endpoints for all
running integrations, including information about the integration, the owning
MetricsIntegration CR, and the root GrafanaAgent resource.

### Scraping integrations

The other new CRD, IntegrationMonitor, will be used to scrape running
integrations. IntegrationMonitors are discovered by a MetricsInstance, and
result in the generation of a integrations-specific scrape job. It is
configured similarly to a `ServiceMonitor`, though without settings for
endpoint port and path.

IntegrationMonitor use `http_sd_config` and the SD endpoint described in the
previous section. The full set of returned integrations will be filtered down
to the set to scrape based on the definition of the IntegrationMonitor and the
metalabels of the discovered target.

After scrape, metrics from integrations will always have the following labels:

* `cluster`: `<GrafanaAgent key>`
* `job`: `integrations/<integration name>`
* `instance`: `<MetricsIntegration key>`
* `monitor`: `<IntegrationMonitor key>`
* `metrics_instance`: `<MetricsIntegration key>`

A key is the set of namespace and name for a resource separated by `/`, e.g.,
`kube_system/kubernetes`.

The combination of cluster, job, and instance should form a unique target. The
`monitor` and `metrics_instance` labels are used for debugging and integration
being scraped more than once within Grafana, and are not intended to be used
for visualization.

[#883]: https://github.com/grafana/agent/issues/883
[#1198]: https://github.com/grafana/agent/pull/1198
[http_sd_config]: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#http_sd_config
