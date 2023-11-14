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

Hierarchy                              | Block                                    | Description                                       | Required
-------------------------------------- | ---------------------------------------- | ------------------------------------------------- | --------
output                                 | [output][]                               | Configures where to send received telemetry data. | yes
ec2                                    | [ec2][]                                  | Configures where to send received telemetry data. | no
ec2 > resource_attributes              | [ec2-resource_attributes][]              | Configures where to send received telemetry data. | no
ecs                                    | [ecs][]                                  | Configures where to send received telemetry data. | no
ecs > resource_attributes              | [ecs-resource_attributes][]              | Configures where to send received telemetry data. | no
eks                                    | [eks][]                                  | Configures where to send received telemetry data. | no
eks > resource_attributes              | [eks-resource_attributes][]              | Configures where to send received telemetry data. | no
elasticbeanstalk                       | [elasticbeanstalk][]                     | Configures where to send received telemetry data. | no
elasticbeanstalk > resource_attributes | [elasticbeanstalk-resource_attributes][] | Configures where to send received telemetry data. | no
lambda                                 | [lambda][]                               | Configures where to send received telemetry data. | no
lambda > resource_attributes           | [lambda-resource_attributes][]           | Configures where to send received telemetry data. | no
azure                                  | [azure][]                                | Configures where to send received telemetry data. | no
azure > resource_attributes            | [azure-resource_attributes][]            | Configures where to send received telemetry data. | no
aks                                    | [aks][]                                  | Configures where to send received telemetry data. | no
aks > resource_attributes              | [aks-resource_attributes][]              | Configures where to send received telemetry data. | no
consul                                 | [consul][]                               | Configures where to send received telemetry data. | no
consul > resource_attributes           | [consul-resource_attributes][]           | Configures where to send received telemetry data. | no
docker                                 | [docker][]                               | Configures where to send received telemetry data. | no
docker > resource_attributes           | [docker-resource_attributes][]           | Configures where to send received telemetry data. | no
gcp                                    | [gcp][]                                  | Configures where to send received telemetry data. | no
gcp > resource_attributes              | [gcp-resource_attributes][]              | Configures where to send received telemetry data. | no
heroku                                 | [heroku][]                               | Configures where to send received telemetry data. | no
heroku > resource_attributes           | [heroku-resource_attributes][]           | Configures where to send received telemetry data. | no
system                                 | [system][]                               | Configures where to send received telemetry data. | no
system > resource_attributes           | [system-resource_attributes][]           | Configures where to send received telemetry data. | no
openshift                              | [openshift][]                            | Configures where to send received telemetry data. | no
openshift > resource_attributes        | [openshift-resource_attributes][]        | Configures where to send received telemetry data. | no
kubernetes_node                        | [kubernetes_node][]                      | Configures where to send received telemetry data. | no
kubernetes_node > resource_attributes  | [kubernetes_node-resource_attributes][]  | Configures where to send received telemetry data. | no

[output]: #output-block
[ec2]: #ec2-block
[ec2-resource_attributes]: #ec2-resource_attributes-block
[ecs]: #ecs-block
[ecs-resource_attributes]: #ecs-resource_attributes-block
[eks]: #eks-block
[eks-resource_attributes]: #eks-resource_attributes-block
[elasticbeanstalk]: #elasticbeanstalk-block
[elasticbeanstalk-resource_attributes]: #elasticbeanstalk-resource_attributes-block
[lambda]: #lambda-block
[lambda-resource_attributes]: #lambda-resource_attributes-block
[azure]: #azure-block
[azure-resource_attributes]: #azure--resource_attributesblock
[aks]: #aks-block
[aks-resource_attributes]: #aks-resource_attributes-block
[consul]: #consul-block
[consul-resource_attributes]: #consul-resource_attributes-block
[docker]: #docker-block
[docker-resource_attributes]: #docker-resource_attributes-block
[gcp]: #gcp-block
[gcp-resource_attributes]: #gcp-resource_attributes-block
[heroku]: #heroku-block
[heroku-resource_attributes]: #heroku-resource_attributes-block
[system]: #system-block
[system-resource_attributes]: #system-resource_attributes-block
[openshift]: #openshift-block
[openshift-resource_attributes]: #openshift-resource_attributes-block
[kubernetes_node]: #kubernetes_node-block
[kubernetes_node-resource_attributes]: #kubernetes_node-resource_attributes-block

### output block

{{< docs/shared lookup="flow/reference/components/output-block.md" source="agent" version="<AGENT VERSION>" >}}

### ec2 block

The `ec2` block uses [AWS SDK for Go](https://docs.aws.amazon.com/sdk-for-go/api/aws/ec2metadata/) to read 
resource information from the [EC2 instance metadata API](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html).

The following attributes are supported:

Name   | Type           | Description                          | Default     | Required
------ |----------------|--------------------------------------|-------------| --------
`tags` | `list(string)` |                                      | `[]`        | no

### ec2 > resource_attributes block

The following attributes are supported:

Name   | Type           | Description                          | Default     | Required
------ |----------------|--------------------------------------|-------------| --------
`name` | `string` |                                      | `[]`        | no

### ecs block

### ecs > resource_attributes block

### eks block

### eks > resource_attributes block

### elasticbeanstalk block

### elasticbeanstalk > resource_attributes block

### lambda block

### lambda > resource_attributes block

### azure block

### azure > resource_attributes block

### aks block

### aks > resource_attributes block

### consul block

### consul > resource_attributes block

### docker block

### docker > resource_attributes block

### gcp block

### gcp > resource_attributes block

### heroku block

### heroku > resource_attributes block

### system block

### system > resource_attributes block

### openshift block

### openshift > resource_attributes block

### kubernetes_node block

### kubernetes_node > resource_attributes block

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

```river
otelcol.processor.resourcedetection "default" {

  output {
    logs = [otelcol.exporter.otlp.default.input]
  }
}
```

### Sample 15% of the logs

```river
otelcol.processor.resourcedetection "default" {
  sampling_percentage = 15

  output {
    logs = [otelcol.exporter.otlp.default.input]
  }
}
```

### Sample logs according to their "logID" attribute

```river
otelcol.processor.resourcedetection "default" {
  sampling_percentage = 15
  attribute_source    = "record"
  from_attribute      = "logID"

  output {
    logs = [otelcol.exporter.otlp.default.input]
  }
}
```

### Sample logs according to a "priority" attribute 

```river
otelcol.processor.resourcedetection "default" {
  sampling_percentage = 15
  sampling_priority   = "priority"

  output {
    logs = [otelcol.exporter.otlp.default.input]
  }
}
```
