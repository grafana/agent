---
aliases:
- /docs/agent/latest/flow/reference/components/prometheus.relabel
title: prometheus.relabel
---

# prometheus.relabel

Prometheus metrics follow the [OpenMetrics](https://openmetrics.io/) format.
Each time series is uniquely identified by its metric name, plus optional
key-value pairs called labels. Each sample, i.e. a datapoint in the time
series, contains a value and an optional timestamp.
```
<metric name>{<label_1>=<label_val_1>, <label_2>=<label_val_2> ...} <value> [timestamp]
```

The `prometheus.relabel` component rewrites the label set of each metric passed
along to the exported receiver by applying one or more `metric_relabel_config`
steps.  If no relabeling steps are defined or applicable to one of the metrics,
then the metric will be forwarded as-is to each receiver passed in the
component's arguments. If no labels are remaining after the relabeling steps
are applied, then the metric will be dropped.

The most common use of `prometheus.relabel` is to filter Prometheus metrics or
standardize the label set that will be passed to one or more downstream
receivers. The `metric_relabel_config` blocks will be applied to the label set
of each metric in order of their appearance in the configuration file.

Labels beginning with a double underscore are reserved for internal used and
will be silently discarded after all relabeling steps have been applied.

Multiple `prometheus.relabel` components can be specified by giving them
different labels.

## Example

```river
prometheus.relabel "keep_mysql_only" {
  forward_to [prometheus.remote_write.onprem.receiver]

  metric_relabel_config {
    source_labels = ["__address__", "instance"]
    separator     = "/"
    target_label  = "host"
    action        = "replace"
  }

  metric_relabel_config {
    source_labels = ["app"]
    action        = "keep"
    regex         = "backend"
  }

  metric_relabel_config {
    action = "labeldrop"
    regex  = "instance"
  }
}
```

Let's see how the previous instance of the `prometheus.relabel` component would
act on some metrics.

```
metric_a{__address__ = "localhost", instance = "development", app = "frontend"} 10
metric_a{__address__ = "localhost", instance = "development", app = "backend"}	2
metric_a{__address__ = "cluster_a", instance = "production",  app = "frontend"} 7
metric_a{__address__ = "cluster_a", instance = "production",  app = "backend"}	9
metric_a{__address__ = "cluster_b", instance = "production",  app = "database"}	4
```

After applying the first `metric_relabel_config` block, the `replace` action
would populate a new label named `host` by the concatenating the two
`__address__` and `instance` labels separated by a slash `/`.

```
metric_a{host = "localhost/development", __address__ = "localhost", instance = "development", app = "frontend"} 10
metric_a{host = "localhost/development", __address__ = "localhost", instance = "development", app = "backend"}	2
metric_a{host = "cluster_a/production",  __address__ = "cluster_a", instance = "production",  app = "frontend"} 7
metric_a{host = "cluster_a/production",  __address__ = "cluster_a", instance = "production",  app = "backend"}	9
metric_a{host = "cluster_b/production",  __address__ = "cluster_a", instance = "production",  app = "database"}	4
```

After applying the second relabeling step, the `keep` action would only keep
the metrics whose `app` label match `regex`, dropping everything else, so the
list of metrics would be trimmed down to 

```
metric_a{host = "localhost/development", __address__ = "localhost", instance = "development", app = "backend"}	2
metric_a{host = "cluster_a/production",  __address__ = "cluster_a", instance = "production",  app = "backend"}	9
```

The third and final relabeling step which uses the `labeldrop` action would
remove the `instance` label from the set of labels

```
metric_a{host = "localhost/development", __address__ = "localhost", app = "backend"}	2
metric_a{host = "cluster_a/production",  __address__ = "cluster_a", app = "backend"}	9
```

Finally, the `__address__` label, as it starts with a double underscore and is
reserved as an 'internal' label and would be implicitly removed.

After filtering down the samples passed through the exported receiver, these
two final metrics would then be propagated to each receiver passed in the
`forward_to` argument.
```
metric_a{host = "localhost/development", app = "backend"}	2
metric_a{host = "cluster_a/production",  app = "backend"}	9
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
forward_to | list(receiver) | Where the metrics should be forwarded to after relabeling takes place | | **yes**

The following subblocks are supported:

Name | Description | Required
---- | ----------- | --------
[`metric_relabel_config`](#metric_relabel_config-block) | Relabeling steps to
apply to received metrics | no

### `metric_relabel_config` block

The `metric_relabel_config` block contains the definition of any relabeling
rules that can be applied to an input metric. If more than one
`metric_relabel_config` block is defined within `prometheus.relabel`, the
transformations will be applied in top-down order.

The following arguments can be used to configure a `metric_relabel_config`
block.
All arguments are optional and any omitted fields will take on their default
values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
source_labels | list(string) | The list of labels whose values should be selected. Their content is concatenated using the `separator` and matched against `regex`. | | no
separator     | string       |  The separator used to concatenate the values present in `source_labels`. | ; | no
regex         | string       | A valid RE2 expression with support for parenthesized capture groups. Used to match the extracted value from the combination of the `source_label` and `separator` fields or filter labels during the labelkeep/labeldrop/labelmap actions. | `(.*)` | no
modulus       | uint         | A positive integer used to calculate the modulus of the hashed source label values. | | no
target_label  | string       | Label to which the resulting value will be written to. | | no
replacement   | string       | The value against which a regex replace is performed, if the regex matched the extracted value. Supports previously captured groups. | $1 | no
action        | string       | The relabeling action to perform. | replace | no

Here's a list of the available actions along with a brief description of their usage.

* replace - This action matches `regex` to the concatenated labels. If there's a match, it replaces the content of the `target_label` using the contents of the `replacement` field.
* keep    - This action only keeps the metrics where `regex` matches the string extracted using the `source_labels` and `separator`.
* drop    - This action drops the metrics where `regex` matches the string extracted using the `source_labels` and `separator`.
* hashmod - This action hashes the concatenated labels, calculates its modulo `modulus` and writes the result to the `target_label`.
* labelmap  - This action matches `regex` against all label names. Any labels that match will be renamed according to the contents of the `replacement` field.
* labeldrop - This action matches `regex` against all label names. Any labels that match will be removed from the metric's label set.
* labelkeep - This action matches `regex` against all label names. Any labels that don't match will be removed from the metric's label set.

Finally, note that the regex capture groups can be referred to using either the `$1` or `$${1}` notation.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
receiver | receiver | The receiver where samples should be sent to, in order to be relabeled.

## Component health

`prometheus.relabel` will only be reported as unhealthy when given an invalid
configuration. In those cases, exported fields will be kept at their last
healthy values.

## Debug information

`prometheus.relabel` does not expose any component-specific debug information.

### Debug metrics

`prometheus.relabel` does not expose any component-specific debug metrics.
