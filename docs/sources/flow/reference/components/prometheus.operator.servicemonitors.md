---
title: prometheus.operator.servicemonitors
labels:
  stage: beta
---

# prometheus.operator.servicemonitors

{{< docs/shared lookup="flow/stability/beta.md" source="agent" >}}

`prometheus.operator.servicemonitors` discovers [ServiceMonitor](https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.ServiceMonitor) resources in your kubernetes cluster and scrapes the targets they reference. This component performs three main functions:

1. Discover ServiceMonitor resources from your Kubernetes cluster.
2. Discover Services and Endpoints in your cluster that match those ServiceMonitors.
3. Scrape metrics from those Endpoints, and forward them to a receiver.

The default configuration assumes the agent is running inside a Kubernetes cluster, and uses the in-cluster config to access the Kubernetes API. It can be run from outside the cluster by supplying connection info in the `client` block, but network level access to discovered endpoints is required to scrape metrics from them.

ServiceMonitors may reference secrets for authenticating to targets to scrape them. In these cases, the secrets are loaded and refreshed only when the ServiceMonitor is updated or when this component refreshes its' internal state, which happens on a 5-minute refresh cycle.

## Usage

```river
prometheus.operator.servicemonitors "LABEL" {
    forward_to = RECEIVER_LIST
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`forward_to` | `list(MetricsReceiver)` | List of receivers to send scraped metrics to. | | yes
`namespaces` | `list(string)` | List of namespaces to search for ServiceMonitor resources. If not specified, all namespaces will be searched. || no

## Blocks

The following blocks are supported inside the definition of `prometheus.operator.servicemonitors`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
client | [client][] | Configures Kubernetes client used to find ServiceMonitors. | no
client > basic_auth | [basic_auth][] | Configure basic authentication to the Kubernetes API. | no
client > authorization | [authorization][] | Configure generic authorization to the Kubernetes API. | no
client > oauth2 | [oauth2][] | Configure OAuth2 for authenticating to the Kubernetes API. | no
client > oauth2 > tls_config | [tls_config][] | Configure TLS settings for connecting to the Kubernetes API. | no
client > tls_config | [tls_config][] | Configure TLS settings for connecting to the Kubernetes API. | no
relabel | [relabel][] | Relabeling rules to apply to discovered targets. | no
selector | [selector][] | Label selector for which ServiceMonitors to discover. | no
selector > match_expression | [match_expression][] | Label selector expression for which ServiceMonitors to discover. | no

The `>` symbol indicates deeper levels of nesting. For example, `client >
basic_auth` refers to a `basic_auth` block defined
inside a `client` block.

[client]: #client-block
[basic_auth]: #basic_auth-block
[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block
[selector]: #selector-block
[match_expression]: #match_expression-block
[relabel]: #relabel-block

### client block

The `client` block configures the Kubernetes client used to discover ServiceMonitors. If the `client` block isn't provided, the default in-cluster
configuration with the service account of the running Grafana Agent pod is
used.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`api_server` | `string` | URL of the Kubernetes API server. | | no
`kubeconfig_file` | `string` | Path of the `kubeconfig` file to use for connecting to Kubernetes. | | no
`bearer_token_file` | `string` | File containing a bearer token to authenticate with. | | no
`proxy_url` | `string` | HTTP proxy to proxy requests through. | | no
`follow_redirects` | `bool` | Whether redirects returned by the server should be followed. | `true` | no
`enable_http2` | `bool` | Whether HTTP2 is supported for requests. | `true` | no

 At most one of the following can be provided:
 - [`bearer_token` argument][client].
 - [`bearer_token_file` argument][client].
 - [`basic_auth` block][basic_auth].
 - [`authorization` block][authorization].
 - [`oauth2` block][oauth2].

### basic_auth block

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" >}}

### authorization block

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" >}}

### oauth2 block

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" >}}

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" >}}

### relabel block

{{< docs/shared lookup="flow/reference/components/rule-block.md" source="agent" >}}

### selector block

The `selector` block describes a Kubernetes label selector for ServiceMonitors.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`match_labels` | `map(string)` | Label keys and values used to discover resources. | `{}` | no

When the `match_labels` argument is empty, all ServiceMonitor resources will be matched.

### match_expression block

The `match_expression` block describes a Kubernetes label matcher expression for
ServiceMonitors discovery.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`key` | `string` | The label name to match against. | | yes
`operator` | `string` | The operator to use when matching. | | yes
`values`| `list(string)` | The values used when matching. | | no

The `operator` argument must be one of the following strings:

* `"In"`
* `"NotIn"`
* `"Exists"`
* `"DoesNotExist"`

If there are multiple `match_expressions` blocks inside of a `selector` block, they are combined together with AND clauses. 

## Exported fields

`prometheus.operator.servicemonitors` does not export any fields. It forwards all metrics it scrapes to the receiver configures with the `forward_to` argument.

## Component health

`prometheus.operator.servicemonitors` is reported as unhealthy when given an invalid configuration, Prometheus components fail to initialize, or the connection to the Kubernetes API could not be established properly.

## Debug information

`prometheus.operator.servicemonitors` reports the status of the last scrape for each configured
scrape job on the component's debug endpoint, including discovered labels, and the last scrape time.

It also exposes some debug information for each ServiceMonitor it has discovered, including any errors found while reconciling the scrape configuration from the ServiceMonitor.

### Debug metrics


## Example

This example discovers all ServiceMonitors in your cluster, and forwards collected logs to a
`prometheus.remote_write` component.

```river
prometheus.remote_write "staging" {
  // Send metrics to a locally running Mimir.
  endpoint {
    url = "http://mimir:9009/api/v1/push"

    basic_auth {
      username = "example-user"
      password = "example-password"
    }
  }
}

prometheus.operator.servicemonitors "services" {
    forward_to = [prometheus.remote_write.staging.receiver]
}
```

This example will limit discovered ServiceMonitors to ones with the label `team=ops` in a specific namespace: `my-app`.

```river
prometheus.operator.servicemonitors "services" {
    forward_to = [prometheus.remote_write.staging.receiver]
    namespaces = ["my-app"]
    selector {
        match_expression {
            key = "team"
            operator = "In"
            values = ["ops"]
        }
    }
}
```

This example will apply additional relabel rules to discovered targets for hashmod sharding. Here we are have 5 shards, and an environment variable is set up to give our current agent's shard.

```river
prometheus.operator.servicemonitors "services" {
    forward_to = [prometheus.remote_write.staging.receiver]
    relabel {
      action = "hashmod"
      modulus = 5
      source_labels = ["__address__"]
      target_label = "__tmp_hash"
    }
    relabel {
      action = "keep"
      regex = env("HASHMOD_SHARD")
      source_labels = ["__tmp_hash"]
    }
    relabel {
      action = "labeldrop"
      regex = "__tmp_hash"
    }
}
```