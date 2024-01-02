---
aliases:
- ../../../configuration/integrations/azure-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/azure-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/azure-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/azure-exporter-config/
description: Learn about azure_exporter_config
title: azure_exporter_config
---

# azure_exporter_config

## Overview
The `azure_exporter_config` block configures the `azure_exporter` integration, an embedded version of
[`azure-metrics-exporter`](https://github.com/webdevops/azure-metrics-exporter), used to
collect metrics from [Azure Monitor](https://azure.microsoft.com/en-us/products/monitor). 

The exporter offers the following two options for gathering metrics.

1. (Default) Use an [Azure Resource Graph](https://azure.microsoft.com/en-us/get-started/azure-portal/resource-graph/#overview) query to identify resources for gathering metrics.
   1. This query will make one API call per resource identified.
   1. Subscriptions with a reasonable amount of resources can hit the [12000 requests per hour rate limit](https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/request-limits-and-throttling#subscription-and-tenant-limits) Azure enforces.
1. Set the regions to gather metrics from and get metrics for all resources across those regions.
   1. This option will make one API call per subscription, dramatically reducing the number of API calls. 
   1. This approach does not work with all resource types, and Azure does not document which resource types do or do not work.
   1. A resource type that is not supported produces errors that look like `Resource type: microsoft.containerservice/managedclusters not enabled for Cross Resource metrics`.
   1. If you encounter one of these errors you must use the default Azure Resource Graph based option to gather metrics.

## List of Supported Services and Metrics
The exporter supports all metrics defined by Azure Monitor. The complete list of available metrics can be found in the [Azure Monitor documentation](https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/metrics-supported).
Metrics for this integration are exposed with the template `azure_{type}_{metric}_{aggregation}_{unit}`. As an example,
the Egress metric for BlobService would be exported as `azure_microsoft_storage_storageaccounts_blobservices_egress_total_bytes`.

## Authentication

The agent must be running in an environment with access to Azure. The exporter uses the Azure SDK for go and supports authentication via https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication?tabs=bash#2-authenticate-with-azure.

The account used by Grafana Agent needs:
* [Read access to the resources that will be queried by Resource Graph](https://learn.microsoft.com/en-us/azure/governance/resource-graph/overview#permissions-in-azure-resource-graph)
* Permissions to call the [Microsoft.Insights Metrics API](https://learn.microsoft.com/en-us/rest/api/monitor/metrics/list) which should be the `Microsoft.Insights/Metrics/Read` permission

## Configuration

### Config Reference

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

  # Required: The metrics to scrape from resources
  # Valid values can be found in the `Metric` column for the`resource_type` https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/metrics-supported
  # Example:
  #   resource_type: Microsoft.Cache/redis
  #   metrics:
  #     - allcachehits
  metrics:
    [ - <string> ... ]

  # Optional: The [kusto query](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/) filter to apply when searching for resources
  # This value will be embedded in to a template query of the form `Resources | where type =~ "<resource_type>" <resource_graph_query_filter> | project id, tags`
  # Can't be used if `regions` is set.
  [resource_graph_query_filter: <string>]
  
  # Optional: The list of regions for gathering metrics. Enables gathering metrics for all resources in the subscription.
  # The list of available `regions` to your subscription can be found by running the Azure CLI command `az account list-locations --query '[].name'`.
  # Can't be used if `resource_graph_query_filter` is set.
  regions:
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
  
  # Optional: Validation is disabled by default to reduce the number of Azure exporter instances required when a `resource_type` has metrics with varying dimensions. 
  # Choosing to enable `validate_dimensions` will require one exporter instance per metric + dimension combination which can be very tedious to maintain.
  [validate_dimensions: <bool> | default = false]
```

### Examples

#### Azure Kubernetes Service Node Metrics
```yaml
  azure_exporter:
    enabled: true
    scrape_interval: 60s
    subscriptions:
      - <subscription_id>
    resource_type: microsoft.containerservice/managedclusters
    metrics:
      - node_cpu_usage_millicores
      - node_cpu_usage_percentage
      - node_disk_usage_bytes
      - node_disk_usage_percentage
      - node_memory_rss_bytes
      - node_memory_rss_percentage
      - node_memory_working_set_bytes
      - node_memory_working_set_percentage
      - node_network_in_bytes
      - node_network_out_bytes
    included_resource_tags:
      - environment
    included_dimensions:
      - node
      - nodepool
      - device
```

#### Blob Storage Metrics
```yaml
  azure_exporter:
    enabled: true
    scrape_interval: 60s
    subscriptions:
      - <subscription_id>
    resource_type: Microsoft.Storage/storageAccounts
    metric_namespace: Microsoft.Storage/storageAccounts/blobServices
    regions:
      - westeurope
    metrics:
      - Availability
      - BlobCapacity
      - BlobCount
      - ContainerCount
      - Egress
      - IndexCapacity
      - Ingress
      - SuccessE2ELatency
      - SuccessServerLatency
      - Transactions
    included_dimensions:
      - ApiName
      - TransactionType
    timespan: PT1H
```

### Multiple Azure Services in a single config

The Azure Metrics API has rather strict limitations on the number of parameters which can be supplied. Due to this, you cannot
gather metrics from multiple `resource_types` in the same `azure_exporter` instance. If you need metrics from multiple resources,
you can enable `integration-next` or configure Agent to expose the exporter via the `azure_exporter` config with data configured through metrics scrape_configs. The following example configuration combines the two examples above in a single Agent configuration.

> **Note**: This is not a complete configuration; blocks have been removed for simplicity.

```yaml
integrations:
  azure_exporter:
    enabled: true
    scrape_integration: false
    azure_cloud_environment: azurecloud

metrics:
  configs:
    - name: integrations
      scrape_configs:
        - job_name: azure-blob-storage
          scrape_interval: 1m
          scrape_timeout: 50s
          static_configs:
            - targets: ["localhost:12345"]
          metrics_path: /integrations/azure_exporter/metrics
          params:
            subscriptions:
              - 179c4f30-ebd8-489e-92bc-fb64588dadb3
            resource_type: ["Microsoft.Storage/storageAccounts"]
            regions:
              - westeurope
            metric_namespace: ["Microsoft.Storage/storageAccounts/blobServices"]
            metrics:
              - Availability
              - BlobCapacity
              - BlobCount
              - ContainerCount
              - Egress
              - IndexCapacity
              - Ingress
              - SuccessE2ELatency
              - SuccessServerLatency
              - Transactions
            included_dimensions:
              - ApiName
              - TransactionType
            timespan: ["PT1H"]
        - job_name: azure-kubernetes-node
          scrape_interval: 1m
          scrape_timeout: 50s
          static_configs:
            - targets: ["localhost:12345"]
          metrics_path: /integrations/azure_exporter/metrics
          params:
            subscriptions:
              - 179c4f30-ebd8-489e-92bc-fb64588dadb3
            resource_type: ["microsoft.containerservice/managedclusters"]
            resource_graph_query_filter: [" where location == 'westeurope'"]
            metrics:
              - node_cpu_usage_millicores
              - node_cpu_usage_percentage
              - node_disk_usage_bytes
              - node_disk_usage_percentage
              - node_memory_rss_bytes
              - node_memory_rss_percentage
              - node_memory_working_set_bytes
              - node_memory_working_set_percentage
              - node_network_in_bytes
              - node_network_out_bytes
            included_resource_tags:
              - environment
            included_dimensions:
              - node
              - nodepool
              - device
```

In this example, all `azure_exporter`-specific configuration settings have been moved to the `scrape_config`. This method supports all available configuration options except `azure_cloud_environment`, which must be configured on the `azure_exporter`. For this method, if a field supports a singular value like `resource_graph_query_filter`, you
must be put it into an array, for example, `resource_graph_query_filter: ["where location == 'westeurope'"]`.
