---
aliases:
- /docs/agent/latest/flow/configuration-language/syntax
title: Syntax
weight: 200
---

# Syntax
The River syntax is designed to be easy to read and write. Essentially, there
are just two high-level elements to it; _Attributes_ and  _Blocks_. 

The main purpose of River is to define components for the Grafana Agent Flow
mode, so that they can be wired together to form programmable telemetry
pipelines.
As such, the ordering of blocks and attributes within the River configuration
file is not important; the language will consider all direct and indirect
dependencies between elements to determine the order which must be used for the
evaluation.

## Attributes and Blocks

### Attributes
_Attributes_ are used to configure individual settings. They always take the
form of `<ATTRIBUTE_NAME> = <ATTRIBUTE_VALUE>`. They can appear either as
top-level elements or nested within blocks.

The following sets the `log_level` attribute to `"debug"`.

```river
log_level = "debug"
```

All `<ATTRIBUTE_NAME>`s must be valid River identifiers: they must be made up
of one or more UTF-8 letters, digits or underscores, but cannot start with a
digit.

The `<ATTRIBUTE_VALUE>` can be either a constant value of a valid River
[type]({{< relref "./expressions/types_and_values.md" >}}) (eg. string,
boolean, number), or an [_expression_]({{< relref "./expressions/_index.md"
>}}) that computes and continuously re-evaluates more complex attribute
values.

### Blocks
_Blocks_ are used to configure components by grouping any number of attributes
or nested blocks using curly braces.

Blocks have a _name_ (the type of the component it defines), an
optional _label_ and a body that contains any number of arguments or nested
unlabeled blocks. Labels are used to disambiguate between multiple top-level
blocks of the same name.

```
// Pattern for creating a block with an optional user-label:
<BLOCK NAME> ["<BLOCK LABEL>"] {
	// Block body can contain attributes and nested unlabeled blocks
	<IDENTIFIER> = <EXPRESSION> // Attribute

	<NESTED_BLOCK_NAME> {
		// Nested block body
	}
}
```

The `<BLOCK_NAME>` must correspond to a registered Flow [component]({{< relref "./components.md" >}}).

If the optional `<BLOCK_LABEL>` is set, it has to be a valid River identifier:
consisting of one or more UTF-8 letters, digits or underscores, but it cannot
begin with a digit.

River does not allow multiple unlabelled blocks with the same name to co-exist,
so `<BLOCK_LABEL>`s are used to create multiple blocks and distinguish them.
Once a block is created, it can be referred to by combining the name and label
in a dot-delimited string like `remote.s3.production`.

The following snippet defines a block named `local.file` with its label set to
"token". The block's body sets the `filename` and `is_secret` attributes.

```river
local.file "token" {
	filename  = "/etc/agent/service-account-token"
	is_secret = true
}
```

## Comments
River configuration files support single-line `//` as well as `/* */` multiline
comments.

## Error reporting
One of the main improvements River brings to the table is its error reporting.
Whenever a syntax error is emitted, River will decorate it with information
about the position of the error, some surrounding context as well as a possible
solution.

```
Error: ./cmd/agent/example-config.river:13:1: unrecognized attribute foo

12 |
13 | foo = bar
   | ^^^
14 |

Error: ./cmd/agent/example-config.river:16:11: expected block label to be a valid identifier

15 |
16 | remote.s3 "instance-one" {
   |            ^
17 |     poll_frequency = "5m"

Error: ./cmd/agent/example-config.river:33:1: expected }, got EOF

32 | }
33 |
   |
```

## Examples
Here's an example of a simple River configuration file to showcase the
language's syntax primitives.

```river

// `logging` is a special block that sets up some global attributes
logging {
	level  = "debug"
	format = "logfmt"
}

/*
The following block defines a `remote.s3` component labelled "token".
The `path` attribute is set to the contents of the S3_TOKEN_PATH environment variable by using an expression.
The `is_secret` attribute is set to the boolean true; the file contents will not be exposed as plaintext.
*/
remote.s3 "token" {
	path      = env("S3_TOKEN_PATH") // Use an expression to read from an env var.
	is_secret = true
}
```

Feel free to move on to the [components]({{< relref "./components.md" >}}) docs
to learn about the available component types, as well as how to configure,
debug, and connect them to form a pipeline.

