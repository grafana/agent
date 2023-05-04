# Authenticator - Sigv4

| Status                   |                      |
| ------------------------ |----------------------|
| Stability                | [beta]               |
| Distributions            | [contrib]            |

This extension provides Sigv4 authentication for making requests to AWS services. For more information on the Sigv4 process, please look [here](https://docs.aws.amazon.com/general/latest/gr/signature-version-4.html).

## Configuration

The configuration fields are as follows:

* `assume_role`: **Optional**. Specifies the configuration needed to assume a role
  * `arn`: The Amazon Resource Name (ARN) of a role to assume
  * `session_name`: **Optional**. The name of a role session
  * `sts_region`: The AWS region where STS is used to assumed the configured role
    * Note that if a role is intended to be assumed, and `sts_region` is not provided, then `sts_region` will default to the value for `region` if `region` is provided
* `region`: **Optional**. The AWS region for the service you are exporting to for AWS Sigv4. This is differentiated from `sts_region` to handle cross region authentication
    * Note that an attempt will be made to obtain a valid region from the endpoint of the service you are exporting to
    * [List of AWS regions](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html)
* `service`: **Optional**. The AWS service for AWS Sigv4
    * Note that an attempt will be made to obtain a valid service from the endpoint of the service you are exporting to


```yaml
extensions:
  sigv4auth:
    assume_role:
      arn: "arn:aws:iam::123456789012:role/aws-service-role/access"
      sts_region: "us-east-1"

receivers:
  hostmetrics:
    scrapers:
      memory:

exporters:
  prometheusremotewrite:
    endpoint: "https://aps-workspaces.us-west-2.amazonaws.com/workspaces/ws-XXX/api/v1/remote_write"
    auth:
      authenticator: sigv4auth

service:
  extensions: [sigv4auth]
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: []
      exporters: [prometheusremotewrite]
```

## Notes

* The collector must have valid AWS credentials as used by the [AWS SDK for Go](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials)

[beta]:https://github.com/open-telemetry/opentelemetry-collector#beta
[contrib]:https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
