---
aliases:
- /docs/agent/latest/flow/configuration-language/expressions/types-and-values
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

### Special Types

#### Secrets

A `secret` is a special type of string which is never displayed to the user.
`string` values may be assigned to an attribute expecting a `secret`, but never
the inverse; it is not possible to convert a secret to a string or assign a
secret to an attribute expecting a string.

Secrets cannot be constructed by users, but are returned by components which
deal with sensitive values.

#### Capsules

River has a special type called a `capsule`, which represents an internal type
used by Grafana Agent for building telemetry pipelines. Capsule values can
never be constructed by the user, but can be used in expressions and assigned
to attributes. Capsules are represented to the user as
`capsule("SOME_INTENRAL_NAME")`. An attribute expecting a capsule can only be
given a capsule of the same internal type. The specific type of capsule
expected is explicitly documented for any component which uses or exports them.


The `metrics.remote_write` component below exports an attribute called
`forwarder`, which is a `capsule("metrics.Receiver")` type. This can then be
used in the `forward_to` attribute of `metrics.scrape`, which expects a list of
`capsule("metrics.Receiver")`:

```river
metrics.remote_write "default" {
  remote_write {
    url = "http://localhost:9090/api/v1/write"
  }
}

metrics.scrape "default" {
  targets = [/* ... */]

  // Have scraped metrics be sent to metrics.remote_write.default's receiver.
  forward_to = [metrics.remote_write.default.receiver]
}
```

## Numbers
River handles integers, unsigned integers and floating-point values as a single
'number' type which simplifies writing and reading River configuration files.

```
3   == 3.00     // true
5.0 == (10 / 2) // true
1e+2 == 100    // true
2e-3 == 0.002  // true
```

## Strings
Strings are represented by sequences of Unicode characters surrounded by double
quotes `""`:
```
"Hello, world!"
```

A `\\` in a string starts an escape sequence to represent a special character.
The supported escape sequences are as follows:

| Sequence | Replacement |
| -------- | ----------- |
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
```
[0, 1, 2, 3]
```

Values in array elements may be placed on separate lines for readability. A
comma after the final value must be present if the closing bracket `]`
is on a different line as the final value:
```
[
  0,
  1,
  2,
]
```

## Objects
Object values are constructed by a sequence of comma separated key-value pairs
surrounded by curly braces `{}`:
```
{
  first_name = "John",
  last_name = "Doe",
}
```
A comma after the final key-value pair may be omitted if the closing curly
brace `}` is on the same line as the final pair:
```
{ name = "John" }
```

## Functions
Function values cannot be constructed by users, but can be called from the
standard library or when exported by a component.

## Null
The null value is represented by the symbol `null`.

