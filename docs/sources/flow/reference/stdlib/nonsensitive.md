---
aliases:
- ../../configuration-language/standard-library/nonsensitive/
canonical: https://grafana.com/docs/grafana/agent/latest/flow/reference/stdlib/nonsensitive/
title: nonsensitive
---

# nonsensitive

`nonsensitive` converts a [secret][] value back into a string.

> **WARNING**: Only use `nonsensitive` when you are positive that the value
> being converted back to a string is not a sensitive value.
>
> Strings resulting from calls to `nonsensitive` will be displayed in plaintext
> in the UI and internal API calls.

[secret]: {{< relref "../../config-language/expressions/types_and_values.md#secrets" >}}

## Examples

```
// Assuming `sensitive_value` is a secret:

> sensitive_value
(secret)
> nonsensitive(sensitive_value)
"Hello, world!"
```
