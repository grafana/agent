---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.processor.k8sattributes/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.processor.k8sattributes/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.processor.k8sattributes/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.processor.k8sattributes/
title: otelcol.processor.k8sattributes
---

# otelcol.processor.k8sattributes

`otelcol.processor.k8sattributes` accepts telemetry data from other `otelcol`
components and adds Kubernetes metadata to the resource attributes of spans, logs, or metrics.

> **NOTE**: `otelcol.processor.k8sattributes` is a wrapper over the upstream
> OpenTelemetry Collector `k8sattributes` processor. Bug reports or feature requests
> will be redirected to the upstream repository, if necessary.

You can specify multiple `otelcol.processor.k8sattributes` components by giving them
different labels.

## Usage

```river
otelcol.processor.k8sattributes "LABEL" {
  output {
    metrics = [...]
    logs    = [...]
    traces  = [...]
  }
}
```

## Arguments

The following arguments are supported:

Name | Type | Description                                | Default         | Required
---- | ---- |--------------------------------------------|-----------------| --------
`auth_type` | `string` | Authentication method when connecting to the Kubernetes API. | `serviceAccount` | no
`passthrough` | `bool` | Passthrough signals as-is, only adding a `k8s.pod.ip` resource attribute. | `false` | no

The supported values for `auth_type` are:
* `none`: No authentication is required.
* `serviceAccount`: Use the built-in service account which Kubernetes automatically provisions for each pod.
* `kubeConfig`: Use local credentials like those used by kubectl.
* `tls`: Use client TLS authentication.

Setting `passthrough` to `true` enables the "passthrough mode" of `otelcol.processor.k8sattributes`:
* Only a `k8s.pod.ip` resource attribute will be added.
* No other metadata will be added.
* The Kubernetes API will not be accessed.
* The Agent must receive spans directly from services to be able to correctly detect the pod IPs.

The `passthrough` setting is useful when configuring the Agent as a Kubernetes Deployment.
An Agent running as a Deployment cannot detect the IP addresses of pods generating telemetry 
data without any of the well-known IP attributes. If the Deployment Agent receives telemetry from 
Agents deployed as DaemonSet, then some of those attributes might be missing. As a workaround 
to this issue, the DaemonSet Agents can be configured with `passthrough` set to `true`.
The supported values for `auth_type` are:
* `none`: No authentication.
* `serviceAccount`: Use the service account token mounted inside the pod.
* `kubeConfig`: Use the Kubernetes config file mounted inside the pod.
* `tls`: Use client TLS authentication.

Setting `passthrough` to `true` enables the "passthrough mode" of `otelcol.processor.k8sattributes`:
* Only a `k8s.pod.ip` resource attribute will be added.
* No other metadata will be added.
* The Kubernetes API will not be accessed.
* The Agent must receive spans directly from services to be able to correctly detect the pod IPs.
The `passthrough` setting is useful when configuring the Agent as a Kubernetes Deployment.
An Agent running as a Deployment cannot detect the IP addresses of pods generating telemetry 
data without any of the well-known IP attributes. If the Deployment Agent receives telemetry from 
Agents deployed as DaemonSet, then some of those attributes might be missing. As a workaround 
to this issue, the DaemonSet Agents can be configured with `passthrough` set to `true`.

## Blocks

The following blocks are supported inside the definition of
`otelcol.processor.k8sattributes`:
Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
output | [output][] | Configures where to send received telemetry data. | yes
extract | [extract][] | Rules for extracting data from Kubernetes. | no
extract > annotation | [extract_field][] | Creating resource attributes from Kubernetes annotations. | no
extract > label | [extract_field][] | Creating resource attributes from Kubernetes labels. | no
filter | [filter][] | Filters the data loaded from Kubernetes. | no
filter > field | [filter_field][] | Filter pods by generic Kubernetes fields. | no
filter > label | [filter_field][] | Filter pods by generic Kubernetes labels. | no
pod_association | [pod_association][] | Rules to associate pod metadata with telemetry signals. | no
pod_association > source | [pod_association_source][] | Source information to identify a pod. | no
exclude | [exclude][] | Exclude pods from being processed. | no
exclude > pod | [pod][] | Pod information. | no


The `>` symbol indicates deeper levels of nesting. For example, `extract > annotation`
refers to an `annotation` block defined inside an `extract` block.

[output]: #output-block
[extract]: #extract-block
[extract_field]: #extract-field-block
[filter]: #filter-block
[filter_field]: #filter-field-block
[pod_association]: #pod-association-block
[pod_association_source]: #pod-association-source-block
[exclude]: #exclude-block
[pod]: #pod-block

### Extract block

The `extract` block configures which metadata, annotations and labels to extract from the pod, and to add to the spans.

The following attributes are supported:

