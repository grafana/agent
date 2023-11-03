---
aliases:
- ../configuration-language/expressions/
- /docs/grafana-cloud/agent/flow/config-language/expressions/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/config-language/expressions/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/config-language/expressions/
canonical: https://grafana.com/docs/agent/latest/flow/config-language/expressions/
title: Expressions
description: Learn about expressions
weight: 400
---

# Expressions

Expressions represent or compute values that can be assigned to attributes
within a configuration.

Basic expressions are literal values, like `"Hello, world!"` or `true`.
Expressions may also do things like [refer to values][] exported by components,
perform arithmetic, or [call functions][].

Expressions can be used when configuring any component. As all component
arguments have an underlying [type][], River will type-check expressions before
assigning the result to an attribute.

[refer to values]: {{< relref "./referencing_exports.md" >}}
[call functions]: {{< relref "./function_calls.md" >}}
[type]: {{< relref "./types_and_values.md" >}}

