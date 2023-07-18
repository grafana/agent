---
aliases:
- /docs/agent/shared/flow/reference/components/match-properties-block/
canonical: https://grafana.com/docs/grafana/agent/latest/shared/flow/reference/components/match-properties-block/
headless: true
---

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`match_type` | `string` | Controls how items in "services" and "span_names" arrays are interpreted. | | yes
`services` | `list(string)` | A list of items to match the service name against. | `[]` | no
`span_names` | `list(string)` | A list of items to match the span name against. | `[]` | no
`log_bodies` | `list(string)` | A list of strings that the LogRecord's body field must match against. | `[]` | no
`log_severity_texts` | `list(string)` | A list of strings that the LogRecord's severity text field must match against. | `[]` | no
`metric_names` | `list(string)` | A list of strings to match the metric name against. | `[]` | no
`span_kinds` | `list(string)` | A list of items to match the span kind against. | `[]` | no

`match_type` is required and must be set to either `"regexp"` or `"strict"`.

For `metric_names`, a match occurs if the metric name matches at least one item in the list.
For `span_kinds`, a match occurs if the span's span kind matches at least one item in the list.
