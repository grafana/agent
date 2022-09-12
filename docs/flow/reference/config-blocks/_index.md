---
aliases:
- /docs/agent/latest/flow/reference/config-blocks
title: Configuration blocks
weight: 200
---

# Configuration blocks

Configuration blocks are optional top-level blocks which can be used to
configure various parts of the Grafana Agent process.

Configuration blocks are _not_ components, so expressions which reference
components are invalid. Expressions which do not reference components (e.g.,
`env("LOG_LEVEL")`) are permitted.

{{< section >}}
