---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.processor.k8sattributes/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.processor.k8sattributes/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.processor.k8sattributes/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.processor.k8sattributes/
title: otelcol.processor.k8sattributes
description: Learn about otelcol.processor.k8sattributes
---

# otelcol.processor.k8sattributes

`otelcol.processor.k8sattributes` accepts telemetry data from other `otelcol`
components and adds Kubernetes metadata to the resource attributes of spans, logs, or metrics.

{{% admonition type="note" %}}
`otelcol.processor.k8sattributes` is a wrapper over the upstream OpenTelemetry
Collector `k8sattributes` processor.  If necessary, bug reports or feature requests
will be redirected to the upstream repository.
{{% /admonition %}}

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
* `serviceAccount`: Use the built-in service account that Kubernetes automatically provisions for each pod.
* `kubeConfig`: Use local credentials like those used by kubectl.
* `tls`: Use client TLS authentication.

Setting `passthrough` to `true` enables the "passthrough mode" of `otelcol.processor.k8sattributes`:
* Only a `k8s.pod.ip` resource attribute will be added.
* No other metadata will be added.
* The Kubernetes API will not be accessed.
* To correctly detect the pod IPs, the Agent must receive spans directly from services.
* The `passthrough` setting is useful when configuring the Agent as a Kubernetes Deployment.
An Agent running as a Deployment cannot detect the IP addresses of pods generating telemetry 
data without any of the well-known IP attributes. If the Deployment Agent receives telemetry from 
Agents deployed as DaemonSet, then some of those attributes might be missing. As a workaround, 
you can configure the DaemonSet Agents with `passthrough` set to `true`.

## Blocks

The following blocks are supported inside the definition of
`otelcol.processor.k8sattributes`:
Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
output | [output][] | Configures where to send received telemetry data. | yes
extract | [extract][] | Rules for extracting data from Kubernetes. | no
extract > annotation | [annotation][] | Creating resource attributes from Kubernetes annotations. | no
extract > label | [extract_label][] | Creating resource attributes from Kubernetes labels. | no
filter | [filter][] | Filters the data loaded from Kubernetes. | no
filter > field | [field][] | Filter pods by generic Kubernetes fields. | no
filter > label | [filter_label][] | Filter pods by Kubernetes labels. | no
pod_association | [pod_association][] | Rules to associate pod metadata with telemetry signals. | no
pod_association > source | [source][] | Source information to identify a pod. | no
exclude | [exclude][] | Exclude pods from being processed. | no
exclude > pod | [pod][] | Pod information. | no


The `>` symbol indicates deeper levels of nesting. For example, `extract > annotation`
refers to an `annotation` block defined inside an `extract` block.

[output]: #output-block
[extract]: #extract-block
[annotation]: #annotation-block
[extract_label]: #extract-label-block
[filter]: #filter-block
[field]: #field-block
[filter_label]: #filter-label-block
[pod_association]: #pod_association-block
[source]: #source-block
[exclude]: #exclude-block
[pod]: #pod-block

### extract block

The `extract` block configures which metadata, annotations, and labels to extract from the pod.

The following attributes are supported:

Name | Type           | Description                          | Default     | Required
---- |----------------|--------------------------------------|-------------| --------
`metadata` | `list(string)` | Pre-configured metadata keys to add. | _See below_ | no

The currently supported `metadata` keys are:

* `k8s.pod.name`
* `k8s.pod.uid`
* `k8s.deployment.name`
* `k8s.node.name`
* `k8s.namespace.name`
* `k8s.pod.start_time`
* `k8s.replicaset.name`
* `k8s.replicaset.uid`
* `k8s.daemonset.name`
* `k8s.daemonset.uid`
* `k8s.job.name`
* `k8s.job.uid`
* `k8s.cronjob.name`
* `k8s.statefulset.name`
* `k8s.statefulset.uid`
* `k8s.container.name`
* `container.image.name`
* `container.image.tag`
* `container.id`

By default, if `metadata` is not specified, the following fields are extracted and added to spans, metrics, and logs as resource attributes:

* `k8s.pod.name`
* `k8s.pod.uid`
* `k8s.pod.start_time`
* `k8s.namespace.name`
* `k8s.node.name`
* `k8s.deployment.name` (if the pod is controlled by a deployment)
* `k8s.container.name` (requires an additional attribute to be set: `container.id`)
* `container.image.name` (requires one of the following additional attributes to be set: `container.id` or `k8s.container.name`)
* `container.image.tag` (requires one of the following additional attributes to be set: `container.id` or `k8s.container.name`)

### annotation block

The `annotation` block configures how to extract Kubernetes annotations.

{{< docs/shared lookup="flow/reference/components/extract-field-block.md" source="agent" version="<AGENT VERSION>" >}}

### label block {#extract-label-block}

The `label` block configures how to extract Kubernetes labels.

{{< docs/shared lookup="flow/reference/components/extract-field-block.md" source="agent" version="<AGENT VERSION>" >}}

### filter block

The `filter` block configures which nodes to get data from and which fields and labels to fetch.

The following attributes are supported:

Name | Type     | Description                                                             | Default | Required
---- |----------|-------------------------------------------------------------------------| ------- | --------
`node` | `string` | Configures a Kubernetes node name or host name. | `""` | no
`namespace` | `string` | Filters all pods by the provided namespace. All other pods are ignored. | `""` | no

If `node` is specified, then any pods not running on the specified node will be ignored by `otelcol.processor.k8sattributes`.

### field block

The `field` block allows you to filter pods by generic Kubernetes fields.

{{< docs/shared lookup="flow/reference/components/field-filter-block.md" source="agent" version="<AGENT VERSION>" >}}

### label block {#filter-label-block}

The `label` block allows you to filter pods by generic Kubernetes labels.

{{< docs/shared lookup="flow/reference/components/field-filter-block.md" source="agent" version="<AGENT VERSION>" >}}

### pod_association block

The `pod_association` block configures rules on how to associate logs/traces/metrics to pods.

The `pod_association` block does not support any arguments and is configured
fully through child blocks.

The `pod_association` block can be repeated multiple times, to configure additional rules.

Example:
```river
pod_association {
    source {
        from = "resource_attribute"
        name = "k8s.pod.ip"
    }
}

pod_association {
    source {
        from = "resource_attribute"
        name = "k8s.pod.uid"
    }
    source {
        from = "connection"
    }
}
```

### source block

The `source` block configures a pod association rule. This is used by the `k8sattributes` processor to determine the
pod associated with a telemetry signal.

When multiple `source` blocks are specified inside a `pod_association` block, both `source` blocks has to match for the
pod to be associated with the telemetry signal.

The following attributes are supported:

Name | Type     | Description                                                                      | Default | Required
---- |----------|----------------------------------------------------------------------------------| ------- | --------
`from` | `string` | The association method. Currently supports `resource_attribute` and `connection` |  | yes
`name` | `string` | Name represents extracted key name. For example, `ip`, `pod_uid`, `k8s.pod.ip`           |  | no


### exclude block

The `exclude` block configures which pods to exclude from the processor.

### pod block

The `pod` block configures a pod to be excluded from the processor.

The following attributes are supported:

Name | Type     | Description         | Default | Required
---- |----------|---------------------| ------- | --------
`name` | `string` | The name of the pod |  | yes

### output block

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
In most cases, this is enough to get started. It'll add these resource attributes to all logs, metrics, and traces:

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

otelcol.exporter.otlp "default" {
  client {
    endpoint = env("OTLP_ENDPOINT")
  }
}
```

### Add additional metadata and labels

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
