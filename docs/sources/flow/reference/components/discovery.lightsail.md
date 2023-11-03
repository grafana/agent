---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/discovery.lightsail/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.lightsail/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/discovery.lightsail/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.lightsail/
title: discovery.lightsail
description: Learn about discovery.lightsail
---

# discovery.lightsail

`discovery.lightsail` allows retrieving scrape targets from Amazon Lightsail instances. The private IP address is used by default, but may be changed to the public IP address with relabeling.

## Usage

```river
discovery.lightsail "LABEL" {
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`endpoint` | `string` | Custom endpoint to be used.| | no
`region` | `string` | The AWS region. If blank, the region from the instance metadata is used. | | no
`access_key` | `string` | The AWS API key ID. If blank, the environment variable `AWS_ACCESS_KEY_ID` is used. | | no
`secret_key` | `string` | The AWS API key secret. If blank, the environment variable `AWS_SECRET_ACCESS_KEY` is used. | | no
`profile` | `string` | Named AWS profile used to connect to the API. | | no
`role_arn` | `string` | AWS Role ARN, an alternative to using AWS API keys. | | no
`refresh_interval` | `string` | Refresh interval to re-read the instance list. | 60s | no
`port` | `int` | The port to scrape metrics from. If using the public IP address, this must instead be specified in the relabeling rule. | 80 | no

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`targets` | `list(map(string))` | The set of discovered Lightsail targets.

Each target includes the following labels:

* `__meta_lightsail_availability_zone`: The availability zone in which the instance is running.
* `__meta_lightsail_blueprint_id`: The Lightsail blueprint ID.
* `__meta_lightsail_bundle_id`: The Lightsail bundle ID.
* `__meta_lightsail_instance_name`: The name of the Lightsail instance.
* `__meta_lightsail_instance_state`: The state of the Lightsail instance.
* `__meta_lightsail_instance_support_code`: The support code of the Lightsail instance.
* `__meta_lightsail_ipv6_addresses`: Comma-separated list of IPv6 addresses assigned to the instance's network interfaces, if present.
* `__meta_lightsail_private_ip`: The private IP address of the instance.
* `__meta_lightsail_public_ip`: The public IP address of the instance, if available.
* `__meta_lightsail_region`: The region of the instance.
* `__meta_lightsail_tag_<tagkey>`: Each tag value of the instance.


## Component health

`discovery.lightsail` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.lightsail` does not expose any component-specific debug information.

## Debug metrics

`discovery.lightsail` does not expose any component-specific debug metrics.

## Example

```river
discovery.lightsail "lightsail" {
  region = "us-east-1"
}

prometheus.scrape "demo" {
  targets    = discovery.lightsail.lightsail.targets
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
