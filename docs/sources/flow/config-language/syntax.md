---
aliases:
- ../configuration-language/syntax/
title: Syntax
weight: 200
---

# Syntax

The River syntax is designed to be easy to read and write. Essentially, there
are just two high-level elements to it: _Attributes_ and _Blocks_.

River is a _declarative_ language used to build programmable pipelines.
As such, the ordering of blocks and attributes within the River configuration
file is not important; the language will consider all direct and indirect
dependencies between elements to determine their relationships.

## Comments

River configuration files support single-line `//` as well as block `/* */`
comments.

## Identifiers

River considers an identifier as valid if it consists of one or more UTF-8
letters (A through Z, both upper- and lower-case), digits or underscores, but
doesn't start with a digit.

## Attributes and Blocks

### Attributes

_Attributes_ are used to configure individual settings. They always take the
form of `ATTRIBUTE_NAME = ATTRIBUTE_VALUE`. They can appear either as
top-level elements or nested within blocks.

The following example sets the `log_level` attribute to `"debug"`.

```river
log_level = "debug"
```

The `ATTRIBUTE_NAME` must be a valid River [identifier](#identifier).

The `ATTRIBUTE_VALUE` can be either a constant value of a valid River
[type]({{< relref "./expressions/types_and_values.md" >}}) (eg. string,
boolean, number) or an [_expression_]({{< relref "./expressions/_index.md" >}})
to represent or compute more complex attribute values.

### Blocks

_Blocks_ are used to configure the Agent behavior as well as Flow components by
grouping any number of attributes or nested blocks using curly braces.
Blocks have a _name_, an optional _label_ and a body that contains any number
of arguments and nested unlabeled blocks.

#### Pattern for creating an unlabeled block

```river
BLOCK_NAME {
  // Block body can contain attributes and nested unlabeled blocks
  IDENTIFIER = EXPRESSION // Attribute

  NESTED_BLOCK_NAME {
    // Nested block body
  }
}
```

#### Pattern for creating a labeled block

```river
// Pattern for creating a labeled block:
BLOCK_NAME "BLOCK_LABEL" {
  // Block body can contain attributes and nested unlabeled blocks
  IDENTIFIER = EXPRESSION // Attribute

  NESTED_BLOCK_NAME {
    // Nested block body
  }
}
```

#### Block naming rules

The `BLOCK_NAME` has to be recognized by Flow as either a valid component
name or a special block for configuring global settings. If the `BLOCK_LABEL`
has to be set, it must be a valid River [identifier](#identifiers) wrapped in
double quotes. In these cases the label will be used to disambiguate between
multiple top-level blocks of the same name.

The following snippet defines a block named `local.file` with its label set to
"token". The block's body sets the to the contents of the `TOKEN_FILE_PATH`
environment variable by using an expression and the `is_secret` attribute is
set to the boolean `true`, marking the file content as sensitive.
```river
local.file "token" {
  filename  = env("TOKEN_FILE_PATH") // Use an expression to read from an env var.
  is_secret = true
}
```

## Terminators

All block and attribute definitions are followed by a newline, which River
calls a _terminator_, as it terminates the current statement.

A newline is treated as terminator when it follows any expression, `]`,
`)` or `}`. Other newlines are ignored by River and and a user can enter as many
newlines as they want.

