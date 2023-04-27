---
title: prometheus.relabel
---

# prometheus.relabel

Prometheus metrics follow the [OpenMetrics](https://openmetrics.io/) format.
Each time series is uniquely identified by its metric name, plus optional
key-value pairs called labels. Each sample represents a datapoint in the
time series and contains a value and an optional timestamp.
```
<metric name>{<label_1>=<label_val_1>, <label_2>=<label_val_2> ...} <value> [timestamp]
```

The `prometheus.relabel` component rewrites the label set of each metric passed
along to the exported receiver by applying one or more relabeling `rule`s. If
no rules are defined or applicable to some metrics, then those metrics are
forwarded as-is to each receiver passed in the component's arguments. If no
labels remain after the relabeling rules are applied, then the metric is
dropped.

The most common use of `prometheus.relabel` is to filter Prometheus metrics or
standardize the label set that will be passed to one or more downstream
receivers. The `rule` blocks are applied to the label set of
each metric in order of their appearance in the configuration file.

Multiple `prometheus.relabel` components can be specified by giving them
different labels.

## Usage

```river
prometheus.relabel "LABEL" {
  forward_to = RECEIVER_LIST

  rule {
    ...
  }

  ...
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`forward_to` | `list(receiver)` | Where the metrics should be forwarded to, after relabeling takes place. | | **yes**

## Blocks

The following blocks are supported inside the definition of `prometheus.relabel`:

Hierarchy | Name | Description | Required
--------- | ---- | ----------- | --------
rule | [rule][] | Relabeling rules to apply to received metrics. | no

[rule]: #rule-block

### rule block

The `rule` block contains the definition of any relabeling
rules that can be applied to an input metric. If more than one
`rule` block is defined within `prometheus.relabel`, the
transformations are applied in top-down order.

The following arguments can be used to configure a `rule`.
All arguments are optional. Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`source_labels` | `list(string)` | The list of labels whose values are to be selected. Their content is concatenated using the `separator` and matched against `regex`. | | no
`separator`     | `string`       | The separator used to concatenate the values present in `source_labels`. | ; | no
`regex`         | `string`       | A valid RE2 expression with support for parenthesized capture groups. Used to match the extracted value from the combination of the `source_label` and `separator` fields or filter labels during the `labelkeep/labeldrop/labelmap` actions. | `(.*)` | no
`modulus`       | `uint`         | A positive integer used to calculate the modulus of the hashed source label values. | | no
`target_label`  | `string`       | Label to which the resulting value will be written to. | | no
`replacement`   | `string`       | The value against which a regex replace is performed, if the regex matches the extracted value. Supports previously captured groups. | $1 | no
`action`        | `string`       | The relabeling action to perform. | replace | no

Here's a list of the available actions, along with a brief description of their usage.

* `replace`   - Matches `regex` to the concatenated labels. If there's a match, it replaces the content of the `target_label` using the contents of the `replacement` field.
* `keep`      - Keeps metrics where `regex` matches the string extracted using the `source_labels` and `separator`.
* `drop`      - Drops metrics where `regex` matches the string extracted using the `source_labels` and `separator`.
* `hashmod`   - Hashes the concatenated labels, calculates its modulo `modulus` and writes the result to the `target_label`.
* `labelmap`  - Matches `regex` against all label names. Any labels that match are renamed according to the contents of the `replacement` field.
* `labeldrop` - Matches `regex` against all label names. Any labels that match are removed from the metric's label set.
* `labelkeep` - Matches `regex` against all label names. Any labels that don't match are removed from the metric's label set.

Finally, note that the regex capture groups can be referred to using either the
`$1` or `$${1}` notation.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`receiver` | `receiver` | The input receiver where samples are sent to be relabeled.

## Component health

`prometheus.relabel` is only reported as unhealthy if given an invalid
configuration. In those cases, exported fields are kept at their last healthy
values.

## Debug information

`prometheus.relabel` does not expose any component-specific debug information.

## Debug metrics

`prometheus.relabel` does not expose any component-specific debug metrics.

## Example

Let's create an instance of a see `prometheus.relabel` component and see how
it acts on the following metrics.

```river
prometheus.relabel "keep_backend_only" {
  forward_to = [prometheus.remote_write.onprem.receiver]

  rule {
    action        = "replace"
    source_labels = ["__address__", "instance"]
    separator     = "/"
    target_label  = "host"
  }
  rule {
    action        = "keep"
    source_labels = ["app"]
    regex         = "backend"
  }
  rule {
    action = "labeldrop"
    regex  = "instance"
  }
}
```

```
metric_a{__address__ = "localhost", instance = "development", app = "frontend"} 10
metric_a{__address__ = "localhost", instance = "development", app = "backend"}  2
metric_a{__address__ = "cluster_a", instance = "production",  app = "frontend"} 7
metric_a{__address__ = "cluster_a", instance = "production",  app = "backend"}  9
metric_a{__address__ = "cluster_b", instance = "production",  app = "database"} 4
```

After applying the first `rule`, the `replace` action populates a new label
named `host` by concatenating the contents of the `__address__` and `instance`
labels, separated by a slash `/`.

```
metric_a{host = "localhost/development", __address__ = "localhost", instance = "development", app = "frontend"} 10
metric_a{host = "localhost/development", __address__ = "localhost", instance = "development", app = "backend"}  2
metric_a{host = "cluster_a/production",  __address__ = "cluster_a", instance = "production",  app = "frontend"} 7
metric_a{host = "cluster_a/production",  __address__ = "cluster_a", instance = "production",  app = "backend"}  9
metric_a{host = "cluster_b/production",  __address__ = "cluster_a", instance = "production",  app = "database"} 4
```

On the second relabeling rule, the `keep` action only keeps the metrics whose
`app` label matches `regex`, dropping everything else, so the list of metrics
is be trimmed down to:

```
metric_a{host = "localhost/development", __address__ = "localhost", instance = "development", app = "backend"}  2
metric_a{host = "cluster_a/production",  __address__ = "cluster_a", instance = "production",  app = "backend"}  9
```

The third and final relabeling rule which uses the `labeldrop` action removes
the `instance` label from the set of labels.

So in this case, the initial set of metrics passed to the exported receiver is:
```
metric_a{host = "localhost/development", __address__ = "localhost", app = "backend"}  2
metric_a{host = "cluster_a/production",  __address__ = "cluster_a", app = "backend"}  9
```

The two resulting metrics are then propagated to each receiver defined in the
`forward_to` argument.
