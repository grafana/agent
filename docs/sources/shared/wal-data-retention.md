---
aliases:
- /docs/agent/shared/wal-data-retention/
- /docs/grafana-cloud/agent/shared/wal-data-retention/
- /docs/grafana-cloud/monitor-infrastructure/agent/shared/wal-data-retention/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/shared/wal-data-retention/
canonical: https://grafana.com/docs/agent/latest/shared/wal-data-retention/
description: Shared content, information about data retention in the WAL
headless: true
---

The `prometheus.remote_write` component reuses the Prometheus Write-Ahead Log
(WAL) implementation for resiliency against network outages and to avoid the
prohibitively expensive in-memory buffering of all metric samples. The
component buffers the received metrics on-disk, in a WAL per configured
endpoint. The queue shards can use the WAL after the network outage is resolved
and flush the buffered metrics to the endpoints.

The WAL records metrics in 128MB-files called segments. To avoid having a WAL
that grows on-disk indefinitely, the component _truncates_ its segments on a
set interval.

On each truncation, the WAL deletes references to series that are no longer
present and also _checkpoints_ roughly the oldest two thirds of the segments
written to it. A checkpoint means that the WAL only keeps track of the unique
identifier for each existing metrics series, and deletes the actual samples
associated with it. If that data has not yet been pushed to remote_write, it is
lost.

This behavior dictates the data retention for the `prometheus.remote_write`
component. It also means that it is not possible to directly correlate data
retention directly to the data age itself, as the truncation logic works on
_segments_, not the samples themselves. This makes data retention less
predictable when the component receives a non-consistent rate of data.

The [wal block][] (in Flow mode) or the [metrics config][] (in static mode)
contain some configurable parameters that can be used to control the tradeoff
between memory usage, disk usage and data retention.

The `truncate_frequency` or `wal_truncate_frequency` parameter configures the
interval at which truncations happen. A lower value leads to reduced memory
usage, but also provides less resiliency to long outages.

When a WAL clean-up starts, the lowest successfully sent timestamp is used to
determine how much data is safe to remove from the WAL.
The `min_keepalive_time` or `min_wal_time` controls the minimum age of samples
considered for removal; no samples more recent than `min_keepalive_time` are
removed. The `max_keepalive_time` or `max_wal_time` controls the maximum age of
samples that can be kept in the WAL; samples older than
`max_keepalive_time` are forcibly removed.

### In cases of `remote_write` outages
When the remote write endpoint is unreachable over a period of time, the lowest
successfully sent timestamp is not updated. In that case, only the
`min_keepalive_time` and `max_keepalive_time` arguments control the age range
of data kept in the WAL.

If the remote write outage is longer than the `max_keepalive_time` parameter,
then the WAL is truncated and the oldest data is lost.

### In cases of intermittent `remote_write` outages
In case the remote write endpoint is intermittently reachable, the lowest
succesfully sent timestamp is updated whenever the connection is successful.
This updates the series' comparison with `min_keepalive_time` and triggers a
truncation on the `truncate_frequency` interval which will checkpoint
approximately two thirds of the data written since the previous truncation.

### In cases of falling behind
In case the queue shards cannot flush data quickly enough to keep
up-to-date with the most recent data buffered in the WAL, we say that the
component is 'falling behind'.
It's not unusual for the component to fall behind 2 or 3 scrape intervals
temporarily.
If the component falls behind more than 1/3rd of the data written since the
last truncate interval, it is possible for the truncate loop to checkpoint data
before they've had a chance to be pushed to the remote_write endpoint.

[wal block]: {{< relref "../flow/reference/components/prometheus.remote_write.md/#wal-block" >}}
[metrics config]: {{< relref "../static/configuration/metrics-config.md" >}}
