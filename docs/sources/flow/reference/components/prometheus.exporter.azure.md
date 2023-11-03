---
aliases:
  - /docs/grafana-cloud/agent/flow/reference/components/prometheus.exporter.azure/
  - /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/prometheus.exporter.azure/
  - /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/prometheus.exporter.azure/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.exporter.azure/
title: prometheus.exporter.azure
description: Learn about prometheus.exporter.azure
---

# prometheus.exporter.azure

The `prometheus.exporter.azure` component embeds [`azure-metrics-exporter`](https://github.com/webdevops/azure-metrics-exporter) to collect metrics from [Azure Monitor](https://azure.microsoft.com/en-us/products/monitor). The exporter uses [Azure Resource Graph](https://azure.microsoft.com/en-us/get-started/azure-portal/resource-graph/#overview) queries to identify resources for gathering metrics.

The exporter supports all metrics defined by Azure Monitor. You can find the complete list of available metrics in the [Azure Monitor documentation](https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/metrics-supported).
Metrics for this integration are exposed with the template `azure_{type}_{metric}_{aggregation}_{unit}` by default. As an example,
the Egress metric for BlobService would be exported as `azure_microsoft_storage_storageaccounts_blobservices_egress_total_bytes`.

## Authentication

Grafana agent must be running in an environment with access to Azure. The exporter uses the Azure SDK for go and supports [authentication](https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication?tabs=bash#2-authenticate-with-azure).

The account used by Grafana Agent needs:

- [Read access to the resources that will be queried by Resource Graph](https://learn.microsoft.com/en-us/azure/governance/resource-graph/overview#permissions-in-azure-resource-graph)
- Permissions to call the [Microsoft.Insights Metrics API](https://learn.microsoft.com/en-us/rest/api/monitor/metrics/list) which should be the `Microsoft.Insights/Metrics/Read` permission

## Usage

```river
prometheus.exporter.azure LABEL {
        subscriptions = [
                SUB_ID_1,
                SUB_ID_2,
                ...
        ]

        resource_type = RESOURCE_TYPE

        metrics = [
                METRIC_1,
                METRIC_2,
                ...
        ]
}
```

## Arguments

You can use the following arguments to configure the exporter's behavior.
Omitted fields take their default values.

| Name                          | Type           | Description                                                          | Default                                                                       | Required |
| ----------------------------- | -------------- | -------------------------------------------------------------------- | ----------------------------------------------------------------------------- | -------- |
| `subscriptions`               | `list(string)` | List of subscriptions to scrap metrics from.                         |                                                                               | yes      |
| `resource_type`               | `string`       | The Azure Resource Type to scrape metrics for.                       |                                                                               | yes      |
| `metrics`                     | `list(string)` | The metrics to scrape from resources.                                |                                                                               | yes      |
| `resource_graph_query_filter` | `string`       | The [Kusto query][] filter to apply when searching for resources.    |                                                                               | no       |
| `metric_aggregations`         | `list(string)` | Aggregations to apply for the metrics produced.                      |                                                                               | no       |
| `timespan`                    | `string`       | [ISO8601 Duration][] over which the metrics are being queried.       | `"PT1M"` (1 minute)                                                           | no       |
| `included_dimensions`         | `list(string)` | List of dimensions to include on the final metrics.                  |                                                                               | no       |
| `included_resource_tags`      | `list(string)` | List of resource tags to include on the final metrics.               | `["owner"]`                                                                   | no       |
| `metric_namespace`            | `string`       | Namespace for `resource_type` which have multiple levels of metrics. |                                                                               | no       |
| `azure_cloud_environment`     | `string`       | Name of the cloud environment to connect to.                         | `"azurecloud"`                                                                | no       |
| `metric_name_template`        | `string`       | Metric template used to expose the metrics.                          | `"azure_{type}_{metric}_{aggregation}_{unit}"`                                | no       |
| `metric_help_template`        | `string`       | Description of the metric.                                           | `"Azure metric {metric} for {type} with aggregation {aggregation} as {unit}"` | no       |

The list of available `resource_type` values and their corresponding `metrics` can be found in [Azure Monitor essentials][].

The `resource_graph_query_filter` can be embedded into a template query of the form `Resources | where type =~ "<resource_type>" <resource_graph_query_filter> | project id, tags`.

Valid values for `metric_aggregations` are `minimum`, `maximum`, `average`, `total`, and `count`. If no aggregation is specified, the value is retrieved from the metric. For example, the aggregation value of the metric `Availability` in [Microsoft.ClassicStorage/storageAccounts](https://learn.microsoft.com/en-us/azure/azure-monitor/reference/supported-metrics/microsoft-classicstorage-storageaccounts-metrics) is `average`.

Every metric has its own set of dimensions. For example, the dimensions for the metric `Availability` in [Microsoft.ClassicStorage/storageAccounts](https://learn.microsoft.com/en-us/azure/azure-monitor/reference/supported-metrics/microsoft-classicstorage-storageaccounts-metrics) are `GeoType`, `ApiName`, and `Authentication`. If a single dimension is requested, it will have the name `dimension`. If multiple dimensions are requested, they will have the name `dimension<dimension_name>`.

Tags in `included_resource_tags` will be added as labels with the name `tag_<tag_name>`.

Valid values for `azure_cloud_environment` are `azurecloud`, `azurechinacloud`, `azuregovernmentcloud` and `azurepprivatecloud`.

[Kusto query]: https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/
[Azure Monitor essentials]: https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/metrics-supported
[ISO8601 Duration]: https://en.wikipedia.org/wiki/ISO_8601#Durations

## Exported fields

{{< docs/shared lookup="flow/reference/components/exporter-component-exports.md" source="agent" version="<AGENT VERSION>" >}}

## Component health

`prometheus.exporter.azure` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last healthy values.

## Debug information

`prometheus.exporter.azure` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.azure` does not expose any component-specific
debug metrics.

## Examples

```river
prometheus.exporter.azure "example" {
	subscriptions    = SUBSCRIPTIONS
	resource_type    = "Microsoft.Storage/storageAccounts"
	metric_namespace = "Microsoft.Storage/storageAccounts/blobServices"
	metrics          = [
		"Availability",
		"BlobCapacity",
		"BlobCount",
		"ContainerCount",
		"Egress",
		"IndexCapacity",
		"Ingress",
		"SuccessE2ELatency",
		"SuccessServerLatency",
		"Transactions",
	]
	timespan                    = "PT1H"
	resource_graph_query_filter = "where location == 'westeurope'"
}

// Configure a prometheus.scrape component to send metrics to.
prometheus.scrape "demo" {
	targets    = prometheus.exporter.azure.example.targets
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

- `SUBSCRIPTIONS`: The Azure subscription IDs holding the resources you are interested in.
- `PROMETHEUS_REMOTE_WRITE_URL`: The URL of the Prometheus remote_write-compatible server to send metrics to.
- `USERNAME`: The username to use for authentication to the remote_write API.
- `PASSWORD`: The password to use for authentication to the remote_write API.
