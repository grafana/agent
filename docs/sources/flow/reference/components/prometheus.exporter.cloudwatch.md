---
# NOTE(thepalbi, from rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.cloudwatch
---

# prometheus.exporter.cloudwatch

The `prometheus.exporter.cloudwatch` component embeds

[`yet-another-cloudwatch-exporter`](https://github.com/nerdswords/yet-another-cloudwatch-exporter). `` lets you collect [CloudWatch metrics](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/WhatIsCloudWatch.html), translate them to prometheus-compatible format and remote write.

This component lets you scrape CloudWatch metrics in a set of configurations that we will call *jobs*. There are
two kind of jobs: [discovery][] and [static][].

[discovery]: #discovery-block
[static]: #static-block

## Authentication

The agent must be running in an environment with access to AWS. The exporter uses the [AWS SDK for Go](https://aws.github.io/aws-sdk-go-v2/docs/getting-started/) and
provides authentication via [AWS's default credential chain](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials). Regardless of the method used to acquire the credentials,
some permissions are needed for the exporter to work.

```
"tag:GetResources",
"cloudwatch:GetMetricData",
"cloudwatch:GetMetricStatistics",
"cloudwatch:ListMetrics"
```

The following IAM permissions are required for the [Transit Gateway](https://aws.amazon.com/transit-gateway/) attachment (tgwa) metrics to work.

```
"ec2:DescribeTags",
"ec2:DescribeInstances",
"ec2:DescribeRegions",
"ec2:DescribeTransitGateway*"
```

The following IAM permission is required to discover tagged [API Gateway](https://aws.amazon.com/es/api-gateway/) REST APIs:

```
"apigateway:GET"
```

The following IAM permissions are required to discover tagged [Database Migration Service](https://aws.amazon.com/dms/) (DMS) replication instances and tasks:

```
"dms:DescribeReplicationInstances",
"dms:DescribeReplicationTasks"
```

To use all of the integration features, use the following AWS IAM Policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Stmt1674249227793",
      "Action": [
        "tag:GetResources",
        "cloudwatch:GetMetricData",
        "cloudwatch:GetMetricStatistics",
        "cloudwatch:ListMetrics",
        "ec2:DescribeTags",
        "ec2:DescribeInstances",
        "ec2:DescribeRegions",
        "ec2:DescribeTransitGateway*",
        "apigateway:GET",
        "dms:DescribeReplicationInstances",
        "dms:DescribeReplicationTasks"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
```

## Usage

```river
prometheus.exporter.cloudwatch "queues" {
    sts_region = "us-east-2"
    discovery {
        type = "sqs"
        regions = ["us-east-2"]
        search_tags = {
            "scrape" = "true",
        }
        metric {
            name = "NumberOfMessagesSent"
            statistics = ["Sum", "Average"]
            period = "1m"
        }
        metric {
            name = "NumberOfMessagesReceived"
            statistics = ["Sum", "Average"]
            period = "1m"
        }
    }
}
```

## Arguments

The following arguments can be used to configure the exporter's behavior.
Omitted fields take their default values.

| Name                      | Type                | Description                                                                                                                                                                                                                            | Default | Required |
| ------------------------- | ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | -------- |
| `sts_region`              | `string`            | AWS region to use when calling [STS][] for retrieving account information.                                                                                                                                                             |         | yes      |
| `fips_disabled`           | `bool`              | Disable use of FIPS endpoints. Set 'true' when running outside of USA regions.                                                                                                                                                         | `true`  | no       |
| `debug`                   | `bool`              | Enable debug logging on CloudWatch exporter internals.                                                                                                                                                                                 | `false` | no       |
| `discovery_exported_tags` | `map(list(string))` | List of tags (value) per service (key) to export in all metrics. For example defining the `["name", "type"]` under `"AWS/EC2"` will export the name and type tags and its values as labels in all metrics. Affects all discovery jobs. | `{}`    | no       |

[STS]: https://docs.aws.amazon.com/STS/latest/APIReference/welcome.html

## Blocks

The following blocks are supported inside the definition of
`prometheus.exporter.cloudwatch` to configure collector-specific options:

| Hierarchy          | Name          | Description                                                                                                                             | Required |
| ------------------ | ------------- | --------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| discovery          | [discovery][] | Configures a discovery job. Multiple jobs can be configured.                                                                            | no*      |
| discovery > role   | [role][]      | Configures the IAM roles the job should assume to scrape metrics. Defaults to the role configured in the environment the agent runs on. | no       |
| discovery > metric | [metric][]    | Configured the list of metrics the job should scrape. Multiple can be defined inside one job. target.                                   | yes      |
| static             | [static][]    | Configures a static job. Multiple jobs can be configured.                                                                               | no*      |
| static > role      | [role][]      | Configures the IAM roles the job should assume to scrape metrics. Defaults to the role configured in the environment the agent runs on. | no       |
| static > metric    | [metric][]    | Configured the list of metrics the job should scrape. Multiple can be defined inside one job. target.                                   | yes      |

Note that both the `static` and `discovery` blocks are marked with required `no*`. The caveat is that at least one job needs to be configured.

[discovery]: #discovery-block
[static]: #static-block
[metric]: #metric-block
[role]: #role-block

## discovery block

The `discovery` block configures the allows the component to scrape CloudWatch metrics just with the AWS service, and a list of metrics under that service/namespace.
The agent will find AWS resources in the specified service for which to scrape these metrics, label them appropriately, and
export them to Prometheus. For example, if we wanted to scrape CPU utilization and network traffic metrics, from all AWS
EC2 instances:

```river
prometheus.exporter.cloudwatch "discover-instances" {
    sts_region = "us-east-2"
    discovery {
        type = "AWS/EC2"
        regions = ["us-east-2"]
        metric {
            name = "CPUUtilization"
            statistics = ["Average"]
            period = "5m"
        }
        metric {
            name = "NetworkPacketsIn"
            statistics = ["Average"]
            period = "5m"
        }
    }
}
```

The `discovery` block can be configured one or multiple times to scrape metrics from different services, or with different `search_tags`.

| Name          | Type           | Description                                                                                                                                                  | Default | Required |
| ------------- | -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------- | -------- |
| `regions`     | `list(string)` | List of AWS regions.                                                                                                                                         |         | yes      |
| `type`        | `string`       | Cloudwatch service alias (`"alb"`, `"ec2"`, etc) or namespace name (`"AWS/EC2"`, `"AWS/S3"`, etc). See [supported-services][] for a complete list.           |         | yes      |
| `custom_tags` | `map(string)`  | Custom tags to be added as a list of key / value pairs. When exported to Prometheus format, the label name follows the following format: `custom_tag_{key}`. | `{}`    | no       |
| `search_tags` | `map(string)`  | List of key / value pairs to use for tag filtering (all must match). Value can be a regex.                                                                   | `{}`    | no       |

[supported-services]: #supported-services-in-discovery-jobs

## static block

The `static` block configures the component to scrape an specific set of CloudWatch metrics. For that, metrics needs to be fully qualified, specifying the following:

1. `namespace`: For example `AWS/EC2`, `AWS/EBS`, `CoolApp` if it were a custom metric, etc.
2. `dimensions`: CloudWatch identifies a metrics by a set of dimensions, which are essentially label / value pairs. For example, all `AWS/EC2` metrics are identified by the `InstanceId` dimension, and the identifier itself.
3. `metric`: Metric name and statistics.

For example, if one wants to scrape the same metrics in the discovery example, but for a specific AWS EC2 instance:

```river
prometheus.exporter.cloudwatch "static-instances" {
    sts_region = "us-east-2"
    static "instances" {
        regions = ["us-east-2"]
        namespace = "AWS/EC2"
        dimensions = {
            "InstanceId" = "i01u29u12ue1u2c",
        }
        metric {
            name = "CPUUsage"
            statistics = ["Sum", "Average"]
            period = "1m"
        }
    }
}
```

The `static` block can be configured one or multiple times to scrape metrics with different sets of `dimensions`.

| Name          | Type           | Description                                                                                                                                                  | Default | Required |
| ------------- | -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------- | -------- |
| `name`        | `string`       | Identifier of the static scraping job. When exported to Prometheus format corresponds to the `name` label.                                                   |         | yes      |
| `regions`     | `list(string)` | List of AWS regions.                                                                                                                                         |         | yes      |
| `namespace`   | `string`       | CloudWatch metric namespace.                                                                                                                                 |         | yes      |
| `dimensions`  | `map(string)`  | CloudWatch metric dimensions as a list of name / value pairs. Must uniquely define all metrics in this job.                                                  |         | yes      |
| `custom_tags` | `map(string)`  | Custom tags to be added as a list of key / value pairs. When exported to Prometheus format, the label name follows the following format: `custom_tag_{key}`. | `{}`    | no       |

All dimensions need to be specified when scraping single metrics like the example above. For example `AWS/Logs` metrics
require `Resource`, `Service`, `Class`, and `Type` dimensions to be specified. Same applies to CloudWatch custom metrics,
all dimensions attached to a metric when saved in CloudWatch are required.

## metric block

Represents an AWS Metrics to scrape. To see available metrics, AWS does not keep a documentation page with all available metrics.
Follow [this guide](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/viewing_metrics_with_cloudwatch.html) on how to explore metrics, to easily pick the ones you need.

| Name         | Type           | Description                                                      | Default | Required |
| ------------ | -------------- | ---------------------------------------------------------------- | ------- | -------- |
| `name`       | `string`       | Metric name.                                                     |         | yes      |
| `statistics` | `list(string)` | List of statistics to scrape. Ex: `"Minimum"`, `"Maximum"`, etc. |         | yes      |
| `period`     | `duration`     | See [period][] section below.                                    |         | yes      |

[period]: #period

## role block

Represents an [AWS IAM Role](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html). If omitted, the AWS role that corresponds to the credentials configured in the environment will be used.

Multiple roles can be useful when scraping metrics from different AWS accounts with a single pair of credentials. In this case, a different role
is configured for the agent to assume prior to calling AWS APIs, therefore, the credentials configured in the system need
permission to assume the target role. See [this documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_permissions-to-switch.html) on how to configure this.


| Name          | Type     | Description                                                           | Default | Required |
| ------------- | -------- | --------------------------------------------------------------------- | ------- | -------- |
| `role_arn`    | `string` | AWS IAM Role ARN the exporter should assume to perform AWS API calls. |         | yes      |
| `external_id` | `string` | External ID used when calling STS AssumeRole API. See [details][].    | `""`    | no       |

[details]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html

## period

Period controls how far back in time CloudWatch metrics are considered, during each agent scrape. We can split how these
settings affects the produced values in two different scenarios.

If all metrics within a job (discovery or static) have the same `Period` value configured, CloudWatch APIs will be requested
for metrics from the scrape time, to `Periods` seconds in the past. The values of these are exported to Prometheus.

![](https://grafana.com/media/docs/agent/cloudwatch-single-period-time-model.png)

On the other hand, if metrics with different `Periods` are configured under an individual job, this works differently.
First, two variables are calculated aggregating all periods: `length`, taking the maximum value of all periods, and
the new `period` value, taking the minimum of all periods. Then, CloudWatch APIs will be requested for metrics from
`now - length` to `now`, aggregating each in samples for `period` seconds. For each metrics, the most recent sample
is exported to CloudWatch.

![](https://grafana.com/media/docs/agent/cloudwatch-multiple-period-time-model.png)

## Exported fields

The following fields are exported and can be referenced by other components.

| Name      | Type                | Description                                                               |
| --------- | ------------------- | ------------------------------------------------------------------------- |
| `targets` | `list(map(string))` | The targets that can be used to collect the scraped `cloudwatch` metrics. |

For example, `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

The exported targets will use the configured [in-memory traffic][] address
specified by the [run command][].

[in-memory traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[run command]: {{< relref "../cli/run.md" >}}

## Component health

`prometheus.exporter.cloudwatch` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.cloudwatch` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.cloudwatch` does not expose any component-specific
debug metrics.

## Example

See the examples described under each [discovery][] and [static] sections.

[discovery]: #discovery-block
[static]: #static-block

## Supported services in discovery jobs

The following is a list of AWS services that are supported in `cloudwatch_exporter` discovery jobs. When configuring a
discovery job, the `type` field of each `discovery_job` must match either the desired job namespace or alias.

- Namespace: `AWS/CertificateManager` or Alias: `acm`
- Namespace: `AmazonMWAA` or Alias: `airflow`
- Namespace: `AWS/ApplicationELB` or Alias: `alb`
- Namespace: `AWS/AppStream` or Alias: `appstream`
- Namespace: `AWS/Backup` or Alias: `backup`
- Namespace: `AWS/ApiGateway` or Alias: `apigateway`
- Namespace: `AWS/AmazonMQ` or Alias: `mq`
- Namespace: `AWS/AppSync` or Alias: `appsync`
- Namespace: `AWS/Athena` or Alias: `athena`
- Namespace: `AWS/AutoScaling` or Alias: `asg`
- Namespace: `AWS/ElasticBeanstalk` or Alias: `beanstalk`
- Namespace: `AWS/Billing` or Alias: `billing`
- Namespace: `AWS/Cassandra` or Alias: `cassandra`
- Namespace: `AWS/CloudFront` or Alias: `cloudfront`
- Namespace: `AWS/Cognito` or Alias: `cognito-idp`
- Namespace: `AWS/DMS` or Alias: `dms`
- Namespace: `AWS/DDoSProtection` or Alias: `shield`
- Namespace: `AWS/DocDB` or Alias: `docdb`
- Namespace: `AWS/DX` or Alias: `dx`
- Namespace: `AWS/DynamoDB` or Alias: `dynamodb`
- Namespace: `AWS/EBS` or Alias: `ebs`
- Namespace: `AWS/ElastiCache` or Alias: `ec`
- Namespace: `AWS/EC2` or Alias: `ec2`
- Namespace: `AWS/EC2Spot` or Alias: `ec2Spot`
- Namespace: `AWS/ECS` or Alias: `ecs-svc`
- Namespace: `ECS/ContainerInsights` or Alias: `ecs-containerinsights`
- Namespace: `AWS/EFS` or Alias: `efs`
- Namespace: `AWS/ELB` or Alias: `elb`
- Namespace: `AWS/ElasticMapReduce` or Alias: `emr`
- Namespace: `AWS/EMRServerless` or Alias: `emr-serverless`
- Namespace: `AWS/ES` or Alias: `es`
- Namespace: `AWS/Firehose` or Alias: `firehose`
- Namespace: `AWS/FSx` or Alias: `fsx`
- Namespace: `AWS/GameLift` or Alias: `gamelift`
- Namespace: `AWS/GlobalAccelerator` or Alias: `ga`
- Namespace: `Glue` or Alias: `glue`
- Namespace: `AWS/IoT` or Alias: `iot`
- Namespace: `AWS/Kafka` or Alias: `kafka`
- Namespace: `AWS/KafkaConnect` or Alias: `kafkaconnect`
- Namespace: `AWS/Kinesis` or Alias: `kinesis`
- Namespace: `AWS/KinesisAnalytics` or Alias: `kinesis-analytics`
- Namespace: `AWS/Lambda` or Alias: `lambda`
- Namespace: `AWS/MediaTailor` or Alias: `mediatailor`
- Namespace: `AWS/Neptune` or Alias: `neptune`
- Namespace: `AWS/NetworkFirewall` or Alias: `nfw`
- Namespace: `AWS/NATGateway` or Alias: `ngw`
- Namespace: `AWS/NetworkELB` or Alias: `nlb`
- Namespace: `AWS/PrivateLinkEndpoints` or Alias: `vpc-endpoint`
- Namespace: `AWS/PrivateLinkServices` or Alias: `vpc-endpoint-service`
- Namespace: `AWS/Prometheus` or Alias: `amp`
- Namespace: `AWS/RDS` or Alias: `rds`
- Namespace: `AWS/Redshift` or Alias: `redshift`
- Namespace: `AWS/Route53Resolver` or Alias: `route53-resolver`
- Namespace: `AWS/Route53` or Alias: `route53`
- Namespace: `AWS/S3` or Alias: `s3`
- Namespace: `AWS/SES` or Alias: `ses`
- Namespace: `AWS/States` or Alias: `sfn`
- Namespace: `AWS/SNS` or Alias: `sns`
- Namespace: `AWS/SQS` or Alias: `sqs`
- Namespace: `AWS/StorageGateway` or Alias: `storagegateway`
- Namespace: `AWS/TransitGateway` or Alias: `tgw`
- Namespace: `AWS/VPN` or Alias: `vpn`
- Namespace: `AWS/WAFV2` or Alias: `wafv2`
- Namespace: `AWS/WorkSpaces` or Alias: `workspaces`
