---
aliases:
- /docs/agent/latest/configuration/integrations/azure-exporter-config/
title: azure_exporter_config
---

# azure_exporter_config

## Overview
The `azure_exporter_config` block configures the `azure_exporter` integration, which is an embedded version of
[`azure-metrics-exporter`](https://github.com/webdevops/azure-metrics-exporter). This allows for
the collection of metrics from [Azure Monitor](https://azure.microsoft.com/en-us/products/monitor). The
exporter uses [Azure Resource Graph](https://azure.microsoft.com/en-us/get-started/azure-portal/resource-graph/#overview) 
queries to identify resources for gathering metrics. 

The exporter supports all metrics defined by [Azure Monitor](https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/metrics-supported).
Metrics for this integration are exposed with the template `azure_{type}_{metric}_{aggregation}_{unit}`. As an example
the Egress metric for BlobService would be exported as `azure_microsoft_storage_storageaccounts_blobservices_egress_total_bytes`.

## Authentication

The agent will need to be running in an environment which has access to azure. The exporter uses the Azure SDK for go and supports
providing authentication via https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication?tabs=bash#2-authenticate-with-azure.

The account used by the agent will need,
* [Read access to the resources which will be queried by Resource Graph](https://learn.microsoft.com/en-us/azure/governance/resource-graph/overview#permissions-in-azure-resource-graph)
* Permissions to call the [Microsoft.Insights Metrics API](https://learn.microsoft.com/en-us/rest/api/monitor/metrics/list) which should be the `Microsoft.Insights/Metrics/Read` permission

## Configuration options:

```yaml
  #
  # Common Integration Settings
  #
  
  # Enables the azure_exporter integration, allowing the Agent to automatically collect metrics or expose azure metrics 
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is self-scraped. Default will be
  # based on subscriptions and ResourceType being monitored
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled, the exporter integration will be run but not 
  # scraped and thus not. remote-written. Metrics for the integration will be exposed at 
  # /integrations/azure_exporter/metrics and can be scraped by an external process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  #
  # Exporter specific configuration
  #
  
  # Required: The azure subscription(s) to scrape metrics from
  subscriptions:
    [ - <string> ... ]
  
  # Required: The Azure Resource Type to scrape metrics for
  # Valid values can be found as the heading names on this page https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/metrics-supported
  # Ex: Microsoft.Cache/redis
  [resource_type: <string>]

  # Required: The [kusto query](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/) filter to apply when searching for resources
  # This value will be embedded in to a template query of the form `Resources | where type =~ "<resource_type>" <resource_graph_query_filter> | project id, tags`
  [resource_graph_query_filter: <string>]

  # Required: The metrics to scrape from resources
  # Valid values can be found in the `Metric` column for the`resource_type` https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/metrics-supported
  # Example: 
  #   resource_type: Microsoft.Cache/redis 
  #   metrics:
  #     - allcachehits
  metrics:
    [ - <string> ... ]
  
  # Optional: Aggregation to apply for the metrics produced. Valid values are minimum, maximum, average, total, and count
  # If no aggregation is specified the value for `Aggregation Type` on the `Metric` is used from https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/metrics-supported 
  metric_aggregations:
    [ - <string> ... ]

  # Optional: An [ISO8601 Duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) used when querying the metric value  
  [timespan: <string> | default = "PT1M"]
  
  # Optional: Used to include `Dimensions` available to a `Metric` definition https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/metrics-supported
  # These will appear as labels on the metrics,
  #   If a single dimension is requested it will have the name `dimension`
  #   If multiple dimensions are requested they will have the name `dimension<dimension_name>`
  # Example: 
  #   resource_type: Microsoft.Cache/redis 
  #   metrics:
  #     - allcachehits
  #   included_dimensions:
  #     - ShardId
  #     - Port
  #     - Primary
  included_dimensions:
    [ - <string> ... ]
  
  # Optional: A list of resource tags to include on the final metrics
  # These are added as labels with the name `tag_<tag_name>`
  included_resource_tags:
    [ - <string> ... ]

  # Optional: used for ResourceTypes which have multiple levels of metrics
  # Example: the resource_type Microsoft.Storage/storageAccounts has metrics for
  #   Microsoft.Storage/storageAccounts (generic metrics which apply to all storage accounts)
  #   Microsoft.Storage/storageAccounts/blobServices (generic metrics + metrics which only apply to blob stores)
  #   Microsoft.Storage/storageAccounts/fileServices (generic metrics + metrics which only apply to file stores)
  #   Microsoft.Storage/storageAccounts/queueServices (generic metrics + metrics which only apply to queue stores)
  #   Microsoft.Storage/storageAccounts/tableServices (generic metrics + metrics which only apply to table stores)
  # If you want blob store metrics you will need to set
  #   resource_type: Microsoft.Storage/storageAccounts
  #   metric_namespace = Microsoft.Storage/storageAccounts/blobServices
  [metric_namespace: <string>]

  # Optional: Which azure cloud environment to connect to, azurecloud, azurechinacloud, azuregovernmentcloud, or azurepprivatecloud 
  [azure_cloud_environment: <string> | default = "azurecloud"]
```

