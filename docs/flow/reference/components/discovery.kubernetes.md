---
aliases:
- /docs/agent/latest/flow/reference/components/discovery.kubernetes
title: discovery.kubernetes
---

# discovery.kubernetes

`discovery.kubernetes` allows you to find scrape targets from kubernetes resources.

## Example

```river
discovery.relabel "keep_backend_only" {
  targets = [
    { "__meta_foo" = "foo", "__address__" = "localhost", "instance" = "one",   "app" = "backend"  },
    { "__meta_bar" = "bar", "__address__" = "localhost", "instance" = "two",   "app" = "database" },
    { "__meta_baz" = "baz", "__address__" = "localhost", "instance" = "three", "app" = "frontend" },
  ]

  relabel_config {
    source_labels = ["__address__", "instance"]
    separator     = "/"
    target_label  = "destination"
    action        = "replace"
  }

  relabel_config {
    source_labels = ["app"]
    action        = "keep"
    regex         = "backend"
  }
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
api_server | string | URL of kubernetes API server | | no
role | string | Type of kubernetes resource to query. Must be one of `node`, `pod`, `service`, `endpoints`, `endpointslice` or `ingress` | | **yes**
kubeconfig_file | string | Path of kubeconfig file to use for kubernetes connection | | no

The following subblocks are supported:

Name | Description | Required
---- | ----------- | --------
[namespaces](#namespaces-block) | Information about which kubernetes namespaces to search | no
[selectors](#selectors-block) | Selectors to limit objects selected | no
[`http_client_config`](#http_client_config-block) | HTTP client configuration for kubernetes requests | no

### `http_client_config` block

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
[`basic_auth`](#basic_auth-block) | Configure basic_auth for authenticating against targets | no
[`authorization`](#authorization-block) | Configure generic authorization against targets | no
[`oauth2`](#oauth2-block) | Configure OAuth2 for authenticating against targets | no
[`tls_config`](#tls_config-block) | Configure TLS settings for connecting to targets | no

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
`tls_config_insecure_skip_verify` | `bool`     | Configuration options for TLS connections. |         | no d

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
trgets | list(map(string)) | The set of targets discovered from the kubernetes api

## Component health

`discovery.kubernetes` will only be reported as unhealthy when given an invalid
configuration. In those cases, exported fields will be kept at their last
healthy values.

## Debug information

`discovery.kubernetes` does not expose any component-specific debug information.

### Debug metrics

`discovery.kubernetes` does not expose any component-specific debug metrics.
