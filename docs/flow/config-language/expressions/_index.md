---
aliases:
- /docs/agent/latest/flow/configuration-language/expressions
title: Expressions
weight: 400
---

# Expressions

Expressions represent or compute values that can be assigned to attributes
within a configuration.

Basic expressions are literal values, like `"Hello, world!"` or `true`.
Expressions may also do things like [refer to values][] exported by components,
perform arithmetic, or [call functions][] from the River standard library as
well as custom ones, dynamically created by components on the spot.

Expressions can be used when configuring any component. As all component
Arguments have an underlying [type][], River will type-check expressions before
assigning the resolved value to an attribute. 

The only exception to this rule are some special reserved blocks like
[`logging`][] which are used to configure the global behavior of Grafana Agent
Flow and do not support expressions but can only use literal values.

[refer to values]: {{< relref "./referencing_exports.md" >}}
[call functions]: {{< relref "./function_calls.md" >}}
[type]: {{< relref "./expressions/types_and_values.md" >}}
[`logging`]: {{< relref "../../controller-config/logging.md" >}}

