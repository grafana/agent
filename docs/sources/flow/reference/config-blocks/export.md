---
aliases:
- /docs/grafana-cloud/agent/flow/reference/config-blocks/export/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/config-blocks/export/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/config-blocks/export/
- /docs/grafana-cloud/send-data/agent/flow/reference/config-blocks/export/
canonical: https://grafana.com/docs/agent/latest/flow/reference/config-blocks/export/
description: Learn about the export configuration block
menuTitle: export
title: export block
---

# export block

`export` is an optional configuration block used to specify an emitted value of a [Module][Modules].
`export` blocks must be given a label which determine the name of the export.

The `export` block may not be specified in the main configuration file given to {{< param "PRODUCT_NAME" >}}.

[Modules]: {{< relref "../../concepts/modules.md" >}}

## Example

```river
export "ARGUMENT_NAME" {
  value = ARGUMENT_VALUE
}
```

## Arguments

The following arguments are supported:

Name    | Type  | Description      | Default | Required
--------|-------|------------------|---------|---------
`value` | `any` | Value to export. |         | yes

The `value` argument determines what the value of the export will be.
To expose an exported field of another component to the module loader, set `value` to an expression which references that exported value.

## Exported fields

The `export` block doesn't export any fields.

## Example

This example creates a module where the output of discovering Kubernetes pods and nodes are exposed to the module loader:

```river
discovery.kubernetes "pods" {
  role = "pod"
}

discovery.kubernetes "nodes" {
  role = "nodes"
}

export "kubernetes_resources" {
  value = concat(
    discovery.kubernetes.pods.targets,
    discovery.kubernetes.nodes.targets,
  )
}
```
