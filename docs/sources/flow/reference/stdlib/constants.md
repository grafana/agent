---
aliases:
- ../../configuration-language/standard-library/constants/
canonical: https://grafana.com/docs/grafana/agent/latest/flow/reference/stdlib/constants/
title: constants
---

# constants

The `constants` object exposes a list of constant values about the system
Grafana Agent is running on:

* `constants.hostname`: The hostname of the machine Grafana Agent is running
  on.
* `constants.os`: The operating system Grafana Agent is running on.
* `constants.arch`: The architecture of the system Grafana Agent is running on.

## Examples

```
> constants.hostname
"my-hostname"

> constants.os
"linux"

> constants.arch
"amd64"
```
