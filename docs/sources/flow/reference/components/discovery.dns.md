---
aliases:
- /docs/agent/latest/flow/reference/components/discovery.dns
- /docs/grafana-cloud/agent/flow/reference/components/discovery.dns/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.dns/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/discovery.dns/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.dns/
title: discovery.dns
description: Learn about discovery.dns
---

# discovery.dns

`discovery.dns` discovers scrape targets from DNS records.

## Usage

```river
discovery.dns "LABEL" {
  names = [NAME_1, NAME_2, ...]
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`names` | `list(string)` | DNS names to look up. | | yes
`port` | `number` | Port to use for collecting metrics. Not used for SRV records. | `0` | no
`refresh_interval` | `duration` | How often to query DNS for updates. | `"30s"` | no
`type` | `string` | Type of DNS record to query. Must be one of SRV, A, AAAA, or MX. | `"SRV"` | no

## Exported fields

The following field is exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`targets` | `list(map(string))` | The set of targets discovered from the docker API.

Each target includes the following labels:

* `__meta_dns_name`: Name of the record that produced the discovered target.
* `__meta_dns_srv_record_target`: Target field of the SRV record.
* `__meta_dns_srv_record_port`: Port field of the SRV record.
* `__meta_dns_mx_record_target`: Target field of the MX record.


## Component health

`discovery.dns` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.dns` does not expose any component-specific debug information.

## Debug metrics

`discovery.dns` does not expose any component-specific debug metrics.

## Example

This example discovers targets from an A record.

```river
discovery.dns "dns_lookup" {
  names = ["myservice.example.com", "myotherservice.example.com"]
  type = "A"
  port = 8080
}

prometheus.scrape "demo" {
  targets    = discovery.dns.dns_lookup.targets
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