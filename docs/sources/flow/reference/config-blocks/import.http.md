---
aliases:
- /docs/grafana-cloud/agent/flow/reference/config-blocks/import.http/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/config-blocks/import.http/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/config-blocks/import.http/
- /docs/grafana-cloud/send-data/agent/flow/reference/config-blocks/import.http/
canonical: https://grafana.com/docs/agent/latest/flow/reference/config-blocks/import.http/
description: Learn about the import.http configuration block
labels:
  stage: beta
title: import.http
---

# import.http

{{< docs/shared lookup="flow/stability/beta.md" source="agent" version="<AGENT_VERSION>" >}}

`import.http` retrieves a module from a HTTP server.

[module]: {{< relref "../../concepts/modules.md" >}}

## Usage

```river
import.http "LABEL" {
  url = URL
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`url` | `string` | URL to poll. | | yes
`method` | `string` | Define HTTP method for the request | `"GET"` | no
`headers` | `map(string)` | Custom headers for the request. | `{}` | no
`poll_frequency` | `duration` | Frequency to poll the URL. | `"1m"` | no
`poll_timeout` | `duration` | Timeout when polling the URL. | `"10s"` | no
`is_secret` | `bool` | Whether the response body should be treated as a secret. | false | no

[secret]: {{< relref "../../concepts/config-language/expressions/types_and_values.md#secrets" >}}

## Component health

`module.http` is reported as healthy if the most recent load of the module was successful.
If the module is not loaded successfully, the current health displays as unhealthy, and the health includes the error from loading the module.

## Debug information

`import.http` does not expose any component-specific debug information.

## Debug metrics

`import.http` does not expose any component-specific debug metrics.

## Example

This example imports a module from a HTTP server and instantiates a custom component from an imported declare to that adds two numbers:

Main configuration:

```river
import.http "math" {
  filename = "/path/to/math.river"
}
math.add "default" {
  a = 15
  b = 45
}
```

Module:

```river
declare "add" {
    argument "a" {}
    argument "b" {}
    export "sum" {
        value = argument.a.value + argument.b.value
    }
}
```