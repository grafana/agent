---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/prometheus.test.metrics/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.test.metrics/
description: Learn about prometheus.test.metrics  
labels:
  stage: experimental
title: prometheus.test.metrics
---

# prometheus.test.metrics
 

`prometheus.test.metrics` configures a locally hosted service discovery that then exposes endpoints 
for collecting generated metrics. It is best used for testing perfomance and will always be considered
experimental.

Multiple `prometheus.test.metrics` components can be specified by giving them
different labels.

## Usage

```
prometheus.test.metrics "LABEL" {
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`number_of_instances` | `int`     | Number of instances to provide. | `1` | no
`number_of_metrics` | `int`     | Number of metrics to provide. | `1` | no  
`number_of_labels` | `int`     | Number of labels to provide per metric. | `0` | no
`metrics_refresh` | `duration`     | How often to refresh metrics and labels. | `1m` | no  
`churn_percent` | `float`     | What percentage of metrics to churn during a refresh, must be between 0 and 1. | `0` | no  

## Exported fields

The following fields are exported and can be referenced by other components:

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The set of targets discovered from the component.


## Component health

`prometheus.test.metrics` is only reported as unhealthy if given an invalid
configuration.

## Example

```river

prometheus.test.metrics "single" {
    number_of_instances = 1
    number_of_metrics = 1000
    number_of_labels = 5
    metrics_refresh = "10m"
    churn_percent = 0.05
}

prometheus.scrape "default" {
  targets = concat(prometheus.test.metrics.single.targets)
  forward_to = [prometheus.remote_write.default.receiver]
}


prometheus.remote_write "default" {
    endpoint {
        url = "http://mimir:9009/api/v1/push"
        basic_auth {
            username = "example-user"
            password = "example-password"
        } 
    }
}
```


