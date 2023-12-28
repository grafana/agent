---
aliases:
- /docs/grafana-cloud/agent/flow/config-language/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/config-language/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/config-language/
- /docs/grafana-cloud/send-data/agent/flow/config-language/
- configuration-language/
canonical: https://grafana.com/docs/agent/latest/flow/config-language/
description: Learn about the configuration language
title: Configuration language
weight: 400
---

# Configuration language

{{< param "PRODUCT_NAME" >}} dynamically configures and connects components with a custom configuration language called River.

River aims to reduce errors in configuration files by making configurations easier to read and write.
River configurations use blocks that can be easily copied and pasted from the documentation to help you get started as quickly as possible.

A River configuration file tells {{< param "PRODUCT_NAME" >}} which components to launch and how to bind them together into a pipeline.

The River syntax uses blocks, attributes, and expressions.

```river
// Create a local.file component labeled my_file.
// This can be referenced by other components as local.file.my_file.
local.file "my_file" {
  filename = "/tmp/my-file.txt"
}

// Pattern for creating a labeled block, which the above block follows:
BLOCK_NAME "BLOCK_LABEL" {
  // Block body
  IDENTIFIER = EXPRESSION // Attribute
}

// Pattern for creating an unlabeled block:
BLOCK_NAME {
  // Block body
  IDENTIFIER = EXPRESSION // Attribute
}
```

[River][RFC] is similar to HCL, the language Terraform and other Hashicorp projects use.
It's a distinct language with custom syntax and features, such as first-class functions.

* Blocks are a group of related settings and usually represent creating a component.
  Blocks have a name that consists of zero or more identifiers separated by `.`, an optional user label, and a body containing attributes and nested blocks.
* Attributes appear within blocks and assign a value to a name.
* Expressions represent a value, either literally or by referencing and combining other values.
  You use expressions to compute a value for an attribute.

River is declarative, so ordering components, blocks, and attributes within a block isn't significant.
The relationship between components determines the order of operations.

## Tooling

You can use one or all of the following tools to help you write configuration files in River.

* Experimental editor support for
  * [vim](https://github.com/rfratto/vim-river)
  * [VSCode](https://github.com/rfratto/vscode-river)
  * [river-mode](https://github.com/jdbaldry/river-mode) for Emacs
* Code formatting using the [`agent fmt` command][fmt]

You can also start developing your own tooling using the {{< param "PRODUCT_ROOT_NAME" >}} repository as a go package or use the
[tree-sitter grammar][] with other programming languages.

[RFC]: https://github.com/grafana/agent/blob/97a55d0d908b26dbb1126cc08b6dcc18f6e30087/docs/rfcs/0005-river.md
[vim]: https://github.com/rfratto/vim-river
[VSCode]: https://github.com/rfratto/vscode-river
[river-mode]: https://github.com/jdbaldry/river-mode
[tree-sitter grammar]: https://github.com/grafana/tree-sitter-river

{{% docs/reference %}}
[fmt]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/cli/fmt"
[fmt]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/cli/fmt"
{{% /docs/reference %}}