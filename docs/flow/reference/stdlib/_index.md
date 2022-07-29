---
aliases:
- /docs/agent/latest/flow/reference/standard-library
title: Standard library
weight: 400
---

# Standard library

> Grafana Agent Flow is still growing, and its standard library isn't mature
> yet. If you have a request for for extensions to the standard library, please
> leave feedback on our dedicated [GitHub discussion for River
> feedback][feedback].

The standard library is a list of functions which can be used in expressions
when assigning values to attributes.

All standard library functions are idempotent: they will always return the same
output if given the same input.

* [`concat`]({{< relref "./concat.md" >}})
* [`env`]({{< relref "./env.md" >}})
* [`json_decode`]({{< relref "./json_decode.md" >}})

[feedback]: https://github.com/grafana/agent/discussions/1969
