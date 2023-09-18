---
aliases:
- /docs/agent/shared/flow/reference/components/filter-field-block/
- /docs/grafana-cloud/agent/shared/flow/reference/components/filter-field-block/
- /docs/grafana-cloud/monitor-infrastructure/agent/shared/flow/reference/components/filter-field-block/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/shared/flow/reference/components/filter-field-block/
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/filter-field-block/
headless: true
---

The following attributes are supported:

Name | Type     | Description                                                                       | Default | Required
---- |----------|-----------------------------------------------------------------------------------|---------| --------
`key` | `string` | Key represents the key or name of the field or labels that a filter can apply on. |         | yes
`value` | `string` | Value represents the value associated with the key that a filter can apply on.    |         | yes
`op` | `string` | Op represents the filter operation to apply on the given Key: Value pair.         | `equals` | no

For `op` the following values are allowed:
* `equals`: The field value must be equal to the provided value.
* `not-equals`: The field value must not be equal to the provided value.
* `exists`: The field value must exist. (Only for `annotation` fields).
* `does-not-exist`: The field value must not exist. (Only for `annotation` fields).
