---
aliases:
- /docs/agent/latest/flow/configuration-language/syntax
title: Syntax
weight: 200
---

# Syntax
The River syntax is designed to be easy to read and write. Essentially, there
are just two high-level elements to it: _Attributes_ and  _Blocks_. 

River is a _declarative_ language used to build programmable pipelines.
As such, the ordering of blocks and attributes within the River configuration
file is not important; the language will consider all direct and indirect
dependencies between elements to determine their relationships.

## Attributes and Blocks

### Attributes
_Attributes_ are used to configure individual settings. They always take the
form of `<ATTRIBUTE_NAME> = <ATTRIBUTE_VALUE>`. They can appear either as
top-level elements or nested within blocks.

The following example  sets the `log_level` attribute to `"debug"`.

```river
log_level = "debug"
```

All `<ATTRIBUTE_NAME>`s must be valid River identifiers.

The `<ATTRIBUTE_VALUE>` can be either a constant value of a valid River
[type]({{< relref "./expressions/types_and_values.md" >}}) (eg. string,
boolean, number) or an [_expression_]({{< relref "./expressions/_index.md" >}})
that computes and continuously re-evaluates more complex attribute values.

### Blocks
_Blocks_ are used to configure components by grouping any number of attributes
or nested blocks using curly braces.

Blocks have a _name_ (the type of the component it defines), an
optional _label_ and a body that contains any number of arguments or nested
unlabeled blocks. Labels are used to disambiguate between multiple top-level
blocks of the same name.

```
// Pattern for creating an unlabeled block:
<BLOCK NAME> {
	// Block body can contain attributes and nested unlabeled blocks
	<IDENTIFIER> = <EXPRESSION> // Attribute

	<NESTED_BLOCK_NAME> {
		// Nested block body
	}
}

// Pattern for creating a labeled block:
<BLOCK NAME> "<BLOCK LABEL>" {
	// Block body can contain attributes and nested unlabeled blocks
	<IDENTIFIER> = <EXPRESSION> // Attribute

	<NESTED_BLOCK_NAME> {
		// Nested block body
	}
}
```

The `<BLOCK_NAME>` must correspond to a registered Flow [component]({{< relref "./components.md" >}}).

If the optional `<BLOCK_LABEL>` is set, it has to be a valid River identifier
wrapped in double quotes.

River does not allow multiple unlabelled blocks with the same name to co-exist,
so `<BLOCK_LABEL>`s are used to create multiple blocks and distinguish them.

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

## Comments
River configuration files support single-line `//` as well as `/* */` block
comments.

## Identifiers
River considers an identifier as valid if it consists of one or more UTF-8
letters, digits or underscores, but it doesn't start with a digit.

