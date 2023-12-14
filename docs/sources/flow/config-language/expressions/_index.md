---
aliases:
- ../configuration-language/expressions/
- /docs/grafana-cloud/agent/flow/config-language/expressions/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/config-language/expressions/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/config-language/expressions/
- /docs/grafana-cloud/send-data/agent/flow/config-language/expressions/
canonical: https://grafana.com/docs/agent/latest/flow/config-language/expressions/
description: Learn about expressions
title: Expressions
weight: 400
---

# Expressions

Expressions represent or compute values you can assign to attributes within a configuration.

Basic expressions are literal values, like `"Hello, world!"` or `true`.
Expressions may also do things like [refer to values][] exported by components, perform arithmetic, or [call functions][].

You use expressions when you configure any component.
All component arguments have an underlying [type][].
River checks the expression type before assigning the result to an attribute.

{{% docs/reference %}}
[refer to values]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/config-language/expressions/referencing_exports"
[refer to values]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/config-language/expressions/referencing_exports"
[call functions]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/config-language/expressions/function_calls"
[call functions]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/config-language/expressions/function_calls"
[type]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/config-language/expressions/types_and_values"
[type]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/config-language/expressions/types_and_values"
{{% /docs/reference %}}
