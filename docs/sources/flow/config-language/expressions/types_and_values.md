---
aliases:
- ../../configuration-language/expressions/types-and-values/
title: Types and values
weight: 100
---

# Types and values

## Types

River uses the following types for its values:

* `number`: Any numeric value, like `3` or `3.14`.
* `string`: A sequence of Unicode characters representing text, like `"Hello, world!"`.
* `bool`: A boolean value, either `true` or `false`.
* `array`: A sequence of values, like `[1, 2, 3]`. Elements within the
  list are indexed by whole numbers, starting with zero.
* `object`: A group of values which are identified by named labels, like
  `{ name = "John" }`.
* `function`: A value representing a routine which can be executed with
  arguments to compute another value, like `env("HOME")`. Functions take zero
  or more arguments as input and always return a single value as output.
* `null`: A type that has no value.

### Naming convention

In addition to the types above, [component reference][] documentation will use
the following conventions for referring to types:

* `any`: A value of any type.
* `map(T)`: an `object` where the value type is `T`. For example, `map(string)`
  is an object where all the values are strings. The key type of an object is
  always a string, or an identifier which is converted into a string.
* `list(T)`: an `array` where the value type is `T`. For example, `list(string`
  is an array where all the values are strings).
* `duration`: a `string` denoting a duration of time, such as `"1d"`, `"1h30m"`,
  `"10s"`. Valid units are `d` (for days), `h` (for hours), `m` (for minutes),
  `s` (for seconds), `ms` (for milliseconds), `ns` (for nanoseconds). Values of
  descending units can be combined to add their values together; `"1h30m"` is
  the same as `"90m"`.

[component reference]: {{< relref "../../reference/components/" >}}

## Numbers

River handles integers, unsigned integers and floating-point values as a single
'number' type which simplifies writing and reading River configuration files.

```river
3    == 3.00     // true
5.0  == (10 / 2) // true
1e+2 == 100      // true
2e-3 == 0.002    // true
```

## Strings

Strings are represented by sequences of Unicode characters surrounded by double
quotes `""`:

```river
"Hello, world!"
```

A `\` in a string starts an escape sequence to represent a special character.
The supported escape sequences are as follows:

| Sequence | Replacement |
| -------- | ----------- |
| `\\` | The `\` character `U+005C` |
| `\a` | The alert or bell character `U+0007` |
| `\b` | The backspace character `U+0008` |
| `\f` | The formfeed character `U+000C` |
| `\n` | The newline character `U+000A` |
| `\r` | The carriage return character `U+000D` |
| `\t` | The horizontal tab character `U+0009` |
| `\v` | The vertical tab character `U+000B` |
| `\'` | The `'` character `U+0027` |
| `\"` | The `"` character `U+0022`, which prevents terminating the string |
| `\NNN` | A literal byte (NNN is three octal digits) |
| `\xNN` | A literal byte (NN is two hexadecimal digits) |
| `\uNNNN` | A Unicode character from the basic multilingual plane (NNNN is four hexadecimal digits) |
| `\UNNNNNNNN` | A Unicode character from supplementary planes (NNNNNNNN is eight hexadecimal digits) |

## Bools

Bools are represented by the symbols `true` and `false`.

## Arrays

Array values are constructed by a sequence of comma separated values surrounded
by square brackets `[]`:

```river
[0, 1, 2, 3]
```

Values in array elements may be placed on separate lines for readability. A
comma after the final value must be present if the closing bracket `]`
is on a different line as the final value:

```river
[
  0,
  1,
  2,
]
```

## Objects

Object values are constructed by a sequence of comma separated key-value pairs
surrounded by curly braces `{}`:

```river
{
  first_name = "John",
  last_name  = "Doe",
}
```

A comma after the final key-value pair may be omitted if the closing curly
brace `}` is on the same line as the final pair:

```river
{ name = "John" }
```

If the key is not a valid identifier, it must be wrapped in double quotes like
a string:

```river
{
  "app.kubernetes.io/name"     = "mysql",
  "app.kubernetes.io/instance" = "mysql-abcxyz",
  namespace                    = "default",
}
```

> **NOTE**: Be careful not to confuse objects with blocks.
>
> An _object_ is a value assigned to an [Attribute][Attributes], where
> commas **must** be provided between key-value pairs on separate lines.
>
> A [Block][Blocks] is a named structural element composed of multiple attributes,
> where commas **must not** be provided between attributes.

[Attributes]: {{< relref "../syntax.md#Attributes" >}}
[Blocks]: {{< relref "../syntax.md#Blocks" >}}

## Functions

Function values cannot be constructed by users, but can be called from the
standard library or when exported by a component.

## Null

The null value is represented by the symbol `null`.

## Special Types

#### Secrets

A `secret` is a special type of string which is never displayed to the user.
`string` values may be assigned to an attribute expecting a `secret`, but never
the inverse; it is not possible to convert a secret to a string or assign a
secret to an attribute expecting a string.

#### Capsules

River has a special type called a `capsule`, which represents a category of
_internal_ types used by Flow. Each capsule type has a unique name and will be
represented to the user as `capsule("SOME_INTERNAL_NAME")`.
Capsule values cannot be constructed by the user, but can be used in
expressions as any other type. Capsules are not inter-compatible and an
attribute expecting a capsule can only be given a capsule of the same internal
type. That means, if an attribute expects a `capsule("prometheus.Receiver")`,
it can only be assigned a `capsule("prometheus.Receiver")` type. The specific
type of capsule expected is explicitly documented for any component which uses
or exports them.

In the following example, the `prometheus.remote_write` component exports a
`receiver`, which is a `capsule("prometheus.Receiver")` type. This can then be
used in the `forward_to` attribute of `prometheus.scrape`, which
expects an array of `capsule("prometheus.Receiver")`s:

```river
prometheus.remote_write "default" {
  endpoint {
    url = "http://localhost:9090/api/v1/write"
  }
}

prometheus.scrape "default" {
  targets    = [/* ... */]
  forward_to = [prometheus.remote_write.default.receiver]
}
```
