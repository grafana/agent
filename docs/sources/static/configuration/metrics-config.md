---
aliases:
- ../../configuration/metrics-config/
- ../../configuration/prometheus-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/metrics-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/metrics-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/metrics-config/
description: Learn about metrics_config
title: metrics_config
weight: 200
---

# metrics_config

The `metrics_config` block is used to define a collection of metrics
instances. Each instance defines a collection of Prometheus-compatible
scrape_configs and remote_write rules. Most users will only need to
define one instance.

```yaml
# Configures the optional scraping service to cluster agents.
[scraping_service: <scraping_service_config>]

# Configures the gRPC client used for agents to connect to other
# clustered agents.
[scraping_service_client: <scraping_service_client_config>]

# Configure values for all Prometheus instances.
[global: <global_config>]

# Configure the directory used by instances to store their WAL.
#
# The Grafana Agent assumes that all folders within wal_directory are managed by
# the agent itself. This means if you are using a PVC, you must point
# wal_directory to a subdirectory of the PVC mount.
[wal_directory: <string> | default = "data-agent/"]

# Configures how long ago an abandoned (not associated with an instance) WAL
# may be written to before being eligible to be deleted
[wal_cleanup_age: <duration> | default = "12h"]

# Configures how often checks for abandoned WALs to be deleted are performed.
# A value of 0 disables periodic cleanup of abandoned WALs
[wal_cleanup_period: <duration> | default = "30m"]

# Allows to disable HTTP Keep-Alives when scraping; the Agent will only use
# outgoing each connection for a single request.
[http_disable_keepalives: <boolean> | default = false]

# Allows to configure the maximum amount of time an idle Keep-Alive connection
# can remain idle before closing itself. Zero means no limit.
# The setting is ignored when `http_disable_keepalives` is enabled.
[http_idle_conn_timeout: <duration> | default = "5m"]

# The list of Prometheus instances to launch with the agent.
configs:
  [- <metrics_instance_config>]

# If an instance crashes abnormally, how long should we wait before trying
# to restart it. 0s disables the backoff period and restarts the agent
# immediately.
[instance_restart_backoff: <duration> | default = "5s"]

# How to spawn instances based on instance configs. Supported values: shared,
# distinct.
[instance_mode: <string> | default = "shared"]
```

## scraping_service_config

The `scraping_service` block configures the [scraping service][scrape], an operational
mode where configurations are stored centrally in a KV store and a cluster of
agents distributes discovery and scrape load between nodes.

```yaml
# Whether to enable scraping service mode. When enabled, local configs
# cannot be used.
[enabled: <boolean> | default = false]

# Note these next 3 configuration options are confusing. Due to backwards compatibility the naming
# is less than ideal.

# How often should the agent manually refresh the configuration. Useful for if KV change
# events are not sent by an agent.
[reshard_interval: <duration> | default = "1m"]

# The timeout for configuration refreshes. This can occur on cluster events or
# on the reshard interval. A timeout of 0 indicates no timeout.
[reshard_timeout: <duration> | default = "30s"]

# The timeout for a cluster reshard events. A timeout of 0 indicates no timeout.
[cluster_reshard_event_timeout: <duration> | default = "30s"]

# Configuration for the KV store to store configurations.
kvstore: <kvstore_config>

# When set, allows configs pushed to the KV store to specify configuration
# fields that can read secrets from files.
#
# This is disabled by default. When enabled, a malicious user can craft an
# instance config that reads arbitrary files on the machine the Agent runs
# on and sends its contents to a specically crafted remote_write endpoint.
#
# If enabled, ensure that no untrusted users have access to the Agent API.
[dangerous_allow_reading_files: <boolean>]

# Configuration for how agents will cluster together.
lifecycler: <lifecycler_config>
```

## kvstore_config

The `kvstore_config` block configures the KV store used as storage for
configurations in the scraping service mode.

```yaml
# Which underlying KV store to use. Can be either consul or etcd
[store: <string> | default = ""]

# Key prefix to store all configurations with. Must end in /.
[prefix: <string> | default = "configurations/"]

# Configuration for a Consul client. Only applies if store
# is "consul"
consul:
  # The hostname and port of Consul.
  [host: <string> | duration = "localhost:8500"]

  # The ACL Token used to interact with Consul.
  [acltoken: <string>]

  # The HTTP timeout when communicating with Consul
  [httpclienttimeout: <duration> | default = 20s]

  # Whether or not consistent reads to Consul are enabled.
  [consistentreads: <boolean> | default = true]

# Configuration for an ETCD v3 client. Only applies if
# store is "etcd"
etcd:
  # The ETCD endpoints to connect to.
  endpoints:
    - <string>

  # The Dial timeout for the ETCD connection.
  [dial_tmeout: <duration> | default = 10s]

  # The maximum number of retries to do for failed ops to ETCD.
  [max_retries: <int> | default = 10]
```

## lifecycler_config

The `lifecycler_config` block configures the lifecycler; the component that
Agents use to cluster together.

