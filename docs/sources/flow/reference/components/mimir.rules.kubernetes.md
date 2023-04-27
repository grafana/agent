---
title: mimir.rules.kubernetes
---

# mimir.rules.kubernetes

`mimir.rules.kubernetes` discovers `PrometheusRule` Kubernetes resources and
loads them into a Mimir instance.

* Multiple `mimir.rules.kubernetes` components can be specified by giving them
  different labels.
* [Kubernetes label selectors][] can be used to limit the `Namespace` and
  `PrometheusRule` resources considered during reconciliation.
* Compatible with the Ruler APIs of Grafana Mimir, Grafana Cloud, and Grafana Enterprise Metrics.
* Compatible with the `PrometheusRule` CRD from the [prometheus-operator][].

[Kubernetes label selectors]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
[prometheus-operator]: https://prometheus-operator.dev/

## Usage

```river
mimir.rules.kubernetes "LABEL" {
  address = MIMIR_RULER_URL
}
```

## Arguments

`mimir.rules.kubernetes` supports the following arguments:

Name                     | Type       | Description                                             | Default | Required
-------------------------|------------|---------------------------------------------------------|---------|---------
`address`                | `string`   | URL of the Mimir ruler.                                 |         | yes
`tenant_id`              | `string`   | Mimir tenant ID.                                         |         | no
`use_legacy_routes`      | `bool`     | Whether to use deprecated ruler API endpoints.           | false   | no
`sync_interval`          | `duration` | Amount of time between reconciliations with Mimir.       | "30s"   | no
`mimir_namespace_prefix` | `string`   | Prefix used to differentiate multiple agent deployments. | "agent" | no

If no `tenant_id` is provided, the component assumes that the Mimir instance at
`address` is running in single-tenant mode and no `X-Scope-OrgID` header is sent.

The `sync_interval` argument determines how often Mimir's ruler API is accessed
to reload the current state of rules. Interaction with the Kubernetes API works
differently. Updates are processed as events from the Kubernetes API server
according to the informer pattern.

The `mimir_namespace_prefix` argument can be used to separate the rules managed
by multiple agent deployments across your infrastructure. It should be set to a
unique value for each deployment.

## Blocks

The following blocks are supported inside the definition of
`mimir.rules.kubernetes`:

Hierarchy                                  | Block                  | Description                                              | Required
-------------------------------------------|------------------------|----------------------------------------------------------|---------
rule_namespace_selector                    | [label_selector][]     | Label selector for `Namespace` resources.                 | no
rule_namespace_selector > match_expression | [match_expression][]   | Label match expression for `Namespace` resources.         | no
rule_selector                              | [label_selector][]     | Label selector for `PrometheusRule` resources.            | no
rule_selector > match_expression           | [match_expression][]   | Label match expression for `PrometheusRule` resources.    | no
http_client_config                         | [http_client_config][] | HTTP client settings when connecting to the endpoint.    | no
http_client_config > basic_auth            | [basic_auth][]         | Configure basic_auth for authenticating to the endpoint. | no
http_client_config > authorization         | [authorization][]      | Configure generic authorization to the endpoint.         | no
http_client_config > oauth2                | [oauth2][]             | Configure OAuth2 for authenticating to the endpoint.     | no
http_client_config > oauth2 > tls_config   | [tls_config][]         | Configure TLS settings for connecting to the endpoint.   | no
http_client_config > tls_config            | [tls_config][]         | Configure TLS settings for connecting to the endpoint.   | no


The `>` symbol indicates deeper levels of nesting. For example,
`http_client_config > basic_auth` refers to a `basic_auth` block defined inside
an `http_client_config` block.

[http_client_config]: #http_client_config-block
[basic_auth]: #basic_auth-block
[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block
[label_selector]: #label_selector-block
[match_expression]: #match_expression-block

### label_selector block

The `label_selector` block describes a Kubernetes label selector for rule or namespace discovery.

The following arguments are supported:

Name           | Type          | Description                                       | Default                     | Required
---------------|---------------|---------------------------------------------------|-----------------------------|---------
`match_labels` | `map(string)` | Label keys and values used to discover resources. | `{}` | yes

When the `match_labels` argument is empty, all resources will be matched. 

### match_expression block

The `match_expression` block describes a Kubernetes label match expression for rule or namespace discovery.

The following arguments are supported:

Name       | Type       | Description                                        | Default | Required
-----------|------------|----------------------------------------------------|---------|---------
`key`      | `string`   | The label name to match against.                   |         | yes
`operator` | `string`   | The operator to use when matching. |         | yes
`values`   | `[]string` | The values used when matching.                     |         | no

The `operator` argument should be one of the following strings:

* `"in"` 
* `"notin"` 
* `"exists"` 

### http_client_config block

The `http_client_config` configures settings used to connect to the Mimir API.

{{< docs/shared lookup="flow/reference/components/http-client-config-block.md" source="agent" >}}

### basic_auth block

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" >}}

### authorization block

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" >}}

### oauth2 block

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" >}}

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" >}}

## Exported fields

`mimir.rules.kubernetes` does not export any fields.

## Component health

`mimir.rules.kubernetes` is reported as unhealthy if given an invalid configuration or an error occurs during reconciliation.

## Debug information

`mimir.rules.kubernetes` exposes resource-level debug information.

The following are exposed per discovered `PrometheusRule` resource:
* The Kubernetes namespace.
* The resource name.
* The resource uid.
* The number of rule groups.

The following are exposed per discovered Mimir rule namespace resource:
* The namespace name.
* The number of rule groups.

Only resources managed by the component are exposed - regardless of how many
actually exist.

## Debug metrics

Metric Name                                   | Type        | Description
----------------------------------------------|-------------|-------------------------------------------------------------------------
`mimir_rules_config_updates_total`            | `counter`   | Number of times the configuration has been updated.
`mimir_rules_events_total`                    | `counter`   | Number of events processed, partitioned by event type.
`mimir_rules_events_failed_total`             | `counter`   | Number of events that failed to be processed, partitioned by event type.
`mimir_rules_events_retried_total`            | `counter`   | Number of events that were retried, partitioned by event type.
`mimir_rules_client_request_duration_seconds` | `histogram` | Duration of requests to the Mimir API.

## Example

This example creates a `mimir.rules.kubernetes` component that loads discovered
rules to a local Mimir instance under the `team-a` tenant. Only namespaces and
rules with the `agent` label set to `yes` are included.

```river
mimir.rules.kubernetes "local" {
    address = "mimir:8080"
    tenant_id = "team-a"

    rule_namespace_selector {
        match_labels = {
            agent = "yes",
        }
    }

    rule_selector {
        match_labels = {
            agent = "yes",
        }
    }
}
```

This example creates a `mimir.rules.kubernetes` component that loads discovered
rules to Grafana Cloud.

```river
mimir.rules.kubernetes "default" {
    address = "GRAFANA_CLOUD_METRICS_URL"
    http_client_config {
        basic_auth {
            username = "GRAFANA_CLOUD_USER"
            password = "GRAFANA_CLOUD_API_KEY"
            // Alternatively, load the password from a file:
            // password_file = "GRAFANA_CLOUD_API_KEY_PATH"
        }
    }
}
```