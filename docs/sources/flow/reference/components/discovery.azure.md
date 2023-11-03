---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/discovery.azure/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.azure/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/discovery.azure/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.azure/
title: discovery.azure
description: Learn about discovery.azure
---

# discovery.azure

`discovery.azure` discovers [Azure][] Virtual Machines and exposes them as targets.

[Azure]: https://azure.microsoft.com/en-us

## Usage

```river
discovery.azure "LABEL" {
}
```

## Arguments

The following arguments are supported:

Name                | Type       | Description                                                            | Default              | Required
------------------- | ---------- | ---------------------------------------------------------------------- | -------------------- | --------
`environment`       | `string`   | Azure environment.                                                     | `"AzurePublicCloud"` | no
`port`              | `number`   | Port to be appended to the `__address__` label for each target.        | `80`                 | no
`subscription_id`   | `string`   | Azure subscription ID.                                                 |                      | no
`refresh_interval`  | `duration` | Interval at which to refresh the list of targets.                      | `5m`                 | no
`proxy_url`         | `string`   | HTTP proxy to proxy requests through.                                  |                      | no
`follow_redirects`  | `bool`     | Whether redirects returned by the server should be followed.           | `true`               | no
`enable_http2`      | `bool`     | Whether HTTP2 is supported for requests.                               | `true`               | no

## Blocks
The following blocks are supported inside the definition of
`discovery.azure`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
oauth | [oauth][] | OAuth configuration for Azure API. | no
managed_identity | [managed_identity][] | Managed Identity configuration for Azure API. | no
tls_config | [tls_config][] | TLS configuration for requests to the Azure API. | no

Exactly one of the `oauth` or `managed_identity` blocks must be specified.

[oauth]: #oauth-block
[managed_identity]: #managed_identity-block
[tls_config]: #tls_config-block

### oauth block
The `oauth` block configures OAuth authentication for the Azure API.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`client_id` | `string` | OAuth client ID. | | yes
`client_secret` | `string` | OAuth client secret. | | yes
`tenant_id` | `string` | OAuth tenant ID. | | yes

### managed_identity block
The `managed_identity` block configures Managed Identity authentication for the Azure API.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`client_id` | `string` | Managed Identity client ID. | | yes

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The set of targets discovered from the Azure API.

Each target includes the following labels:

* `__meta_azure_subscription_id`: The Azure subscription ID.
* `__meta_azure_tenant_id`: The Azure tenant ID.
* `__meta_azure_machine_id`: The UUID of the Azure VM.
* `__meta_azure_machine_resource_group`: The name of the resource group the VM is in.
* `__meta_azure_machine_name`: The name of the VM.
* `__meta_azure_machine_computer_name`: The host OS name of the VM.
* `__meta_azure_machine_os_type`: The OS the VM is running (either `Linux` or `Windows`).
* `__meta_azure_machine_location`: The region the VM is in.
* `__meta_azure_machine_private_ip`: The private IP address of the VM.
* `__meta_azure_machine_public_ip`: The public IP address of the VM.
* `__meta_azure_machine_tag_*`: A tag on the VM. There will be one label per tag.
* `__meta_azure_machine_scale_set`: The name of the scale set the VM is in.
* `__meta_azure_machine_size`: The size of the VM.

Each discovered VM maps to a single target. The `__address__` label is set to the `private_ip:port` (`[private_ip]:port` if the private IP is an IPv6 address) of the VM.

## Component health

`discovery.azure` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.azure` does not expose any component-specific debug information.

## Debug metrics

`discovery.azure` does not expose any component-specific debug metrics.

## Example

```river
discovery.azure "example" {
  port = 80
  subscription_id = AZURE_SUBSCRIPTION_ID
  oauth {
      client_id = AZURE_CLIENT_ID
      client_secret = AZURE_CLIENT_SECRET
      tenant_id = AZURE_TENANT_ID
  }
}

prometheus.scrape "demo" {
  targets    = discovery.azure.example.targets
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
  - `AZURE_SUBSCRIPTION_ID`: Your Azure subscription ID.
  - `AZURE_CLIENT_ID`: Your Azure client ID.
  - `AZURE_CLIENT_SECRET`: Your Azure client secret.
  - `AZURE_TENANT_ID`: Your Azure tenant ID.
  - `PROMETHEUS_REMOTE_WRITE_URL`: The URL of the Prometheus remote_write-compatible server to send metrics to.
  - `USERNAME`: The username to use for authentication to the remote_write API.
  - `PASSWORD`: The password to use for authentication to the remote_write API.
