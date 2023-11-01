---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.linode/
title: discovery.linode
description: Learn about discovery.linode
---

# discovery.linode

`discovery.linode` allows you to retrieve scrape targets from [Linode's](https://www.linode.com/) Linode APIv4.
This service discovery uses the public IPv4 address by default, but that can be changed with relabeling.

## Usage

```river
discovery.linode "LABEL" {
	bearer_token = LINODE_API_TOKEN
}
```

{{% admonition type="note" %}}
The linode APIv4 Token must be created with the scopes: `linodes:read_only`, `ips:read_only`, and `events:read_only`.
{{% /admonition %}}

## Arguments

The following arguments are supported:

Name               | Type           | Description                                                    | Default       | Required
------------------ | -------------- | -------------------------------------------------------------- | ------------- | --------
`refresh_interval` | `duration`     | The time to wait between polling update requests.              | `"60s"`       | no
`port`    | `int`     | Port that metrics are scraped from. | `80`    | no
`tag_separator` | `string`  | The string by which Linode Instance tags are joined into the tag label. | `,` | no

`bearer_token` | `secret` | Bearer token to authenticate with. | | no
`bearer_token_file` | `string` | File containing a bearer token to authenticate with. | | no
`proxy_url` | `string` | HTTP proxy to proxy requests through. | | no
`follow_redirects` | `bool` | Whether redirects returned by the server should be followed. | `true` | no
`enable_http2` | `bool` | Whether HTTP2 is supported for requests. | `true` | no

 You can provide one of the following arguments for authentication:
 - [`bearer_token` argument](#arguments).
 - [`bearer_token_file` argument](#arguments).
 - [`authorization` block][authorization].
 - [`oauth2` block][oauth2].

The following blocks are supported inside the definition of
`discovery.linode`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
authorization | [authorization][] | Configure generic authorization to the endpoint. | no
oauth2 | [oauth2][] | Configure OAuth2 for authenticating to the endpoint. | no
oauth2 > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no

The `>` symbol indicates deeper levels of nesting. For example,
`oauth2 > tls_config` refers to a `tls_config` block defined inside
an `oauth2` block.

[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block

### authorization block

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" version="<AGENT VERSION>" >}}

### oauth2 block

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" version="<AGENT VERSION>" >}}

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" version="<AGENT VERSION>" >}}


## Exported fields

The following fields are exported and can be referenced by other components:

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The set of targets discovered from the Linode API.

The following meta labels are available on targets and can be used by the
discovery.relabel component:

* `__meta_linode_instance_id`: the id of the Linode instance
* `__meta_linode_instance_label`: the label of the Linode instance
* `__meta_linode_image`: the slug of the Linode instance's image
* `__meta_linode_private_ipv4`: the private IPv4 of the Linode instance
* `__meta_linode_public_ipv4`: the public IPv4 of the Linode instance
* `__meta_linode_public_ipv6`: the public IPv6 of the Linode instance
* `__meta_linode_region`: the region of the Linode instance
* `__meta_linode_type`: the type of the Linode instance
* `__meta_linode_status`: the status of the Linode instance
* `__meta_linode_tags`: a list of tags of the Linode instance joined by the tag separator
* `__meta_linode_group`: the display group a Linode instance is a member of
* `__meta_linode_hypervisor`: the virtualization software powering the Linode instance
* `__meta_linode_backups`: the backup service status of the Linode instance
* `__meta_linode_specs_disk_bytes`: the amount of storage space the Linode instance has access to
* `__meta_linode_specs_memory_bytes`: the amount of RAM the Linode instance has access to
* `__meta_linode_specs_vcpus`: the number of VCPUS this Linode has access to
* `__meta_linode_specs_transfer_bytes`: the amount of network transfer the Linode instance is allotted each month
* `__meta_linode_extra_ips`: a list of all extra IPv4 addresses assigned to the Linode instance joined by the tag separator

## Component health

`discovery.linode` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.linode` does not expose any component-specific debug information.

## Debug metrics

`discovery.linode` does not expose any component-specific debug metrics.

## Example

```river
discovery.linode "example" {
    bearer_token = env("LINODE_TOKEN")
    port = 8876
}
prometheus.scrape "demo" {
	targets    = discovery.linode.example.targets
	forward_to = [prometheus.remote_write.demo.receiver]
}
prometheus.remote_write "demo" {
	endpoint {
		url = PROMETHEUS_REMOTE_WRITE_URL
		basic_auth {
			username = USERNAME
			password = PASSWORD
		}
	}
}
```
Replace the following:
  - `PROMETHEUS_REMOTE_WRITE_URL`: The URL of the Prometheus remote_write-compatible server to send metrics to.
  - `USERNAME`: The username to use for authentication to the remote_write API.
  - `PASSWORD`: The password to use for authentication to the remote_write API.

### Using private IP address:

```
discovery.linode "example" {
    bearer_token = env("LINODE_TOKEN")
    port = 8876
}
discovery.relabel "private_ips" {
	targets = discovery.linode.example.targets
	rule {
    	source_labels = ["__meta_linode_private_ipv4"]
    	replacement     = "[$1]:8876"
    	target_label  = "__address__"
  	}
}
prometheus.scrape "demo" {
	targets    = discovery.relabel.private_ips.targets
	forward_to = [prometheus.remote_write.demo.receiver]
}
prometheus.remote_write "demo" {
	endpoint {
		url = PROMETHEUS_REMOTE_WRITE_URL
		basic_auth {
			username = USERNAME
			password = PASSWORD
		}
	}
}
```