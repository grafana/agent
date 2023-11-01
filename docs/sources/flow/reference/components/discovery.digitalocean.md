---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/discovery.digitalocean/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.digitalocean/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/discovery.digitalocean/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.digitalocean/
title: discovery.digitalocean
description: Learn about discovery.digitalocean
---

# discovery.digitalocean

`discovery.digitalocean` discovers [DigitalOcean][] Droplets and exposes them as targets.

[DigitalOcean]: https://www.digitalocean.com/

## Usage

```river
discovery.digitalocean "LABEL" {
    // Use one of:
    // bearer_token      = BEARER_TOKEN
    // bearer_token_file = PATH_TO_BEARER_TOKEN_FILE
}
```

## Arguments

The following arguments are supported:

Name                | Type       | Description                                                                      | Default | Required
------------------- | ---------- | ---------------------------------------------------------------------- | ------- | --------
`port`              | `number`   | Port to be appended to the `__address__` label for each Droplet.       | `80`    | no
`refresh_interval`  | `duration` | Frequency to refresh list of Droplets.                                 | `"1m"`  | no
`bearer_token`      | `secret`   | Bearer token to authenticate with.                                     |         | no
`bearer_token_file` | `string`   | File containing a bearer token to authenticate with.                   |         | no
`proxy_url`         | `string`   | HTTP proxy to proxy requests through.                                  |         | no
`follow_redirects`  | `bool`     | Whether redirects returned by the server should be followed.           | `true`  | no
`enable_http2`      | `bool`     | Whether HTTP2 is supported for requests.                               | `true`  | no

The DigitalOcean API uses bearer tokens for authentication, see more about it in the [DigitalOcean API documentation](https://docs.digitalocean.com/reference/api/api-reference/#section/Authentication).

Exactly one of the [`bearer_token`](#arguments) and [`bearer_token_file`](#arguments) arguments must be specified to authenticate against DigitalOcean.

[arguments]: #arguments

## Blocks
The `discovery.digitalocean` component does not support any blocks, and is configured 
fully through arguments.


## Exported fields

The following fields are exported and can be referenced by other components:

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The set of targets discovered from the DigitalOcean API.

Each target includes the following labels:

* `__meta_digitalocean_droplet_id`: ID of the Droplet.
* `__meta_digitalocean_droplet_name`: Name of the Droplet.
* `__meta_digitalocean_image`: The image slug (unique text identifier of the image) used to create the Droplet.
* `__meta_digitalocean_image_name`: Name of the image used to create the Droplet.
* `__meta_digitalocean_private_ipv4`: The private IPv4 address of the Droplet.
* `__meta_digitalocean_public_ipv4`: The public IPv4 address of the Droplet.
* `__meta_digitalocean_public_ipv6`: The public IPv6 address of the Droplet.
* `__meta_digitalocean_region`: The region the Droplet is running in.
* `__meta_digitalocean_size`: The size of the Droplet.
* `__meta_digitalocean_status`: The current status of the Droplet.
* `__meta_digitalocean_features`: Optional properties configured for the Droplet, such as IPV6 networking, private networking, or backups.
* `__meta_digitalocean_tags`: The tags assigned to the Droplet.
* `__meta_digitalocean_vpc`: The ID of the VPC where the Droplet is located.

Each discovered Droplet maps to one target.

## Component health

`discovery.digitalocean` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.digitalocean` does not expose any component-specific debug information.

## Debug metrics

`discovery.digitalocean` does not expose any component-specific debug metrics.

## Example

This would result in targets with `__address__` labels like: `192.0.2.1:8080`:
```river
discovery.digitalocean "example" {
  port             = 8080
  refresh_interval = "5m"
  bearer_token     = "my-secret-bearer-token"
}

prometheus.scrape "demo" {
  targets    = discovery.digitalocean.example.targets
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
