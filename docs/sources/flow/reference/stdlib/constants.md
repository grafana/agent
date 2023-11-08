---
aliases:
- ../../configuration-language/standard-library/constants/
- /docs/grafana-cloud/agent/flow/reference/stdlib/constants/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/stdlib/constants/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/stdlib/constants/
- /docs/grafana-cloud/send-data/agent/flow/reference/stdlib/constants/
canonical: https://grafana.com/docs/agent/latest/flow/reference/stdlib/constants/
description: Learn about constants
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