```yaml
# Configures the distributed hash ring storage.
ring:
  # KV store for getting and sending distributed hash ring updates.
  kvstore: <kvstore_config>

  # Specifies when other agents in the clsuter should be considered
  # unhealthy if they haven't sent a heartbeat within this duration.
  [heartbeat_timeout: <duration> | default = "1m"]

# Number of tokens to generate for the distributed hash ring.
[num_tokens: <int> | default = 128]

# How often agents should send a heartbeat to the distributed hash
# ring.
[heartbeat_period: <duration> | default = "5s"]

# How long to wait for tokens from other agents after generating
# a new set to resolve collisions. Useful only when using a gossip
# KV store.
[observe_period: <duration> | default = "0s"]

# Period to wait before joining the ring. 0s means to join immediately.
[join_after: <duration> | default = "0s"]

# Minimum duration to wait before marking the agent as ready to receive
# traffic. Used to work around race conditions for multiple agents exiting
# the distributed hash ring at the same time.
[min_ready_duration: <duration> | default = "1m"]

# Network interfaces to resolve addresses defined by other agents
# registered in distributed hash ring.
[interface_names: <string array> | default = ["eth0", "en0"]]

# Duration to sleep before exiting. Ensures that metrics get scraped
# before the process quits.
[final_sleep: <duration> | default = "30s"]

# File path to store tokens. If empty, tokens will not be stored during
# shutdown and will not be restored at startup.
[tokens_file_path: <string> | default = ""]

# Availability zone of the host the agent is running on. Default is an
# empty string which disables zone awareness for writes.
[availability_zone: <string> | default = ""]
```

## scraping_service_client_config

The `scraping_service_client_config` block configures how clustered Agents will
generate gRPC clients to connect to each other.

```yaml
grpc_client_config:
  # Maximum size in bytes the gRPC client will accept from the connected server.
  [max_recv_msg_size: <int> | default = 104857600]

  # Maximum size in bytes the gRPC client will sent to the connected server.
  [max_send_msg_size: <int> | default = 16777216]

  # Whether messages should be gzipped.
  [use_gzip_compression: <boolean> | default = false]

  # The rate limit for gRPC clients; 0 means no rate limit.
  [rate_limit: <float64> | default = 0]

  # gRPC burst allowed for rate limits.
  [rate_limit_burst: <int> | default = 0]

  # Controls if when a rate limit is hit whether the client should
  # retry the request.
  [backoff_on_ratelimits: <boolean> | default = false]

  # Configures the retry backoff when backoff_on_ratelimits is
  # true.
  backoff_config:
    # The minimum delay when backing off.
    [min_period: <duration> | default = "100ms"]

    # The maximum delay when backing off.
    [max_period: <duration> | default = "10s"]

    # The number of times to backoff and retry before failing.
    [max_retries: <int> | default = 10]
```

## global_config

The `global_config` block configures global values for all launched Prometheus
instances.

```yaml
# How frequently should Prometheus instances scrape.
[scrape_interval: duration | default = "1m"]

# How long to wait before timing out a scrape from a target.
[scrape_timeout: duration | default = "10s"]

# A list of static labels to add for all metrics.
external_labels:
  { <string>: <string> }

# Default set of remote_write endpoints. If an instance doesn't define any
# remote_writes, it will use this list.
remote_write:
  - [<remote_write>]
```

> **Note:** For more information on remote_write, refer to the [Prometheus documentation](https://prometheus.io/docs/prometheus/2.45/configuration/configuration/#remote_write).
>
> The following default values set by Grafana Agent Static Mode are different than the default set by Prometheus:
> - `remote_write`: `send_exemplars` default value is `true`
> - `remote_write`: `queue_config`: `retry_on_http_429` default value is `true`

## metrics_instance_config

The `metrics_instance_config` block configures an individual metrics
instance, which acts as its own mini Prometheus-compatible agent, though
without support for the TSDB.

```yaml
# Name of the instance. Must be present. Will be added as a label to agent
# metrics.
name: string

# Whether this agent instance should only scrape from targets running on the
# same machine as the agent process.
[host_filter: <boolean> | default = false]

# Relabel configs to apply against discovered targets. The relabeling is
# temporary and just used for filtering targets.
host_filter_relabel_configs:
  [ - <relabel_config> ... ]

# How frequently the WAL truncation process should run. Every iteration of
# the truncation will checkpoint old series and remove old samples. If data
# has not been sent within this window, some of it may be lost.
#
# The size of the WAL will increase with less frequent truncations. Making
# truncations more frequent reduces the size of the WAL but increases the
# chances of data loss when remote_write is failing for longer than the
# specified frequency.
[wal_truncate_frequency: <duration> | default = "60m"]

# The minimum amount of time that series and samples should exist in the WAL
# before being considered for deletion. The consumed disk space of the WAL will
# increase by making this value larger.
#
# Setting this value to 0s is valid, but may delete series before all
# remote_write shards have been able to write all data, and may cause errors on
# slower machines.
[min_wal_time: <duration> | default = "5m"]

# The maximum amount of time that series and samples may exist within the WAL
# before being considered for deletion. Series that have not received writes
# since this period will be removed, and all samples older than this period will
# be removed.
#
# This value is useful in long-running network outages, preventing the WAL from
# growing forever.
#
# Must be larger than min_wal_time.
[max_wal_time: <duration> | default = "4h"]

# Deadline for flushing data when a Prometheus instance shuts down
# before giving up and letting the shutdown proceed.
[remote_flush_deadline: <duration> | default = "1m"]

# When true, writes staleness markers to all active series to
# remote_write.
[write_stale_on_shutdown: <boolean> | default = false]

# A list of scrape configuration rules.
scrape_configs:
  - [<scrape_config>]

# A list of remote_write targets.
remote_write:
  - [<remote_write>]
```

> **Note:** More information on the following types can be found on the Prometheus
> website:
>
> * [`relabel_config`](https://prometheus.io/docs/prometheus/2.45/configuration/configuration/#relabel_config)
> * [`scrape_config`](https://prometheus.io/docs/prometheus/2.45/configuration/configuration/#scrape_config)
> * [`remote_write`](https://prometheus.io/docs/prometheus/2.45/configuration/configuration/#remote_write)

## Data retention

{{< docs/shared source="agent" lookup="/wal-data-retention.md" version="<AGENT_VERSION>" >}}

{{% docs/reference %}}
[scrape]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/scraping-service"
[scrape]: "/docs/grafana-cloud/ -> ./scraping-service"
{{% /docs/reference %}}
