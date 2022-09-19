---
aliases:
- /docs/agent/latest/flow/reference/components/discovery.kubernetes
title: discovery.kubernetes
---

# discovery.kubernetes

`discovery.kubernetes` allows you to find scrape targets from Kubernetes resources. It watches cluster state, and ensures targets are continually synced with what is currently running in your cluster.

Internally, this component uses Kubernetes Service Discovery from Prometheus. [The documentation](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config) on the Prometheus site offers further details on the metadata labels applied to targets and other usage considerations.

If you supply no connection information, this component defaults to using an in-cluster config. Otherwise, you can specify a kubeconfig file, or manual apiserver configuration.

## Example

Simple in-cluster discovery of all pods:

```river
discovery.kubernetes "k8s_pods" {
  role = "pod"
}
```

Specific namespace only with kubeconfig file:

```river
discovery.kubernetes "k8s_pods" {
  role = "pod"
  kubeconfig_file = "/path/to/kubeconfig"
  namepsaces {
    names = ["myapp"]
  }
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`api_server` | `string` | URL of kubernetes API server | | no
`role` | `string` | Type of kubernetes resource to query. Must be one of `node`, `pod`, `service`, `endpoints`, `endpointslice` or `ingress` | | **yes**
`kubeconfig_file` | `string` | Path of kubeconfig file to use for kubernetes connection | | no

The following sub-blocks are supported:

Name | Description | Required
---- | ----------- | --------
[`namespaces`](#namespaces-block) | Information about which kubernetes namespaces to search. | no
[`selectors`](#selectors-block) | Selectors to limit objects selected | no
[`http_client_config`](#http_client_config-block) | HTTP client configuration for kubernetes requests | no

### `namespaces` block

The `namespaces` block limits the namespaces to discover resources in. If omitted, all namespaces are searched.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`own_namespace` | `bool`   | Include the namespace the agent is running in? | | no
`names` | `[]string` | List of namespaces | | no

### `selectors` block

The `selectors` block contains optional label and field selectors to limit the discovery process to a subset of resources.
See [https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/](https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/)
and [https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) to learn more about the possible
filters that can be used.

Note: When making decision about using field/label selector make sure that this
is the best approach - it will prevent Prometheus from reusing single list/watch
for all scrape configs. This might result in a bigger load on the Kubernetes API,
because per each selector combination there will be additional LIST/WATCH. On the other hand,
if you just want to monitor small subset of pods in large cluster it's recommended to use selectors.
The decision, to use selectors or not depends on the particular situation.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`role` | `string`   | role of the selector | | yes
`label`| `string`   | label selector string | | no
`field` | `string`   | field selector string | | no

### `http_client_config` block

The `http_client_config` block configures settings used to connect to the Kubernetes API server.
Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`bearer_token`             | `secret`   | Use to set up the Bearer Token. | | no
`bearer_token_file`        | `string`   | Use to set up the Bearer Token file. | | no
`proxy_url`                | `string`   | Use to set up a proxy URL. | | no
`follow_redirects`         | `bool`     | Whether the scraper should follow redirects. | `true` | no
`enable_http_2`            | `bool`     | Whether the scraper should use HTTP2. | `true` | no

The following sub-blocks are supported for `http_client_config`:

Name | Description | Required
---- | ----------- | --------
[`basic_auth`](#basic_auth-block) | Configure basic_auth for authenticating against Kubernetes | no
[`authorization`](#authorization-block) | Configure generic authorization against Kubernetes | no
[`oauth2`](#oauth2-block) | Configure OAuth2 for authenticating against Kubernetes | no
[`tls_config`](#tls_config-block) | Configure TLS settings for connecting to Kubernetes | no

#### `basic_auth` block

Name          | Type     | Description                                     | Default | Required
------------- | -------- | ----------------------------------------------- | ------- | -------
`username`      | `string`   | Setup of Basic HTTP authentication credentials. |         | no
`password`      | `secret`   | Setup of Basic HTTP authentication credentials. |         | no
`password_file` | `string`   | Setup of Basic HTTP authentication credentials. |         | no

#### `authorization` block

Name                  | Type       | Description                              | Default | Required
--------------------- | ---------- | ---------------------------------------- | ------- | --------
`type`                | `string`   | Setup of HTTP Authorization credentials. |         | no
`credential`          | `secret`   | Setup of HTTP Authorization credentials. |         | no
`credentials_file`    | `string`   | Setup of HTTP Authorization credentials. |         | no

#### `oauth2` block

Name                 | Type                 | Description                              | Default | Required
-------------------- | -------------------- | ---------------------------------------- | ------- | --------
`client_id`          | `string`             | Setup of the OAuth2 client.              |         | no
`client_secret`      | `secret`             | Setup of the OAuth2 client.              |         | no
`client_secret_file` | `string`             | Setup of the OAuth2 client.              |         | no
`scopes`             | `list(string)`       | Setup of the OAuth2 client.              |         | no
`token_url`          | `string`             | Setup of the OAuth2 client.              |         | no
`endpoint_params`    | `map(string)`        | Setup of the OAuth2 client.              |         | no
`proxy_url`          | `string`             | Setup of the OAuth2 client.              |         | no

The `oauth2` block may also contain its own separate `tls_config` sub-block.

#### `tls_config` block

Name                              | Type       | Description                                | Default | Required
--------------------------------- | ---------- | ------------------------------------------ | ------- | --------
`tls_config_ca_file`              | `string`   | Configuration options for TLS connections. |         | no
`tls_config_cert_file`            | `string`   | Configuration options for TLS connections. |         | no
`tls_config_key_file`             | `string`   | Configuration options for TLS connections. |         | no
`tls_config_server_name`          | `string`   | Configuration options for TLS connections. |         | no
`tls_config_insecure_skip_verify` | `bool`     | Configuration options for TLS connections. |         | no

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
targets | list(map(string)) | The set of targets discovered from the Kubernetes API

## Component health

`discovery.kubernetes` will only be reported as unhealthy when given an invalid
configuration. In those cases, exported fields will be kept at their last
healthy values.

## Debug information

`discovery.kubernetes` does not expose any component-specific debug information.

### Debug metrics

`discovery.kubernetes` does not expose any component-specific debug metrics.
