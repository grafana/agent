---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.processor.resourcedetection/
labels:
  stage: beta
title: otelcol.processor.resourcedetection
description: Learn about otelcol.processor.resourcedetection
---

# otelcol.processor.resourcedetection

{{< docs/shared lookup="flow/stability/beta.md" source="agent" version="<AGENT VERSION>" >}}

`otelcol.processor.resourcedetection` detects resource information and adds resource attributes to logs, 
metrics, and traces telemetry data.

{{% admonition type="note" %}}
`otelcol.processor.resourcedetection` is a wrapper over the upstream
OpenTelemetry Collector Contrib `resourcedetection` processor. If necessary, 
bug reports or feature requests will be redirected to the upstream repository.
{{% /admonition %}}

You can specify multiple `otelcol.processor.resourcedetection` components by giving them
different labels.

## Usage

```river
otelcol.processor.resourcedetection "LABEL" {
  output {
    logs    = [...]
    traces  = [...]
  }
}
```

## Arguments

`otelcol.processor.resourcedetection` supports the following arguments:

Name        | Type           | Description                                                                            | Default | Required
----------- | -------------- | -------------------------------------------------------------------------------------- |-------- | --------
`detectors` | `list(string)` | An ordered list of named detectors which should be ran to detect resource information. |         | yes
`override`  | `bool`         | Configures whether existing resource attributes should be overriden or preserved.      | `true`  | no
`timeout`   | `duration`     | Timeout by which all specified detectors must complete.                                | `"5s"`  | no

## Blocks

The following blocks are supported inside the definition of
`otelcol.processor.resourcedetection`:

Hierarchy                              | Block                                                       | Description                                       | Required
-------------------------------------- | ----------------------------------------------------------- | ------------------------------------------------- | --------
output                                 | [output][]                                                  | Configures where to send received telemetry data. | yes
ec2                                    | [ec2][]                                                     |                                                   | no
ec2 > resource_attributes              | [resource_attributes][ec2-resource_attributes]              |                                                   | no
ecs                                    | [ecs][]                                                     |                                                   | no
ecs > resource_attributes              | [resource_attributes][ecs-resource_attributes]              |                                                   | no
eks                                    | [eks][]                                                     |                                                   | no
eks > resource_attributes              | [resource_attributes][eks-resource_attributes]              |                                                   | no
elasticbeanstalk                       | [elasticbeanstalk][]                                        |                                                   | no
elasticbeanstalk > resource_attributes | [resource_attributes][elasticbeanstalk-resource_attributes] |                                                   | no
lambda                                 | [lambda][]                                                  |                                                   | no
lambda > resource_attributes           | [resource_attributes][lambda-resource_attributes]           |                                                   | no
azure                                  | [azure][]                                                   |                                                   | no
azure > resource_attributes            | [resource_attributes][azure-resource_attributes]            |                                                   | no
aks                                    | [aks][]                                                     |                                                   | no
aks > resource_attributes              | [resource_attributes][aks-resource_attributes]              |                                                   | no
consul                                 | [consul][]                                                  |                                                   | no
consul > resource_attributes           | [resource_attributes][consul-resource_attributes]           |                                                   | no
docker                                 | [docker][]                                                  |                                                   | no
docker > resource_attributes           | [resource_attributes][docker-resource_attributes]           |                                                   | no
gcp                                    | [gcp][]                                                     |                                                   | no
gcp > resource_attributes              | [resource_attributes][gcp-resource_attributes]              |                                                   | no
heroku                                 | [heroku][]                                                  |                                                   | no
heroku > resource_attributes           | [resource_attributes][heroku-resource_attributes]           |                                                   | no
system                                 | [system][]                                                  |                                                   | no
system > resource_attributes           | [resource_attributes][system-resource_attributes]           |                                                   | no
openshift                              | [openshift][]                                               |                                                   | no
openshift > resource_attributes        | [resource_attributes][openshift-resource_attributes]        |                                                   | no
kubernetes_node                        | [kubernetes_node][]                                         |                                                   | no
kubernetes_node > resource_attributes  | [resource_attributes][kubernetes_node-resource_attributes]  |                                                   | no

