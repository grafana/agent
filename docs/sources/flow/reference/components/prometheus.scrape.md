---
title: prometheus.scrape
---

# prometheus.scrape

`prometheus.scrape` configures a Prometheus scraping job for a given set of
`targets`. The scraped metrics are forwarded to the list of receivers passed in
`forward_to`.

Multiple `prometheus.scrape` components can be specified by giving them
different labels.

## Usage

```
prometheus.scrape "LABEL" {
  targets    = TARGET_LIST
  forward_to = RECEIVER_LIST
}
```

## Arguments

The component configures and starts a new scrape job to scrape all of the
input targets. The list of arguments that can be used to configure the block is
presented below.

The scrape job name defaults to the component's unique identifier.

Any omitted fields take on their default values. In case that conflicting
attributes are being passed (eg. defining both a BearerToken and
BearerTokenFile or configuring both Basic Authorization and OAuth2 at the same
time), the component reports an error.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`targets`                  | `list(map(string))`     | List of targets to scrape. | | yes
`forward_to`               | `list(MetricsReceiver)` | List of receivers to send scraped metrics to. | | yes
`job_name`                 | `string`   | The job name to override the job label with. | component name | no
`extra_metrics`            | `bool`     | Whether extra metrics should be generated for scrape targets. | `false` | no
`honor_labels`             | `bool`     | Indicator whether the scraped metrics should remain unmodified. | `false` | no
`honor_timestamps`         | `bool`     | Indicator whether the scraped timestamps should be respected. | `true` | no
`params`                   | `map(list(string))` | A set of query parameters with which the target is scraped. | | no
`scrape_interval`          | `duration` | How frequently to scrape the targets of this scrape config. | `"60s"` | no
`scrape_timeout`           | `duration` | The timeout for scraping targets of this config. | `"10s"` | no
`metrics_path`             | `string`   | The HTTP resource path on which to fetch metrics from targets. | `/metrics` | no
`scheme`                   | `string`   | The URL scheme with which to fetch metrics from targets. | | no
`body_size_limit`          | `int`      | An uncompressed response body larger than this many bytes causes the scrape to fail. 0 means no limit. | | no
`sample_limit`             | `uint`     | More than this many samples post metric-relabeling causes the scrape to fail | | no
`target_limit`             | `uint`     | More than this many targets after the target relabeling causes the scrapes to fail. | | no
`label_limit`              | `uint`     | More than this many labels post metric-relabeling causes the scrape to fail. | | no
`label_name_length_limit`  | `uint`     | More than this label name length post metric-relabeling causes the scrape to fail. | | no
`label_value_length_limit` | `uint`     | More than this label value length post metric-relabeling causes the scrape to fail. | | no
`bearer_token` | `secret` | Bearer token to authenticate with. | | no
`bearer_token_file` | `string` | File containing a bearer token to authenticate with. | | no
`proxy_url` | `string` | HTTP proxy to proxy requests through. | | no
`follow_redirects` | `bool` | Whether redirects returned by the server should be followed. | `true` | no
`enable_http2` | `bool` | Whether HTTP2 is supported for requests. | `true` | no

 At most one of the following can be provided:
 - [`bearer_token` argument](#arguments).
 - [`bearer_token_file` argument](#arguments). 
 - [`basic_auth` block][basic_auth].
 - [`authorization` block][authorization].
 - [`oauth2` block][oauth2].

## Blocks

The following blocks are supported inside the definition of `prometheus.scrape`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
basic_auth | [basic_auth][] | Configure basic_auth for authenticating to targets. | no
authorization | [authorization][] | Configure generic authorization to targets. | no
oauth2 | [oauth2][] | Configure OAuth2 for authenticating to targets. | no
oauth2 > tls_config | [tls_config][] | Configure TLS settings for connecting to targets via OAuth2. | no
tls_config | [tls_config][] | Configure TLS settings for connecting to targets. | no

The `>` symbol indicates deeper levels of nesting. For example,
`oauth2 > tls_config` refers to a `tls_config` block defined inside
an `oauth2` block.

[arguments]: #arguments
[basic_auth]: #basic_auth-block
[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block

### basic_auth block

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" >}}

### authorization block

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" >}}

### oauth2 block

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" >}}

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" >}}

## Exported fields

`prometheus.scrape` does not export any fields that can be referenced by other
components.

## Component health

`prometheus.scrape` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`prometheus.scrape` reports the status of the last scrape for each configured
scrape job on the component's debug endpoint.

## Debug metrics

* `agent_prometheus_fanout_latency` (histogram): Write latency for sending to direct and indirect components.
* `agent_prometheus_scrape_targets_gauge` (gauge): Number of targets this component is configured to scrape.
* `agent_prometheus_forwarded_samples_total` (counter): Total number of samples sent to downstream components.

## Scraping behavior

The `prometheus.scrape` component borrows the scraping behavior of Prometheus.
Prometheus, and by extent this component, uses a pull model for scraping
metrics from a given set of _targets_.
Each scrape target is defined as a set of key-value pairs called _labels_.
The set of targets can either be _static_, or dynamically provided periodically
by a service discovery component such as `discovery.kubernetes`. The special
label `__address__` _must always_ be present and corresponds to the
`<host>:<port>` that is used for the scrape request.

By default, the scrape job tries to scrape all available targets' `/metrics`
endpoints using HTTP, with a scrape interval of 1 minute and scrape timeout of
10 seconds. The metrics path, protocol scheme, scrape interval and timeout,
query parameters, as well as any other settings can be configured using the
component's arguments.

The scrape job expects the metrics exposed by the endpoint to follow the
[OpenMetrics](https://openmetrics.io/) format. All metrics are then propagated
to each receiver listed in the component's `forward_to` argument.

Labels coming from targets, that start with a double underscore `__` are
treated as _internal_, and are removed prior to scraping.

The `prometheus.scrape` component regards a scrape as successful if it
responded with an HTTP `200 OK` status code and returned a body of valid
metrics.

If the scrape request fails, the component's debug UI section contains more
detailed information about the failure, the last successful scrape, as well as
the labels last used for scraping.

The following labels are automatically injected to the scraped time series and
can help pin down a scrape target.

Label                 | Description
--------------------- | ----------
job                   | The configured job name that the target belongs to. Defaults to the fully formed component name.
instance              | The `__address__` or `<host>:<port>` of the scrape target's URL.


Similarly, these metrics that record the behavior of the scrape targets are
also automatically available.
Metric Name                | Description
-------------------------- | -----------
`up`                       | 1 if the instance is healthy and reachable, or 0 if the scrape failed.
`scrape_duration_seconds`  | Duration of the scrape in seconds.
`scrape_samples_scraped`   | The number of samples the target exposed.
`scrape_samples_post_metric_relabeling` | The number of samples remaining after metric relabeling was applied.
`scrape_series_added`      | The approximate number of new series in this scrape.
`scrape_timeout_seconds`   | The configured scrape timeout for a target. Useful for measuring how close a target was to timing out using `scrape_duration_seconds / scrape_timeout_seconds`
`scrape_sample_limit`      | The configured sample limit for a target. Useful for measuring how close a target was to reaching the sample limit using `scrape_samples_post_metric_relabeling / (scrape_sample_limit > 0)`
`scrape_body_size_bytes`   | The uncompressed size of the most recent scrape response, if successful. Scrapes failing because the `body_size_limit` is exceeded report -1, other scrape failures report 0.

The `up` metric is particularly useful for monitoring and alerting on the
health of a scrape job. It is set to `0` in case anything goes wrong with the
scrape target, either because it is not reachable, because the connection
times out while scraping, or because the samples from the target could not be
processed. When the target is behaving normally, the `up` metric is set to
`1`.

## Example

The following example sets up the scrape job with certain attributes (scrape
endpoint, scrape interval, query parameters) and lets it scrape two instances
of the [blackbox exporter](https://github.com/prometheus/blackbox_exporter/).
The exposed metrics are sent over to the provided list of receivers, as
defined by other components.

```river
prometheus.scrape "blackbox_scraper" {
  targets = [
    {"__address__" = "blackbox-exporter:9115", "instance" = "one"},
    {"__address__" = "blackbox-exporter:9116", "instance" = "two"},
  ]

  forward_to = [prometheus.remote_write.grafanacloud.receiver, prometheus.remote_write.onprem.receiver]

  scrape_interval = "10s"
  params          = { "target" = ["grafana.com"], "module" = ["http_2xx"] }
  metrics_path    = "/probe"
}
```

Here's the the endpoints that are being scraped every 10 seconds:
```
http://blackbox-exporter:9115/probe?target=grafana.com&module=http_2xx
http://blackbox-exporter:9116/probe?target=grafana.com&module=http_2xx
```


