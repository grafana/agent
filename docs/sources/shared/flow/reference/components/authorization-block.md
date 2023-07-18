---
aliases:
- /docs/agent/shared/flow/reference/components/authorization-block/
canonical: https://grafana.com/docs/grafana/agent/latest/shared/flow/reference/components/authorization-block/
headless: true
---

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`type` | `string` | Authorization type, for example, "Bearer". | | no
`credentials` | `secret` | Secret value. | | no
`credentials_file` | `string` | File containing the secret value. | | no

`credential` and `credentials_file` are mutually exclusive and only one can be
provided inside of an `authorization` block.
