---
title: discovery.ec2
---

# discovery.ec2

`discovery.ec2` lets you retrieve scrape targets from EC2 instances. The private IP address is used by default, but you can change it to the public IP address using relabeling.

The IAM credentials used must have the `ec2:DescribeInstances` permission to discover scrape targets, and may optionally have the `ec2:DescribeAvailabilityZones` permission to make the availability zone ID available as a label.

## Usage

```river
discovery.ec2 "LABEL" {
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
`role_arn` | `string` | AWS Role Amazon Resource Name (ARN), an alternative to using AWS API keys. | | no
`refresh_interval` | `string` | Refresh interval to re-read the instance list. | 60s | no
`port` | `int` | The port to scrape metrics from. If using the public IP address, this must instead be specified in the relabeling rule. | 80 | no

## Blocks

The following blocks are supported inside the definition of
`discovery.ec2`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
filter | [filter][] | Filters discoverable resources. | no

[filter]: #filter-block

### filter block

Filters can be used optionally to filter the instance list by other criteria.
Available filter criteria can be found in the [Amazon EC2 documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html).

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`name` | `string` | Filter name to use. | | yes
`values` | `list(string)` | Values to pass to the filter. | | yes

Refer to [Filter API documentation][filter api] from the aws
documentation for the list of supported filters and their meaning.

[filter api]: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Filter.html

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`targets` | `list(map(string))` | The set of discovered EC2 targets.

Each target includes the following labels:

* `__meta_ec2_ami`: the EC2 Amazon Machine Image
* `__meta_ec2_architecture`: the architecture of the instance
* `__meta_ec2_availability_zone`: the availability zone in which the instance is running
* `__meta_ec2_availability_zone_id`: the availability zone ID in which the instance is running (requires `ec2:DescribeAvailabilityZones`)
* `__meta_ec2_instance_id`: the EC2 instance ID
* `__meta_ec2_instance_lifecycle`: the lifecycle of the EC2 instance, set only for 'spot' or 'scheduled' instances, absent otherwise
* `__meta_ec2_instance_state`: the state of the EC2 instance
* `__meta_ec2_instance_type`: the type of the EC2 instance
* `__meta_ec2_ipv6_addresses`: comma separated list of IPv6 addresses assigned to the instance's network interfaces, if present
* `__meta_ec2_owner_id`: the ID of the AWS account that owns the EC2 instance
* `__meta_ec2_platform`: the Operating System platform, set to 'windows' on Windows servers, absent otherwise
* `__meta_ec2_primary_subnet_id`: the subnet ID of the primary network interface, if available
* `__meta_ec2_private_dns_name`: the private DNS name of the instance, if available
* `__meta_ec2_private_ip`: the private IP address of the instance, if present
* `__meta_ec2_public_dns_name`: the public DNS name of the instance, if available
* `__meta_ec2_public_ip`: the public IP address of the instance, if available
* `__meta_ec2_region`: the region of the instance
* `__meta_ec2_subnet_id`: comma separated list of subnets IDs in which the instance is running, if available
* `__meta_ec2_tag_<tagkey>`: each tag value of the instance
* `__meta_ec2_vpc_id`: the ID of the VPC in which the instance is running, if available

## Component health

`discovery.ec2` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.ec2` does not expose any component-specific debug information.

### Debug metrics

`discovery.ec2` does not expose any component-specific debug metrics.

## Examples

```river
discovery.ec2 "ec2" {
  region = "us-east-1"
}
```
