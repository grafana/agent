---
aliases:
- /docs/agent/latest/configuration/integrations/cloudwatch-exporter-config/
  title: cloudwatch_exporter_config
---

# cloudwatch_exporter_config

The `cloudwatch_exporter_config` block configures the `cloudwatch_exporter` integration,
which is an embedded version of
[`YACE`](https://github.com/nerdswords/yet-another-cloudwatch-exporter/). This allows the collection of [AWS CloudWatch](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/WhatIsCloudWatch.html) metrics.

Configuration reference:

```yaml
  autoscrape:
    # Enables autoscrape of integrations.
    [enable: <boolean> | default = true]

    # Specifies the metrics instance name to send metrics to. Instance
    # names are located at metrics.configs[].name from the top-level config.
    # The instance must exist.
    #
    # As it is common to use the name "default" for your primary instance,
    # we assume the same here.
    [metrics_instance: <string> | default = "default"]

    # Autoscrape interval and timeout. Defaults are inherited from the global
    # section of the top-level metrics config.
    [scrape_interval: <duration> | default = <metrics.global.scrape_interval>]
    [scrape_timeout: <duration> | default = <metrics.global.scrape_timeout>]

  # Integration instance name. 
  # The default value for this integration is "cloudwatch_exporter".
  [instance: <string> | default = "cloudwatch_exporter" ]
  
  # AWS region to use when calling STS (https://docs.aws.amazon.com/STS/latest/APIReference/welcome.html) for retrieving
  # account information.
  sts_region: <string>
  
  discovery:
    exported_tags:
      { <string>: [<string>] }
    jobs: [<discovery_job>]

  static: [<static_job>]
```

# discovery_job

```yaml
  regions: [<string>]
  roles: [<aws_role>]
  custom_tags: [<aws_tag>]
  type: <string>
  metrics: [<metric>] 
```

# static_job

```yaml
  regions: [<string>]
  roles: [<aws_role>]
  custom_tags: [<aws_tag>]
  name: <string>
  namespace: <string>
  dimensions:
  metrics: [<metric>] 
```

# aws_role

```yaml
  role_arn: <string>
  external_id: <string>
```

# aws_dimension
```yaml
  name: <string>
  value: <string>
```

# aws_tag
```yaml
  key: <string>
  value: <string>
```

# metric
```yaml
  name: <string>
  statistics: [<string>]
  period: <duration>
```

```yaml
  role_arn: <string>
  external_id: <string>
```

# Quick configuration example

```yaml
integrations:
  snowflake_configs:
    - account_name: XXXXXXX-YYYYYYY
      username: snowflake-user
      password: snowflake-pass
      warehouse: SNOWFLAKE_WAREHOUSE
      role: ACCOUNTADMIN
      autoscrape:
        enable: true
        metrics_instance: default
        scrape_interval: 30m

metrics:
  wal_directory: /tmp/grafana-agent-wal
server:
  log_level: debug
```

# Supported services in discovery jobs

The following is a list of AWS services that are supported in `cloudwatch_exporter` discovery jobs. When configuring a
discovery job, the `type` field of each `dicsovery_job` must match either one of the desired job namespace or alias.

- Namespace: "AWS/CertificateManager" or Alias: "acm"
- Namespace: "AmazonMWAA" or Alias: "airflow"
- Namespace: "AWS/ApplicationELB" or Alias: "alb"
- Namespace: "AWS/AppStream" or Alias: "appstream"
- Namespace: "AWS/Backup" or Alias: "backup"
- Namespace: "AWS/ApiGateway" or Alias: "apigateway"
- Namespace: "AWS/AmazonMQ" or Alias: "mq"
- Namespace: "AWS/AppSync" or Alias: "appsync"
- Namespace: "AWS/Athena" or Alias: "athena"
- Namespace: "AWS/AutoScaling" or Alias: "asg"
- Namespace: "AWS/ElasticBeanstalk" or Alias: "beanstalk"
- Namespace: "AWS/Billing" or Alias: "billing"
- Namespace: "AWS/Cassandra" or Alias: "cassandra"
- Namespace: "AWS/CloudFront" or Alias: "cloudfront"
- Namespace: "AWS/Cognito" or Alias: "cognito-idp"
- Namespace: "AWS/DMS" or Alias: "dms"
- Namespace: "AWS/DDoSProtection" or Alias: "shield"
- Namespace: "AWS/DocDB" or Alias: "docdb"
- Namespace: "AWS/DX" or Alias: "dx"
- Namespace: "AWS/DynamoDB" or Alias: "dynamodb"
- Namespace: "AWS/EBS" or Alias: "ebs"
- Namespace: "AWS/ElastiCache" or Alias: "ec"
- Namespace: "AWS/EC2" or Alias: "ec2"
- Namespace: "AWS/EC2Spot" or Alias: "ec2Spot"
- Namespace: "AWS/ECS" or Alias: "ecs-svc"
- Namespace: "ECS/ContainerInsights" or Alias: "ecs-containerinsights"
- Namespace: "AWS/EFS" or Alias: "efs"
- Namespace: "AWS/ELB" or Alias: "elb"
- Namespace: "AWS/ElasticMapReduce" or Alias: "emr"
- Namespace: "AWS/EMRServerless" or Alias: "emr-serverless"
- Namespace: "AWS/ES" or Alias: "es"
- Namespace: "AWS/Firehose" or Alias: "firehose"
- Namespace: "AWS/FSx" or Alias: "fsx"
- Namespace: "AWS/GameLift" or Alias: "gamelift"
- Namespace: "AWS/GlobalAccelerator" or Alias: "ga"
- Namespace: "Glue" or Alias: "glue"
- Namespace: "AWS/IoT" or Alias: "iot"
- Namespace: "AWS/Kafka" or Alias: "kafka"
- Namespace: "AWS/KafkaConnect" or Alias: "kafkaconnect"
- Namespace: "AWS/Kinesis" or Alias: "kinesis"
- Namespace: "AWS/KinesisAnalytics" or Alias: "kinesis-analytics"
- Namespace: "AWS/Lambda" or Alias: "lambda"
- Namespace: "AWS/MediaTailor" or Alias: "mediatailor"
- Namespace: "AWS/Neptune" or Alias: "neptune"
- Namespace: "AWS/NetworkFirewall" or Alias: "nfw"
- Namespace: "AWS/NATGateway" or Alias: "ngw"
- Namespace: "AWS/NetworkELB" or Alias: "nlb"
- Namespace: "AWS/PrivateLinkEndpoints" or Alias: "vpc-endpoint"
- Namespace: "AWS/PrivateLinkServices" or Alias: "vpc-endpoint-service"
- Namespace: "AWS/Prometheus" or Alias: "amp"
- Namespace: "AWS/RDS" or Alias: "rds"
- Namespace: "AWS/Redshift" or Alias: "redshift"
- Namespace: "AWS/Route53Resolver" or Alias: "route53-resolver"
- Namespace: "AWS/Route53" or Alias: "route53"
- Namespace: "AWS/S3" or Alias: "s3"
- Namespace: "AWS/SES" or Alias: "ses"
- Namespace: "AWS/States" or Alias: "sfn"
- Namespace: "AWS/SNS" or Alias: "sns"
- Namespace: "AWS/SQS" or Alias: "sqs"
- Namespace: "AWS/StorageGateway" or Alias: "storagegateway"
- Namespace: "AWS/TransitGateway" or Alias: "tgw"
- Namespace: "AWS/VPN" or Alias: "vpn"
- Namespace: "AWS/WAFV2" or Alias: "wafv2"
- Namespace: "AWS/WorkSpaces" or Alias: "workspaces"
