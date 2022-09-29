---
aliases:
- /docs/agent/latest/flow/configuration-language
title: Configuration language
weight: 400
---

# Configuration language

Grafana Agent Flow contains a custom configuration language called River to
dynamically configure and connect components.

River aims to reduce errors in configuration files by making configurations
easier to read and write. River configurations are done in blocks which can be
easily copied-and-pasted from documentation to help users get started as
quickly as possible.

A River configuration file tells Grafana Agent Flow which components to launch
and how to bind them together into a pipeline.

The syntax of River is centered around blocks, attributes, and expressions:

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

> You may have noticed that River looks similar to HCL, the language used by
> Terraform and other Hashicorp projects. River was inspired by HCL, but is a
> distinct language with different syntax and features, such as first-class
> functions. If you are already familiar with HCL or Terraform, writing River
> should seem mostly natural to you.

> For historical context on why we decided to introduce River, you can read the
> original [RFC][].

* Blocks are a group of related settings, and usually represent creating a
  component. Blocks have a name which consist of zero or more identifiers
  separated by `.` (like `my_block` or `local.file` above), an optional user
  label, and a body which contains attributes and nested blocks.

* Attributes appear within blocks and assign a value to a name.

* Expressions represent a value, either literally or by referencing and
  combining other values. Expressions are used to compute a value for an
  attribute.

River is declarative, so the ordering of components, blocks, and attributes
within a block is not significant. The order of operations is determined by the
relationship between components.

[RFC]: https://github.com/grafana/agent/blob/97a55d0d908b26dbb1126cc08b6dcc18f6e30087/docs/rfcs/0005-river.md