Name | Type           | Description                                                                                                                                                                                                                                                 | Default | Required
---- |----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------| ------- | --------
`metadata` | `list(string)` | Pre-configured metadata keys to add. See [k8sattributeprocessor](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/processor/k8sattributesprocessor/v0.85.0/processor/k8sattributesprocessor/config.go#L119) for the list of keywords. |  | no


### Extract field block

The `label` or `annotation` block configures which metadata or labels to extract from the pod and add to the spans.

The following attributes are supported:

Name | Type           | Description                                                                                           | Default | Required
---- |----------------|-------------------------------------------------------------------------------------------------------|---------| --------
`tag_name` | `list(string)` | TagName represents the name of the resource attribute that will be added to logs, metrics or spans.   |         | no
`key` | `list(string)` | Key represents the annotation (or label) name. This must exactly match an annotation (or label) name. |         | no
`key_regex` | `list(string)` | KeyRegex is a regular expression used to extract a Key that matches the regex.                   |         | no
`regex` | `list(string)` | Regex is an optional field used to extract a sub-string from a complex field value.                                                                                                  |         | no
`from` | `list(string)` | From represents the source of the labels/annotations. Allowed values are "pod" and "namespace".                                                | `pod`    | no


### filter block

The `filter` block configures which nodes to get data from, and which fields and labels to fetch.

The following attributes are supported:

Name | Type     | Description                                                             | Default | Required
---- |----------|-------------------------------------------------------------------------| ------- | --------
`node` | `string` | Configures a Kubernetes node name or host name. | `""` | no
`namespace` | `string` | Filters all pods by the provided namespace. All other pods are ignored. | `""` | no

If `node` is specified, then any pods not running on the specified node will be ignored by `otelcol.processor.k8sattributes`.

### Filter field block

The `field` block allows to filter pods by generic k8s fields.

The following attributes are supported:

Name | Type     | Description                                                                       | Default | Required
---- |----------|-----------------------------------------------------------------------------------|---------| --------
`key` | `string` | Key represents the key or name of the field or labels that a filter can apply on. |         | yes
`value` | `string` | Value represents the value associated with the key that a filter can apply on.    |         | yes
`op` | `string` | Op represents the filter operation to apply on the given Key: Value pair.         | `equals` | no

For `op` the following values are allowed:
* `equals`: The field value must be equal to the provided value.
* `not-equals`: The field value must not be equal to the provided value.
* `exists`: The field value must exist. (Only for `annotation` fields).
* `does-not-exist`: The field value must not exist. (Only for `annotation` fields).

### pod_association block

The `pod_association` block configures rules on how to associate logs/traces/metrics to pods.

The `pod_association` block does not support any arguments, and is configured
fully through child blocks.

### Pod association source block

The `source` block configures 

The following attributes are supported:

Name | Type     | Description                                                                      | Default | Required
---- |----------|----------------------------------------------------------------------------------| ------- | --------
`from` | `string` | The association method. Currently supports `resource_attribute` and `connection` |  | yes
`name` | `string` | Name represents extracted key name. e.g. `ip`, `pod_uid`, `k8s.pod.ip`           |  | no


### Exclude block

The `exclude` block configures which pods to exclude from the processor.

### Pod block

The `pod` block configures a pod to be excluded from the processor.

The following attributes are supported:

Name | Type     | Description         | Default | Required
---- |----------|---------------------| ------- | --------
`name` | `string` | The name of the pod |  | yes

### Output block

{{< docs/shared lookup="flow/reference/components/output-block.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics, logs, or traces).

## Component health

`otelcol.processor.k8sattributes` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.processor.k8sattributes` does not expose any component-specific debug
information.

## Examples

### Basic usage
In most cases this is enough to get started. It'll add these attributes to all telemetry data:
* `k8s.namespace.name`
* `k8s.pod.name`
* `k8s.pod.uid`
* `k8s.pod.start_time`
* `k8s.deployment.name`
* `k8s.node.name`

Example:
```river
otelcol.receiver.otlp "default" {
  http {}
  grpc {}

  output {
    metrics = [otelcol.processor.k8sattributes.default.input]
    logs    = [otelcol.processor.k8sattributes.default.input]
    traces  = [otelcol.processor.k8sattributes.default.input]
  }
}

otelcol.processor.k8sattributes "default" {
    output {
        metrics = [otelcol.exporter.otlp.default.input]
        logs    = [otelcol.exporter.otlp.default.input]
        traces  = [otelcol.exporter.otlp.default.input]
    }
}
```

### Add additional metadata and labels to spans

```river
otelcol.receiver.otlp "default" {
  http {}
  grpc {}

  output {
    metrics = [otelcol.processor.k8sattributes.default.input]
    logs    = [otelcol.processor.k8sattributes.default.input]
    traces  = [otelcol.processor.k8sattributes.default.input]
  }
}

otelcol.processor.k8sattributes "default" {
    extract {
		label {
			from      = "pod"
			key_regex = "(.*)/(.*)"
			tag_name  = "$1.$2"
		}

		metadata = [
			"k8s.namespace.name",
			"k8s.deployment.name",
			"k8s.statefulset.name",
			"k8s.daemonset.name",
			"k8s.cronjob.name",
			"k8s.job.name",
			"k8s.node.name",
			"k8s.pod.name",
			"k8s.pod.uid",
			"k8s.pod.start_time",
		]
	}

    output {
        metrics = [otelcol.exporter.otlp.default.input]
        logs    = [otelcol.exporter.otlp.default.input]
        traces  = [otelcol.exporter.otlp.default.input]
    }
}

otelcol.exporter.otlp "default" {
  client {
    endpoint = env("OTLP_ENDPOINT")
  }
}
```
