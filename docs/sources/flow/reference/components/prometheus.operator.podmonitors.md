---
title: prometheus.operator.podmonitors
---

# prometheus.operator.podmonitors

`prometheus.operator.podmonitors` discovers [podMonitor]() resources in your kubernetes cluster and scrape the targets they reference.


## Usage

```river
prometheus.operator.podmonitors "LABEL" {
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`kubeconfig_file` | `string` | Path of the file on disk to a kubernetes config  | | no
`forward_to` | `list(MetricsReceiver)` | List of receivers to send scraped metrics to. | | yes
`namespaces` | `list(string)` | List of namespaces to search for PodMonitor resources. If not specified, all namespaces will be searched. || no
`label_selector` | string | [LabelSelector][] to filter which PodMonitor resources are discovered. || no
`field_selector` | string | [FieldSelector][] to filter which PodMonitor resources are discovered. || no

[LabelSelector]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
[FieldSelector]: https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/

## Blocks

The following blocks are supported inside the definition of `prometheus.operator.podmonitors`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
api_server | [api_server][] | Configure how to connect to the kubernetes api server. | no

### api_server block

The `api_server` block configures the profiling settings when scraping
targets.

The block contains the following attributes:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`host` | `string` | A valid string consisting of a hostname or IP followed by an optional port number. | | yes


## Exported fields

`prometheus.operator.podmonitors` does not export any fields. It will forward all metrics it scrapes to the receiver configures with the `forward_to` argument.

## Component health



## Debug information


### Debug metrics


## Example
