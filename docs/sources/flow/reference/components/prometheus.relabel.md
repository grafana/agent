---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/prometheus.relabel/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/prometheus.relabel/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/prometheus.relabel/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.relabel/
title: prometheus.relabel
description: Learn about prometheus.relabel
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
standardize the label set that is passed to one or more downstream
receivers. The `rule` blocks are applied to the label set of each metric in
order of their appearance in the configuration file. The configured rules can
be retrieved by calling the function in the `rules` export field.

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
`forward_to` | `list(receiver)` | Where the metrics should be forwarded to, after relabeling takes place. | | yes

## Blocks

The following blocks are supported inside the definition of `prometheus.relabel`:

Hierarchy | Name | Description | Required
--------- | ---- | ----------- | --------
rule | [rule][] | Relabeling rules to apply to received metrics. | no

[rule]: #rule-block

### rule block

{{< docs/shared lookup="flow/reference/components/rule-block.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`receiver` | `receiver` | The input receiver where samples are sent to be relabeled.
`rules`    | `RelabelRules` | The currently configured relabeling rules.

## Component health

`prometheus.relabel` is only reported as unhealthy if given an invalid
configuration. In those cases, exported fields are kept at their last healthy
values.

## Debug information

`prometheus.relabel` does not expose any component-specific debug information.

## Debug metrics


* `agent_prometheus_relabel_metrics_processed` (counter): Total number of metrics processed.
* `agent_prometheus_relabel_metrics_written` (counter): Total number of metrics written.
* `agent_prometheus_relabel_cache_misses` (counter): Total number of cache misses.
* `agent_prometheus_relabel_cache_hits` (counter): Total number of cache hits.
* `agent_prometheus_relabel_cache_size` (gauge): Total size of relabel cache.
* `agent_prometheus_fanout_latency` (histogram): Write latency for sending to direct and indirect components.
* `agent_prometheus_forwarded_samples_total` (counter): Total number of samples sent to downstream components.

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
is trimmed down to:

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