[output]: #output
[ec2]: #ec2
[ec2-resource_attributes]: #ec2--resource_attributes
[ecs]: #ecs
[ecs-resource_attributes]: #ecs--resource_attributes
[eks]: #eks
[eks-resource_attributes]: #eks--resource_attributes
[elasticbeanstalk]: #elasticbeanstalk
[elasticbeanstalk-resource_attributes]: #elasticbeanstalk--resource_attributes
[lambda]: #lambda
[lambda-resource_attributes]: #lambda--resource_attributes
[azure]: #azure
[azure-resource_attributes]: #azure--resource_attributes
[aks]: #aks
[aks-resource_attributes]: #aks--resource_attributes
[consul]: #consul
[consul-resource_attributes]: #consul--resource_attributes
[docker]: #docker
[docker-resource_attributes]: #docker--resource_attributes
[gcp]: #gcp
[gcp-resource_attributes]: #gcp--resource_attributes
[heroku]: #heroku
[heroku-resource_attributes]: #heroku--resource_attributes
[system]: #system
[system-resource_attributes]: #system--resource_attributes
[openshift]: #openshift
[openshift-resource_attributes]: #openshift--resource_attributes
[kubernetes_node]: #kubernetes_node
[kubernetes_node-resource_attributes]: #kubernetes_node--resource_attributes

### output

{{< docs/shared lookup="flow/reference/components/output-block.md" source="agent" version="<AGENT VERSION>" >}}

### ec2

The `ec2` block uses [AWS SDK for Go](https://docs.aws.amazon.com/sdk-for-go/api/aws/ec2metadata/) to read 
resource information from the [EC2 instance metadata API](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html).

The following attributes are supported:

Attribute   | Type           | Description                                                                 | Default     | Required
----------- |----------------| --------------------------------------------------------------------------- |-------------| --------
`tags`      | `list(string)` | A list of regular expressions to match against tag keys of an EC2 instance. | `[]`        | no

The following blocks are supported:

Block                                          | Description                                       | Required
---------------------------------------------- | ------------------------------------------------- | --------
[resource_attributes][ec2-resource_attributes] | Configures which resource attributes to add.      | yes

### ec2 > resource_attributes

The following blocks are supported:

Block                                          | Description                                        | Required
---------------------------------------------- | -------------------------------------------------- | --------
[cloud.account.id](#resource-attribute-config) | Enables the `cloud.account.id` resource attribute. | no
`cloud.availability_zone`                      |                                                    | no
`cloud.platform`                               |                                                    | no
`cloud.provider`                               |                                                    | no
`cloud.region`                                 |                                                    | no
`host.id`                                      |                                                    | no
`host.image.id`                                |                                                    | no
`host.name`                                    |                                                    | no
`host.type`                                    |                                                    | no

### ecs

### ecs > resource_attributes

### eks

### eks > resource_attributes

### elasticbeanstalk

### elasticbeanstalk > resource_attributes

### lambda

### lambda > resource_attributes

### azure

### azure > resource_attributes

### aks

### aks > resource_attributes

### consul

### consul > resource_attributes

### docker

### docker > resource_attributes

### gcp

### gcp > resource_attributes

### heroku

### heroku > resource_attributes

### system

### system > resource_attributes

### openshift

### openshift > resource_attributes

### kubernetes_node

### kubernetes_node > resource_attributes

## Common configuration

### Resource attribute config

The following attributes are supported:

Attribute |  Type   | Description                          | Default     | Required
--------- | ------- |--------------------------------------|-------------| --------
`enabled` |  `bool` |                                      |             | yes

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` OTLP-formatted data for any telemetry signal of these types:
* logs
* metrics
* traces

## Component health

`otelcol.processor.resourcedetection` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.processor.resourcedetection` does not expose any component-specific debug
information.

## Examples

### Basic usage
