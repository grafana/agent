---
aliases:
- /docs/grafana-cloud/agent/flow/reference/config-blocks/declare/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/config-blocks/declare/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/config-blocks/declare/
- /docs/grafana-cloud/send-data/agent/flow/reference/config-blocks/declare/
canonical: https://grafana.com/docs/agent/latest/flow/reference/config-blocks/declare/
description: Learn about the declare configuration block
menuTitle: declare
title: declare block
refs:
  import:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/concepts/modules/#importing-modules
  export:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/config-blocks/export/
  declare:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/config-blocks/declare/
    - pattern: /docs/grafana-cloud/
      destination: /docs/grafana-cloud/send-data/agent/flow/reference/config-blocks/declare/
  argument:
    - pattern: /docs/agent/
      destination: /docs/agent/<AGENT_VERSION>/flow/reference/config-blocks/argument/
---

# declare block

`declare` is an optional configuration block used to define a new [custom component][].
`declare` blocks must be given a label that determines the name of the custom component.

## Example

```river
declare "COMPONENT_NAME" {
    COMPONENT_DEFINITION
}
```

## Arguments

The `declare` block has no predefined schema for its arguments.
The body of the `declare` block is used as the component definition.
The body can contain the following:

* [argument](ref:argument) blocks
* [export](ref:export) blocks
* [declare](ref:declare) blocks
* [import](ref:import) blocks
* Component definitions (either built-in or custom components)

The `declare` block may not contain any configuration blocks that aren't listed above.

## Exported fields

The `declare` block has no predefined schema for its exports.
The fields exported by the `declare` block are determined by the [export blocks](ref:export) found in its definition.

## Example

This example creates and uses a custom component that self-collects process metrics and forwards them to an argument specified by the user of the custom component:

```river
declare "self_collect" {
  argument "metrics_output" {
    optional = false
    comment  = "Where to send collected metrics."
  }

  prometheus.scrape "selfmonitor" {
    targets = [{
      __address__ = "127.0.0.1:12345",
    }]

    forward_to = [argument.metrics_output.value]
  }
}

self_collect "example" {
  metrics_output = prometheus.remote_write.example.receiver
}

prometheus.remote_write "example" {
  endpoint {
    url = REMOTE_WRITE_URL
  }
}
```

