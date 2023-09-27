---
aliases:
- /docs/agent/shared/flow/reference/components/basic-auth-block/
- /docs/grafana-cloud/agent/shared/flow/reference/components/basic-auth-block/
- /docs/grafana-cloud/monitor-infrastructure/agent/shared/flow/reference/components/basic-auth-block/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/shared/flow/reference/components/basic-auth-block/
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/basic-auth-block/
description: Shared content, basic auth block
headless: true
---

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`username` | `string` | Basic auth username. | | no
`password` | `secret` | Basic auth password. | | no
`password_file` | `string` | File containing the basic auth password. | | no

`password` and `password_file` are mutually exclusive and only one can be
provided inside of a `basic_auth` block.
