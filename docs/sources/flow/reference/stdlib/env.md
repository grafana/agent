---
aliases:
- ../../configuration-language/standard-library/env/
title: env
---

# env

The `env` function gets the value of an environment variable from the system
Grafana Agent is running on. If the environment variable does not exist, `env`
returns an empty string. An optional default value can be provided, if the environment
variable is not present.

## Examples

```
> env("HOME")
"/home/grafana-agent"

> env("DOES_NOT_EXIST")
""

> env("DOES_NOT_EXIST", "default")
"default"
```
