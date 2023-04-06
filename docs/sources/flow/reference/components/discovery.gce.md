---
title: discovery.gce
---

# discovery.gce

`discovery.gce` allows retrieving scrape targets from [GCP GCE](https://cloud.google.com/compute) instances. The private IP address is used by default, but may be changed to the public IP address with relabeling.


## Usage

```river
discovery.gce "LABEL" {
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`project` | `string` | The GCP Project.| | yes
`zone` | `string` | The zone of the scrape targets. | | yes
`filter` | `string` | Filter can be used optionally to filter the instance list by other criteria. | | no
`refresh_interval` | `duration` | Refresh interval to re-read the instance list. | `60s`| no
`port` | `int` | The port to scrape metrics from. If using the public IP address, this must instead be specified in the relabeling rule. | `80`| no
`tag_separator` | `string` | The tag separator is used to separate the tags on concatenation. | `","`| no

Syntax of `filter` is described [here](https://cloud.google.com/compute/docs/reference/latest/instances/list).

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`targets` | `list(map(string))` | The set of discovered GCE targets.

Each target includes the following labels:

* `__meta_gce_instance_id`: the numeric id of the instance
* `__meta_gce_instance_name`: the name of the instance
* `__meta_gce_label_<labelname>`: each GCE label of the instance
* `__meta_gce_machine_type`: full or partial URL of the machine type of the instance
* `__meta_gce_metadata_<name>`: each metadata item of the instance
* `__meta_gce_network`: the network URL of the instance
* `__meta_gce_private_ip`: the private IP address of the instance
* `__meta_gce_interface_ipv4_<name>`: IPv4 address of each named interface
* `__meta_gce_project`: the GCP project in which the instance is running
* `__meta_gce_public_ip`: the public IP address of the instance, if present
* `__meta_gce_subnetwork`: the subnetwork URL of the instance
* `__meta_gce_tags`: comma separated list of instance tags
* `__meta_gce_zone`: the GCE zone URL in which the instance is running


## Component health

`discovery.gce` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.gce` does not expose any component-specific debug information.

### Debug metrics

`discovery.gce` does not expose any component-specific debug metrics.

## Examples

```river
discovery.gce "gce" {
  project = "agent"
  zone = "us-east1-a"
}
```
