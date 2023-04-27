---
title: Configuration blocks
weight: 200
---

# Configuration blocks

Configuration blocks are optional top-level blocks that can be used to
configure various parts of the Grafana Agent process. Each config block can
only be defined once.

Configuration blocks are _not_ components, so expressions that reference
components are invalid. Expressions that do not reference components (e.g.,
`env("LOG_LEVEL")`) are permitted.

{{< section >}}
