---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/discovery.puppetdb/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.puppetdb/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/discovery.puppetdb/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/discovery.puppetdb/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.puppetdb/
description: Learn about discovery.puppetdb
title: discovery.puppetdb
---

# discovery.puppetdb

`discovery.puppetdb` allows you to retrieve scrape targets from [PuppetDB](https://www.puppet.com/docs/puppetdb/7/overview.html) resources.

This SD discovers resources and creates a target for each resource returned by the API.

The resource address is the `certname` of the resource, and can be changed during relabeling.

The queries for this component are expected to be valid [PQL (Puppet Query Language)](https://puppet.com/docs/puppetdb/latest/api/query/v4/pql.html).

## Usage

```river
discovery.puppetdb "LABEL" {
  url = PUPPET_SERVER
}
```

## Arguments

The following arguments are supported:

Name                 | Type       | Description                                                      | Default | Required
---------------------|------------|------------------------------------------------------------------|---------|---------
`query`              | `string`   | Puppet Query Language (PQL) query. Only resources are supported. |         | yes
`url`                | `string`   | The URL of the PuppetDB root query endpoint.                     |         | yes
`bearer_token_file`  | `string`   | File containing a bearer token to authenticate with.             |         | no
`bearer_token`       | `secret`   | Bearer token to authenticate with.                               |         | no
`enable_http2`       | `bool`     | Whether HTTP2 is supported for requests.                         | `true`  | no
`follow_redirects `  | `bool`     | Whether redirects returned by the server should be followed.     | `true`  | no
`include_parameters` | `bool`     | Whether to include the parameters as meta labels. Due to the differences between parameter types and Prometheus labels, some parameters might not be rendered. The format of the parameters might also change in future releases. Make sure that you don't have secrets exposed as parameters if you enable this. | `false` | no
`port`               | `int`      | The port to scrape metrics from..                                | `80`    | no
`proxy_url`          | `string`   | HTTP proxy to proxy requests through.                            |         | no
`refresh_interval`   | `duration` | Frequency to refresh targets.                                    | `"30s"` | no

 You can provide one of the following arguments for authentication:

- [`authorization` block][authorization].
- [`basic_auth` block][basic_auth].
- [`bearer_token_file` argument](#arguments).
- [`bearer_token` argument](#arguments).
- [`oauth2` block][oauth2].

[arguments]: #arguments

## Blocks

The following blocks are supported inside the definition of `discovery.puppetdb`:

Hierarchy           | Block             | Description                                              | Required
--------------------|-------------------|----------------------------------------------------------|---------
authorization       | [authorization][] | Configure generic authorization to the endpoint.         | no
basic_auth          | [basic_auth][]    | Configure basic_auth for authenticating to the endpoint. | no
oauth2              | [oauth2][]        | Configure OAuth2 for authenticating to the endpoint.     | no
oauth2 > tls_config | [tls_config][]    | Configure TLS settings for connecting to the endpoint.   | no

The `>` symbol indicates deeper levels of nesting.
For example, `oauth2 > tls_config` refers to a `tls_config` block defined inside an `oauth2` block.

[basic_auth]: #basic_auth-block
[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block

### authorization

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" version="<AGENT_VERSION>" >}}

### basic_auth

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" version="<AGENT_VERSION>" >}}

### oauth2

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" version="<AGENT_VERSION>" >}}

### oauth2 > tls_config

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" version="<AGENT_VERSION>" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name      | Type                | Description
----------|---------------------|---------------------------------------------
`targets` | `list(map(string))` | The set of targets discovered from PuppetDB.

Each target includes the following labels:

* `__meta_puppetdb_certname`: The name of the node associated with the resource.
* `__meta_puppetdb_environment`: The environment of the node associated with the resource.
* `__meta_puppetdb_exported`: Whether the resource is exported ("true" or "false").
* `__meta_puppetdb_file`: The manifest file in which the resource was declared.
* `__meta_puppetdb_parameter_<parametername>`: The parameters of the resource.
* `__meta_puppetdb_query`: The Puppet Query Language (PQL) query.
* `__meta_puppetdb_resource`: A SHA-1 hash of the resource’s type, title, and parameters, for identification.
* `__meta_puppetdb_tags`: Comma separated list of resource tags.
* `__meta_puppetdb_title`: The resource title.
* `__meta_puppetdb_type`: The resource type.

## Component health

`discovery.puppetdb` is only reported as unhealthy when given an invalid configuration.
In those cases, exported fields retain their last healthy values.

## Debug information

`discovery.puppetdb` doesn't expose any component-specific debug information.

## Debug metrics

`discovery.puppetdb` doesn't expose any component-specific debug metrics.

## Example

The following example discovers targets from PuppetDB for all the servers that have a specific package defined:

```river
discovery.puppetdb "example" {
	url   = "http://puppetdb.local:8080"
	query = "resources { type = \"Package\" and title = \"node_exporter\" }"
	port  = 9100
}

prometheus.scrape "demo" {
	targets    = discovery.puppetdb.example.targets
	forward_to = [prometheus.remote_write.demo.receiver]
}

prometheus.remote_write "demo" {
	endpoint {
		url = <PROMETHEUS_REMOTE_WRITE_URL>

		basic_auth {
			username = <USERNAME>
			password = <PASSWORD>
		}
	}
}
```

Replace the following:
- _`<PROMETHEUS_REMOTE_WRITE_URL>`_: The URL of the Prometheus remote_write-compatible server to send metrics to.
- _`<USERNAME>`_: The username to use for authentication to the remote_write API.
- _`<PASSWORD>`_: The password to use for authentication to the remote_write API.
