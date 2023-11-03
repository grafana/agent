---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/prometheus.operator.probes/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/prometheus.operator.probes/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/prometheus.operator.probes/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.operator.probes/
labels:
  stage: beta
title: prometheus.operator.probes
description: Learn about prometheus.operator.probes
---

# prometheus.operator.probes

{{< docs/shared lookup="flow/stability/beta.md" source="agent" version="<AGENT VERSION>" >}}

`prometheus.operator.probes` discovers [Probe](https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.Probe) resources in your Kubernetes cluster and scrapes the targets they reference. This component performs three main functions:

1. Discover Probe resources from your Kubernetes cluster.
2. Discover targets or ingresses that match those Probes.
3. Scrape metrics from those endpoints, and forward them to a receiver.

The default configuration assumes the agent is running inside a Kubernetes cluster, and uses the in-cluster config to access the Kubernetes API. It can be run from outside the cluster by supplying connection info in the `client` block, but network level access to pods is required to scrape metrics from them.

Probes may reference secrets for authenticating to targets to scrape them. In these cases, the secrets are loaded and refreshed only when the Probe is updated or when this component refreshes its' internal state, which happens on a 5-minute refresh cycle.

## Usage

```river
prometheus.operator.probes "LABEL" {
    forward_to = RECEIVER_LIST
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`forward_to` | `list(MetricsReceiver)` | List of receivers to send scraped metrics to. | | yes
`namespaces` | `list(string)` | List of namespaces to search for Probe resources. If not specified, all namespaces will be searched. || no

## Blocks

The following blocks are supported inside the definition of `prometheus.operator.probes`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
client | [client][] | Configures Kubernetes client used to find Probes. | no
client > basic_auth | [basic_auth][] | Configure basic authentication to the Kubernetes API. | no
client > authorization | [authorization][] | Configure generic authorization to the Kubernetes API. | no
client > oauth2 | [oauth2][] | Configure OAuth2 for authenticating to the Kubernetes API. | no
client > oauth2 > tls_config | [tls_config][] | Configure TLS settings for connecting to the Kubernetes API. | no
client > tls_config | [tls_config][] | Configure TLS settings for connecting to the Kubernetes API. | no
rule | [rule][] | Relabeling rules to apply to discovered targets. | no
scrape | [scrape][] | Default scrape configuration to apply to discovered targets. | no
selector | [selector][] | Label selector for which Probes to discover. | no
selector > match_expression | [match_expression][] | Label selector expression for which Probes to discover. | no
clustering | [clustering][] | Configure the component for when the Agent is running in clustered mode. | no

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
[rule]: #rule-block
[scrape]: #scrape-block
[clustering]: #clustering-experimental

### client block

The `client` block configures the Kubernetes client used to discover Probes. If the `client` block isn't provided, the default in-cluster
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

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" version="<AGENT VERSION>" >}}

### authorization block

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" version="<AGENT VERSION>" >}}

### oauth2 block

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" version="<AGENT VERSION>" >}}

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" version="<AGENT VERSION>" >}}

### rule block

{{< docs/shared lookup="flow/reference/components/rule-block.md" source="agent" version="<AGENT VERSION>" >}}

### scrape block

{{< docs/shared lookup="flow/reference/components/prom-operator-scrape.md" source="agent" version="<AGENT VERSION>" >}}

### selector block

The `selector` block describes a Kubernetes label selector for Probes.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`match_labels` | `map(string)` | Label keys and values used to discover resources. | `{}` | no

When the `match_labels` argument is empty, all Probe resources will be matched.

### match_expression block

The `match_expression` block describes a Kubernetes label matcher expression for
Probes discovery.

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

### clustering (experimental)

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `bool` | Enables sharing targets with other cluster nodes. | `false` | yes

When the agent is running in [clustered mode][], and `enabled` is set to true,
then this component instance opts-in to participating in
the cluster to distribute scrape load between all cluster nodes.

Clustering assumes that all cluster nodes are running with the same
configuration file, and that all
`prometheus.operator.probes` components that have opted-in to using clustering, over
the course of a scrape interval have the same configuration.

All `prometheus.operator.probes` components instances opting in to clustering use target
labels and a consistent hashing algorithm to determine ownership for each of
the targets between the cluster peers. Then, each peer only scrapes the subset
of targets that it is responsible for, so that the scrape load is distributed.
When a node joins or leaves the cluster, every peer recalculates ownership and
continues scraping with the new target set. This performs better than hashmod
sharding where _all_ nodes have to be re-distributed, as only 1/N of the
target's ownership is transferred, but is eventually consistent (rather than
fully consistent like hashmod sharding is).

If the agent is _not_ running in clustered mode, then the block is a no-op, and
`prometheus.operator.probes` scrapes every target it receives in its arguments.

[clustered mode]: {{< relref "../cli/run.md#clustering-beta" >}}

## Exported fields

`prometheus.operator.probes` does not export any fields. It forwards all metrics it scrapes to the receiver configures with the `forward_to` argument.

## Component health

`prometheus.operator.probes` is reported as unhealthy when given an invalid configuration, Prometheus components fail to initialize, or the connection to the Kubernetes API could not be established properly.

## Debug information

`prometheus.operator.probes` reports the status of the last scrape for each configured
scrape job on the component's debug endpoint, including discovered labels, and the last scrape time.

It also exposes some debug information for each Probe it has discovered, including any errors found while reconciling the scrape configuration from the Probe.

## Debug metrics

`prometheus.operator.probes` does not expose any component-specific debug metrics.

## Example

This example discovers all Probes in your cluster, and forwards collected metrics to a `prometheus.remote_write` component.

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

prometheus.operator.probes "pods" {
    forward_to = [prometheus.remote_write.staging.receiver]
}
```

This example will limit discovered Probes to ones with the label `team=ops` in a specific namespace: `my-app`.

```river
prometheus.operator.probes "pods" {
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

This example will apply additional relabel rules to discovered targets to filter by hostname. This may be useful if running the agent as a DaemonSet.

```river
prometheus.operator.probes "probes" {
    forward_to = [prometheus.remote_write.staging.receiver]
    rule {
      action = "keep"
      regex = env("HOSTNAME")
      source_labels = ["__meta_kubernetes_pod_node_name"]
    }
}
```
