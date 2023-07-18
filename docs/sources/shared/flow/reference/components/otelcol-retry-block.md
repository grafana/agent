---
aliases:
- /docs/agent/shared/flow/reference/components/otelcol-retry-block/
canonical: https://grafana.com/docs/grafana/agent/latest/shared/flow/reference/components/otelcol-retry-block/
headless: true
---

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `boolean` | Enables retrying failed requests. | `true` | no
`initial_interval` | `duration` | Initial time to wait before retrying a failed request. | `"5s"` | no
`max_interval` | `duration` | Maximum time to wait between retries. | `"30s"` | no
`max_elapsed_time` | `duration` | Maximum amount of time to wait before discarding a failed batch. | `"5m"` | no

When `enabled` is `true`, failed batches are retried after a given interval.
The `initial_interval` argument specifies how long to wait before the first
retry attempt. If requests continue to fail, the time to wait before retrying
increases exponentially. The `max_interval` argument specifies the upper bound
of how long to wait between retries.

If a batch has not sent successfully, it is discarded after the time specified
by `max_elapsed_time` elapses. If `max_elapsed_time` is set to `"0s"`, failed
requests are retried forever until they succeed.
