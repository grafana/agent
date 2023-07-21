---
aliases:
- /docs/agent/shared/flow/reference/components/otelcol-queue-block/
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/otelcol-queue-block/
headless: true
---

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `boolean` | Enables an in-memory buffer before sending data to the client. | `true` | no
`num_consumers` | `number` | Number of readers to send batches written to the queue in parallel. | `10` | no
`queue_size` | `number` | Maximum number of unwritten batches allowed in the queue at once. | `5000` | no

When `enabled` is `true`, data is first written to an in-memory buffer before
sending it to the configured server. Batches sent to the component's `input`
exported field are added to the buffer as long as the number of unsent batches
does not exceed the configured `queue_size`.

`queue_size` is used to determine how long an endpoint outage is tolerated for.
Assuming 100 requests/second, the default queue size `5000` provides about 50
seconds of outage tolerance. To calculate the correct value for `queue_size`,
multiply the average number of outgoing requests per second by the amount of
time in seconds that outages should be tolerated for.

The `num_consumers` argument controls how many readers read from the buffer and
send data in parallel. Larger values of `num_consumers` allow data to be sent
more quickly at the expense of increased network traffic.
