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

The `prometheus.remote_write` component uses a Write Ahead Log (WAL) to prevent
data loss during network outages. The component buffers the received metrics in
a WAL for each configured endpoint. The queue shards can use the WAL after the
network outage is resolved and flush the buffered metrics to the endpoints.

The WAL records metrics in 128 MB files called segments. To avoid having a WAL
that grows on-disk indefinitely, the component _truncates_ its segments on a
set interval.

On each truncation, the WAL deletes references to series that are no longer
present and also _checkpoints_ roughly the oldest two thirds of the segments
(rounded down to the nearest integer) written to it since the last truncation
period. A checkpoint means that the WAL only keeps track of the unique
identifier for each existing metrics series, and can no longer use the samples
for remote writing. If that data has not yet been pushed to the remote
endpoint, it is lost.

This behavior dictates the data retention for the `prometheus.remote_write`
component. It also means that it is impossible to directly correlate data
retention directly to the data age itself, as the truncation logic works on
_segments_, not the samples themselves. This makes data retention less
predictable when the component receives a non-consistent rate of data.

The [WAL block][] in Flow mode, or the [metrics config][] in Static mode
contain some configurable parameters that can be used to control the tradeoff
between memory usage, disk usage, and data retention.

The `truncate_frequency` or `wal_truncate_frequency` parameter configures the
interval at which truncations happen. A lower value leads to reduced memory
usage, but also provides less resiliency to long outages.

When a WAL clean-up starts, the most recently successfully sent timestamp is
used to determine how much data is safe to remove from the WAL.
The `min_keepalive_time` or `min_wal_time` controls the minimum age of samples
considered for removal. No samples more recent than `min_keepalive_time` are
removed. The `max_keepalive_time` or `max_wal_time` controls the maximum age of
samples that can be kept in the WAL. Samples older than
`max_keepalive_time` are forcibly removed.

### In cases of `remote_write` outages
When the remote write endpoint is unreachable over a period of time, the most
recent successfully sent timestamp is not updated. The
`min_keepalive_time` and `max_keepalive_time` arguments control the age range
of data kept in the WAL.

If the remote write outage is longer than the `max_keepalive_time` parameter,
then the WAL is truncated, and the oldest data is lost.

### In cases of intermittent `remote_write` outages
If the remote write endpoint is intermittently reachable, the most recent
successfully sent timestamp is updated whenever the connection is successful.
A successful connection updates the series' comparison with
`min_keepalive_time` and triggers a truncation on the next `truncate_frequency`
interval which checkpoints two thirds of the segments (rounded down to the
nearest integer) written since the previous truncation.

### In cases of falling behind
If the queue shards cannot flush data quickly enough to keep
up-to-date with the most recent data buffered in the WAL, we say that the
component is 'falling behind'.
It's not unusual for the component to temporarily fall behind 2 or 3 scrape intervals.
If the component falls behind more than one third of the data written since the
last truncate interval, it is possible for the truncate loop to checkpoint data
before being pushed to the remote_write endpoint.

[WAL block]: {{< relref "../flow/reference/components/prometheus.remote_write.md/#wal-block" >}}
[metrics config]: {{< relref "../static/configuration/metrics-config.md" >}}
