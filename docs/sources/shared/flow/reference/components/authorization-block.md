---
aliases:
- /docs/agent/shared/flow/reference/components/authorization-block/
headless: true
---

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`type` | `string` | Authorization type, for example, "Bearer". | | no
`credential` | `secret` | Secret value. | | no
`credentials_file` | `string` | File containing the secret value. | | no

`credential` and `credentials_file` are mututally exclusive and only one can be
provided inside of an `authorization` block.
