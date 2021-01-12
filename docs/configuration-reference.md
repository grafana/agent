# Configuration Reference

The Grafana Cloud Agent is configured in a YAML file (usually called
`agent.yaml`) which contains information on the Grafana Cloud Agent and its
Prometheus instances.

* [server_config](#server_config)
* [prometheus_config](#prometheus_config)
* [loki_config](#loki_config)
* [tempo_config](#tempo_config)
* [integrations_config](#integrations_config)

## Variable Substitution

You can use environment variables in the configuration file to set values that
need to be configurable during deployment. To enable this functionality, you
must pass `-config.expand-env` as a command-line flag to the Agent.

To refer to an environment variable in the config file, use:

```
${VAR}
```

Where VAR is the name of the environment variable.

Each variable reference is replaced at startup by the value of the environment
variable. The replacement is case-sensitive and occurs before the YAML file is
parsed. References to undefined variables are replaced by empty strings unless
you specify a default value or custom error text.

To specify a default value, use:

```
${VAR:default_value}
```

Where default_value is the value to use if the environment variable is
undefined.

## File Format

To specify which configuration file to load, pass the `-config.file` flag at
the command line. The file is written in the [YAML
format](https://en.wikipedia.org/wiki/YAML), defined by the scheme below.
Brackets indicate that a parameter is optional. For non-list parameters the
value is set to the specified default.

Generic placeholders are defined as follows:

* `<boolean>`: a boolean that can take the values `true` or `false`
* `<int>`: any integer matching the regular expression `[1-9]+[0-9]*`
* `<duration>`: a duration matching the regular expression `[0-9]+(ns|us|µs|ms|[smh])`
* `<labelname>`: a string matching the regular expression `[a-zA-Z_][a-zA-Z0-9_]*`
* `<labelvalue>`: a string of unicode characters
* `<filename>`: a valid path relative to current working directory or an
    absolute path.
* `<host>`: a valid string consisting of a hostname or IP followed by an optional port number
* `<string>`: a regular string
* `<secret>`: a regular string that is a secret, such as a password

Support contents and default values of `agent.yaml`:

```yaml
# Configures the server of the Agent used to enable self-scraping.
[server: <server_config>]

# Configures Prometheus instances.
[prometheus: <prometheus_config>]

# Configures Loki log collection.
[loki: <loki_config>]

# Configures Tempo trace collection.
[tempo: <tempo_config>]

# Configures integrations for the Agent.
[integrations: <integrations_config>]
```

## server_config

The `server_config` block configures the Agent's behavior as an HTTP server,
gRPC server, and the log level for the whole process.

The Agent exposes an HTTP server for scraping its own metrics and gRPC for the
scraping service mode.

```yaml
# HTTP server listen host
[http_listen_address: <string>]

# HTTP server listen port
[http_listen_port: <int> | default = 80]

# gRPC server listen host. Unused.
[grpc_listen_address: <string>]

# gRPC server listen port. Unused.
[grpc_listen_port: <int> | default = 9095]

# Register instrumentation handlers (/metrics, etc.)
[register_instrumentation: <boolean> | default = true]

# Timeout for graceful shutdowns
[graceful_shutdown_timeout: <duration> | default = 30s]

# Read timeout for HTTP server
[http_server_read_timeout: <duration> | default = 30s]

# Write timeout for HTTP server
[http_server_write_timeout: <duration> | default = 30s]

# Idle timeout for HTTP server
[http_server_idle_timeout: <duration> | default = 120s]

# Max gRPC message size that can be received. Unused.
[grpc_server_max_recv_msg_size: <int> | default = 4194304]

# Max gRPC message size that can be sent. Unused.
[grpc_server_max_send_msg_size: <int> | default = 4194304]

# Limit on the number of concurrent streams for gRPC calls (0 = unlimited).
# Unused.
[grpc_server_max_concurrent_streams: <int> | default = 100]

# Log only messages with the given severity or above. Supported values [debug,
# info, warn, error]. This level affects logging for the whole application, not
# just the Agent's HTTP/gRPC server.
[log_level: <string> | default = "info"]

# Base path to server all API routes from (e.g., /v1/). Unused.
[http_path_prefix: <string>]
```

## prometheus_config

The `prometheus_config` block is used to define a collection of Prometheus
Instances, each of which is its own mini-Agent. Most users will only need to
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
[wal_directory: <string> | default = ""]

# Configures how long ago an abandoned (not associated with an instance) WAL
# may be written to before being eligible to be deleted
[wal_cleanup_age: <duration> | default = "12h"]

# Configures how often checks for abandoned WALs to be deleted are performed.
# A value of 0 disables periodic cleanup of abandoned WALs
[wal_cleanup_period: <duration> | default = "30m"]

# The list of Prometheus instances to launch with the agent.
configs:
  [- <prometheus_instance_config>]

# If an instance crashes abnormally, how long should we wait before trying
# to restart it. 0s disables the backoff period and restarts the agent
# immediately.
[instance_restart_backoff: <duration> | default = "5s"]

# How to spawn instances based on instance configs. Supported values: shared,
# distinct.
[instance_mode: <string> | default = "shared"]
```

### scraping_service_config

The `scraping_service` block configures the
[scraping service](./scraping-service.md), an operational
mode where configurations are stored centrally in a KV store and a cluster of
agents distribute discovery and scrape load between nodes.

```yaml
# Whether to enable scraping service mode. When enabled, local configs
# cannot be used.
[enabled: boolean | default = false]

# How often should the agent manually reshard. Useful for if KV change
# events are not sent by an agent.
[reshard_interval: <duration> | default = "1m"]

# The timeout for a reshard. Applies to a cluster-wide reshard (done when
# joining or leaving the cluster) and local reshards (done every
# reshard_interval). A timeout of 0 indicates no timeout.
[reshard_timeout: <duration> | default = "30s"]

# Configuration for the KV store to store metrics
kvstore: <kvstore_config>

# Configuration for how agents will cluster together.
lifecycler: <lifecycler_config>
```

### kvstore_config

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

### lifecycler_config

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

### scraping_service_client_config

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

### global_config

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
```

### prometheus_instance_config

The `prometheus_instance_config` block configures an individual Prometheus
instance, which acts as its own mini Prometheus agent.

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
# truncation will checkpoint old series, create a new segment for new samples,
# and remove old samples that have been succesfully sent via remote_write.
# If there are are multiple remote_write endpoints, the endpoint with the
# earliest timestamp is used for the cutoff period, ensuring that no data
# gets truncated until all remote_write configurations have been able to
# send the data.
[wal_truncate_frequency: <duration> | default = "1m"]

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

### scrape_config

A `scrape_config` section specifies a set of targets and parameters describing
how to scrape them. In the general case, one scrape configuration specifies a
single job. In advanced configurations, this may change.

Targets may be statically configured via the `static_configs` parameter or
dynamically discovered using one of the supported service-discovery mechanisms.

Additionally, `relabel_configs` allow advanced modifications to any target and
its labels before scraping.

```yaml
# The job name assigned to scraped metrics by default.
job_name: <job_name>

# How frequently to scrape targets from this job.
[ scrape_interval: <duration> | default = <global_config.scrape_interval> ]

# Per-scrape timeout when scraping this job.
[ scrape_timeout: <duration> | default = <global_config.scrape_timeout> ]

# The HTTP resource path on which to fetch metrics from targets.
[ metrics_path: <path> | default = /metrics ]

# honor_labels controls how Prometheus handles conflicts between labels that are
# already present in scraped data and labels that Prometheus would attach
# server-side ("job" and "instance" labels, manually configured target
# labels, and labels generated by service discovery implementations).
#
# If honor_labels is set to "true", label conflicts are resolved by keeping label
# values from the scraped data and ignoring the conflicting server-side labels.
#
# If honor_labels is set to "false", label conflicts are resolved by renaming
# conflicting labels in the scraped data to "exported_<original-label>" (for
# example "exported_instance", "exported_job") and then attaching server-side
# labels.
#
# Setting honor_labels to "true" is useful for use cases such as federation and
# scraping the Pushgateway, where all labels specified in the target should be
# preserved.
#
# Note that any globally configured "external_labels" are unaffected by this
# setting. In communication with external systems, they are always applied only
# when a time series does not have a given label yet and are ignored otherwise.
[ honor_labels: <boolean> | default = false ]

# honor_timestamps controls whether Prometheus respects the timestamps present
# in scraped data.
#
# If honor_timestamps is set to "true", the timestamps of the metrics exposed
# by the target will be used.
#
# If honor_timestamps is set to "false", the timestamps of the metrics exposed
# by the target will be ignored.
[ honor_timestamps: <boolean> | default = true ]

# Configures the protocol scheme used for requests.
[ scheme: <scheme> | default = http ]

# Optional HTTP URL parameters.
params:
  [ <string>: [<string>, ...] ]

# Sets the `Authorization` header on every scrape request with the
# configured username and password.
# password and password_file are mutually exclusive.
basic_auth:
  [ username: <string> ]
  [ password: <secret> ]
  [ password_file: <string> ]

# Sets the `Authorization` header on every scrape request with
# the configured bearer token. It is mutually exclusive with `bearer_token_file`.
[ bearer_token: <secret> ]

# Sets the `Authorization` header on every scrape request with the bearer token
# read from the configured file. It is mutually exclusive with `bearer_token`.
[ bearer_token_file: /path/to/bearer/token/file ]

# Configures the scrape request's TLS settings.
tls_config:
  [ <tls_config> ]

# Optional proxy URL.
[ proxy_url: <string> ]

# List of Azure service discovery configurations.
azure_sd_configs:
  [ - <azure_sd_config> ... ]

# List of Consul service discovery configurations.
consul_sd_configs:
  [ - <consul_sd_config> ... ]

# List of Digitalocean service discovery configurations.
digitalocean_sd_config:
  [ - <digitalocean_sd_config> ... ]

# List of Docker Swarm service discovery configurations.
dockerswarm_sd_configs:
  [ - <dockerswarm_sd_config> ... ]

# List of DNS service discovery configurations.
dns_sd_configs:
  [ - <dns_sd_config> ... ]

# List of EC2 service discovery configurations.
ec2_sd_configs:
  [ - <ec2_sd_config> ... ]

# List of OpenStack service discovery configurations.
openstack_sd_configs:
  [ - <openstack_sd_config> ... ]

# List of file service discovery configurations.
file_sd_configs:
  [ - <file_sd_config> ... ]

# List of GCE service discovery configurations.
gce_sd_configs:
  [ - <gce_sd_config> ... ]

# List of Hetzner service discovery configurations.
hetzner_sd_configs:
  [ - <hetzner_sd_config> ... ]

# List of Kubernetes service discovery configurations.
kubernetes_sd_configs:
  [ - <kubernetes_sd_config> ... ]

# List of Marathon service discovery configurations.
marathon_sd_configs:
  [ - <marathon_sd_config> ... ]

# List of AirBnB's Nerve service discovery configurations.
nerve_sd_configs:
  [ - <nerve_sd_config> ... ]

# List of Zookeeper Serverset service discovery configurations.
serverset_sd_configs:
  [ - <serverset_sd_config> ... ]

# List of Triton service discovery configurations.
triton_sd_configs:
  [ - <triton_sd_config> ... ]

# List of Eureka service discovery configurations.
eureka_sd_configs:
  [ - <eureka_sd_config> ... ]

# List of labeled statically configured targets for this job.
static_configs:
  [ - <static_config> ... ]

# List of target relabel configurations.
relabel_configs:
  [ - <relabel_config> ... ]

# List of metric relabel configurations.
metric_relabel_configs:
  [ - <relabel_config> ... ]

# Per-scrape limit on number of scraped samples that will be accepted.
# If more than this number of samples are present after metric relabelling
# the entire scrape will be treated as failed. 0 means no limit.
[ sample_limit: <int> | default = 0 ]

# Per-scrape config limit on number of unique targets that will be accepted.
# If more than this number of targets are present after target relabeling,
# Prometheus will mark the targets as failed without scraping them. 0 means
# no limit. This is an experimental feature of Prometheus and the behavior
# may change in the future.
[ target_limit: <int> | default = 0]
```

### azure_sd_config

Azure SD configurations allow retrieving scrape targets from Azure VMs.

The following meta labels are available on targets during relabeling:

* `__meta_azure_machine_id`: the machine ID
* `__meta_azure_machine_location`: the location the machine runs in
* `__meta_azure_machine_name`: the machine name
* `__meta_azure_machine_os_type`: the machine operating system
* `__meta_azure_machine_private_ip`: the machine's private IP
* `__meta_azure_machine_public_ip`: the machine's public IP if it exists
* `__meta_azure_machine_resource_group`: the machine's resource group
* `__meta_azure_machine_tag_<tagname>`: each tag value of the machine
* `__meta_azure_machine_scale_set`: the name of the scale set which the vm is part of (this value is only set if you are using a scale set)
* `__meta_azure_subscription_id`: the subscription ID
* `__meta_azure_tenant_id`: the tenant ID

See below for the configuration options for Azure discovery:

```yaml
# The information to access the Azure API.
# The Azure environment.
[ environment: <string> | default = AzurePublicCloud ]

# The authentication method, either OAuth or ManagedIdentity.
# See https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview
[ authentication_method: <string> | default = OAuth]
# The subscription ID. Always required.
subscription_id: <string>
# Optional tenant ID. Only required with authentication_method OAuth.
[ tenant_id: <string> ]
# Optional client ID. Only required with authentication_method OAuth.
[ client_id: <string> ]
# Optional client secret. Only required with authentication_method OAuth.
[ client_secret: <secret> ]

# Refresh interval to re-read the instance list.
[ refresh_interval: <duration> | default = 300s ]

# The port to scrape metrics from. If using the public IP address, this must
# instead be specified in the relabeling rule.
[ port: <int> | default = 80 ]
```

### consul_sd_config

Consul SD configurations allow retrieving scrape targets from Consul's Catalog API.

The following meta labels are available on targets during relabeling:

* `__meta_consul_address`: the address of the target
* `__meta_consul_dc`: the datacenter name for the target
* `__meta_consul_tagged_address_<key>`: each node tagged address key value of the target
* `__meta_consul_metadata_<key>`: each node metadata key value of the target
* `__meta_consul_node`: the node name defined for the target
* `__meta_consul_service_address`: the service address of the target
* `__meta_consul_service_id`: the service ID of the target
* `__meta_consul_service_metadata_<key>`: each service metadata key value of the target
* `__meta_consul_service_port`: the service port of the target
* `__meta_consul_service`: the name of the service the target belongs to
* `__meta_consul_tags`: the list of tags of the target joined by the tag separator

```yaml
# The information to access the Consul API. It is to be defined
# as the Consul documentation requires.
[ server: <host> | default = "localhost:8500" ]
[ token: <secret> ]
[ datacenter: <string> ]
[ scheme: <string> | default = "http" ]
[ username: <string> ]
[ password: <secret> ]

tls_config:
  [ <tls_config> ]

# A list of services for which targets are retrieved. If omitted, all services
# are scraped.
services:
  [ - <string> ]

# See https://www.consul.io/api/catalog.html#list-nodes-for-service to know more
# about the possible filters that can be used.

# An optional list of tags used to filter nodes for a given service. Services must contain all tags in the list.
tags:
  [ - <string> ]

# Node metadata used to filter nodes for a given service.
[ node_meta:
  [ <name>: <value> ... ] ]

# The string by which Consul tags are joined into the tag label.
[ tag_separator: <string> | default = , ]

# Allow stale Consul results (see https://www.consul.io/api/features/consistency.html). Will reduce load on Consul.
[ allow_stale: <bool> ]

# The time after which the provided names are refreshed.
# On large setup it might be a good idea to increase this value because the catalog will change all the time.
[ refresh_interval: <duration> | default = 30s ]
```

### digitalocean_sd_config

DigitalOcean SD configurations allow retrieving scrape targets from
[DigitalOcean's](https://www.digitalocean.com/) Droplets API. This service
discovery uses the public IPv4 address by default, by that can be changed with
relabelling, as demonstrated in the [Prometheus digitalocean-sd configuration
file](https://github.com/prometheus/prometheus/blob/release-2.22/documentation/examples/prometheus-digitalocean.yml).

The following meta labels are available on targets during relabeling:

- `__meta_digitalocean_droplet_id`: the id of the droplet
- `__meta_digitalocean_droplet_name`: the name of the droplet
- `__meta_digitalocean_image`: the image name of the droplet
- `__meta_digitalocean_private_ipv4`: the private IPv4 of the droplet
- `__meta_digitalocean_public_ipv4`: the public IPv4 of the droplet
- `__meta_digitalocean_public_ipv6`: the public IPv6 of the droplet
- `__meta_digitalocean_region`: the region of the droplet
- `__meta_digitalocean_size`: the size of the droplet
- `__meta_digitalocean_status`: the status of the droplet
- `__meta_digitalocean_features`: the comma-separated list of features of the droplet
- `__meta_digitalocean_tags`: the comma-separated list of tags of the droplet

```yaml
# Authentication information used to authenticate to the API server.
# Note that `basic_auth`, `bearer_token` and `bearer_token_file` options are
# mutually exclusive.
# password and password_file are mutually exclusive.

# Optional HTTP basic authentication information, not currently supported by DigitalOcean.
basic_auth:
  [ username: <string> ]
  [ password: <secret> ]
  [ password_file: <string> ]

# Optional bearer token authentication information.
[ bearer_token: <secret> ]

# Optional bearer token file authentication information.
[ bearer_token_file: <filename> ]

# Optional proxy URL.
[ proxy_url: <string> ]

# TLS configuration.
tls_config:
  [ <tls_config> ]

# The port to scrape metrics from.
[ port: <int> | default = 80 ]

# The time after which the droplets are refreshed.
[ refresh_interval: <duration> | default = 60s ]
```

### dockerswarm_sd_config

Docker Swarm SD configurations allow retrieving scrape targets from [Docker
Swarm](https://docs.docker.com/engine/swarm/) engine.

One of the following roles can be configured to discover targets:

#### services

The `services` role discovers all [Swarm
services](https://docs.docker.com/engine/swarm/key-concepts/#services-and-tasks)
and exposes their ports as targets. For each published port of a service, a
single target is generated. If a service has no published ports, a target per
service is created using the port parameter defined in the SD configuration.

Available meta labels:

- `__meta_dockerswarm_service_id`: the id of the service
- `__meta_dockerswarm_service_name`: the name of the service
- `__meta_dockerswarm_service_mode`: the mode of the service
- `__meta_dockerswarm_service_endpoint_port_name`: the name of the endpoint port, if available
- `__meta_dockerswarm_service_endpoint_port_publish_mode`: the publish mode of the endpoint port
- `__meta_dockerswarm_service_label_<labelname>`: each label of the service
- `__meta_dockerswarm_service_task_container_hostname`: the container hostname of the target, if available
- `__meta_dockerswarm_service_task_container_image`: the container image of the target
- `__meta_dockerswarm_service_updating_status`: the status of the service, if available
- `__meta_dockerswarm_network_id`: the ID of the network
- `__meta_dockerswarm_network_name`: the name of the network
- `__meta_dockerswarm_network_ingress`: whether the network is ingress
- `__meta_dockerswarm_network_internal`: whether the network is internal
- `__meta_dockerswarm_network_label_<labelname>`: each label of the network
- `__meta_dockerswarm_network_scope`: the scope of the network

#### tasks

The `tasks` role discovers all [Swarm
tasks](https://docs.docker.com/engine/swarm/key-concepts/#services-and-tasks)
and exposes their ports as targets. For each published port of a task, a single
target is generated. If a task has no published ports, a target per task is
created using the `port` parameter defined in the SD configuration.

Available meta labels:

- `__meta_dockerswarm_task_id`: the id of the task
- `__meta_dockerswarm_task_container_id`: the container id of the task
- `__meta_dockerswarm_task_desired_state`: the desired state of the task
- `__meta_dockerswarm_task_label_<labelname>`: each label of the task
- `__meta_dockerswarm_task_slot`: the slot of the task
- `__meta_dockerswarm_task_state`: the state of the task
- `__meta_dockerswarm_task_port_publish_mode`: the publish mode of the task port
- `__meta_dockerswarm_service_id`: the id of the service
- `__meta_dockerswarm_service_name`: the name of the service
- `__meta_dockerswarm_service_mode`: the mode of the service
- `__meta_dockerswarm_service_label_<labelname>`: each label of the service
- `__meta_dockerswarm_network_id`: the ID of the network
- `__meta_dockerswarm_network_name`: the name of the network
- `__meta_dockerswarm_network_ingress`: whether the network is ingress
- `__meta_dockerswarm_network_internal`: whether the network is internal
- `__meta_dockerswarm_network_label_<labelname>`: each label of the network
- `__meta_dockerswarm_network_label`: each label of the network
- `__meta_dockerswarm_network_scope`: the scope of the network
- `__meta_dockerswarm_node_id`: the ID of the node
- `__meta_dockerswarm_node_hostname`: the hostname of the node
- `__meta_dockerswarm_node_address`: the address of the node
- `__meta_dockerswarm_node_availability`: the availability of the node
- `__meta_dockerswarm_node_label_<labelname>`: each label of the node
- `__meta_dockerswarm_node_platform_architecture`: the architecture of the node
- `__meta_dockerswarm_node_platform_os`: the operating system of the node
- `__meta_dockerswarm_node_role`: the role of the node
- `__meta_dockerswarm_node_status`: the status of the node

The `__meta_dockerswarm_network_*` meta labels are not populated for ports which are published with mode=host.

#### nodes

The `nodes` role is used to discover [Swarm
nodes](https://docs.docker.com/engine/swarm/key-concepts/#nodes).

Available meta labels:

- `__meta_dockerswarm_node_address`: the address of the node
- `__meta_dockerswarm_node_availability`: the availability of the node
- `__meta_dockerswarm_node_engine_version`: the version of the node engine
- `__meta_dockerswarm_node_hostname`: the hostname of the node
- `__meta_dockerswarm_node_id`: the ID of the node
- `__meta_dockerswarm_node_label_<labelname>`: each label of the node
- `__meta_dockerswarm_node_manager_address`: the address of the manager component of the node
- `__meta_dockerswarm_node_manager_leader`: the leadership status of the manager component of the node (true or false)
- `__meta_dockerswarm_node_manager_reachability`: the reachability of the manager component of the node
- `__meta_dockerswarm_node_platform_architecture`: the architecture of the node
- `__meta_dockerswarm_node_platform_os`: the operating system of the node
- `__meta_dockerswarm_node_role`: the role of the node
- `__meta_dockerswarm_node_status`: the status of the node

See below for the configuration options for Docker Swarm discovery:

```yaml
# Address of the Docker daemon.
host: <string>

# Optional proxy URL.
[ proxy_url: <string> ]

# TLS configuration.
tls_config:
  [ <tls_config> ]

# Role of the targets to retrieve. Must be `services`, `tasks`, or `nodes`.
role: <string>

# The port to scrape metrics from, when `role` is nodes, and for discovered
# tasks and services that don't have published ports.
[ port: <int> | default = 80 ]

# The time after which the droplets are refreshed.
[ refresh_interval: <duration> | default = 60s ]

# Authentication information used to authenticate to the Docker daemon.
# Note that `basic_auth`, `bearer_token` and `bearer_token_file` options are
# mutually exclusive.
# password and password_file are mutually exclusive.

# Optional HTTP basic authentication information.
basic_auth:
  [ username: <string> ]
  [ password: <secret> ]
  [ password_file: <string> ]

# Optional bearer token authentication information.
[ bearer_token: <secret> ]

# Optional bearer token file authentication information.
[ bearer_token_file: <filename> ]
```

See [this example Prometheus configuration
file](https://github.com/prometheus/prometheus/blob/release-2.22/documentation/examples/prometheus-dockerswarm.yml)
for a detailed example of configuring Prometheus for Docker Swarm.

### dns_sd_config

A DNS-based service discovery configuration allows specifying a set of DNS
domain names which are periodically queried to discover a list of targets. The
DNS servers to be contacted are read from `/etc/resolv.conf`.

This service discovery method only supports basic DNS A, AAAA and SRV record
queries, but not the advanced DNS-SD approach specified in RFC6763.

During the relabeling phase, the meta label `__meta_dns_name` is available on
each target and is set to the record name that produced the discovered target.

```yaml
# A list of DNS domain names to be queried.
names:
  [ - <domain_name> ]

# The type of DNS query to perform.
[ type: <query_type> | default = 'SRV' ]

# The port number used if the query type is not SRV.
[ port: <number>]

# The time after which the provided names are refreshed.
[ refresh_interval: <duration> | default = 30s ]
```

### ec2_sd_config

EC2 SD configurations allow retrieving scrape targets from AWS EC2 instances.
The private IP address is used by default, but may be changed to the public IP
address with relabeling.

The following meta labels are available on targets during relabeling:

* `__meta_ec2_availability_zone`: the availability zone in which the instance is running
* `__meta_ec2_instance_id`: the EC2 instance ID
* `__meta_ec2_instance_state`: the state of the EC2 instance
* `__meta_ec2_instance_type`: the type of the EC2 instance
* `__meta_ec2_owner_id`: the ID of the AWS account that owns the EC2 instance
* `__meta_ec2_platform`: the Operating System platform, set to 'windows' on Windows servers, absent otherwise
* `__meta_ec2_primary_subnet_id`: the subnet ID of the primary network interface, if available
* `__meta_ec2_private_dns_name`: the private DNS name of the instance, if available
* `__meta_ec2_private_ip`: the private IP address of the instance, if present
* `__meta_ec2_public_dns_name`: the public DNS name of the instance, if available
* `__meta_ec2_public_ip`: the public IP address of the instance, if available
* `__meta_ec2_subnet_id`: comma separated list of subnets IDs in which the instance is running, if available
* `__meta_ec2_tag_<tagkey>`: each tag value of the instance
* `__meta_ec2_vpc_id`: the ID of the VPC in which the instance is running, if available

See below for the configuration options for EC2 discovery:

```yaml
# The information to access the EC2 API.

# The AWS region. If blank, the region from the instance metadata is used.
[ region: <string> ]

# Custom endpoint to be used.
[ endpoint: <string> ]

# The AWS API keys. If blank, the environment variables `AWS_ACCESS_KEY_ID`
# and `AWS_SECRET_ACCESS_KEY` are used.
[ access_key: <string> ]
[ secret_key: <secret> ]
# Named AWS profile used to connect to the API.
[ profile: <string> ]

# AWS Role ARN, an alternative to using AWS API keys.
[ role_arn: <string> ]

# Refresh interval to re-read the instance list.
[ refresh_interval: <duration> | default = 60s ]

# The port to scrape metrics from. If using the public IP address, this must
# instead be specified in the relabeling rule.
[ port: <int> | default = 80 ]

# Filters can be used optionally to filter the instance list by other criteria.
# Available filter criteria can be found here:
# https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html
# Filter API documentation: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Filter.html
filters:
  [ - name: <string>
      values: <string>, [...] ]
```

### openstack_sd_config

OpenStack SD configurations allow retrieving scrape targets from OpenStack Nova
instances.

One of the following `<openstack_role>` types can be configured to discover
targets:

#### hypervisor

The hypervisor role discovers one target per Nova hypervisor node. The target
address defaults to the host_ip attribute of the hypervisor.

The following meta labels are available on targets during relabeling:

* `__meta_openstack_hypervisor_host_ip`: the hypervisor node's IP address.
* `__meta_openstack_hypervisor_name`: the hypervisor node's name.
* `__meta_openstack_hypervisor_state`: the hypervisor node's state.
* `__meta_openstack_hypervisor_status`: the hypervisor node's status.
* `__meta_openstack_hypervisor_type`: the hypervisor node's type.

#### instance

The instance role discovers one target per network interface of Nova instance.
The target address defaults to the private IP address of the network interface.

The following meta labels are available on targets during relabeling:

* `__meta_openstack_address_pool`: the pool of the private IP.
* `__meta_openstack_instance_flavor`: the flavor of the OpenStack instance.
* `__meta_openstack_instance_id`: the OpenStack instance ID.
* `__meta_openstack_instance_name`: the OpenStack instance name.
* `__meta_openstack_instance_status`: the status of the OpenStack instance.
* `__meta_openstack_private_ip`: the private IP of the OpenStack instance.
* `__meta_openstack_project_id`: the project (tenant) owning this instance.
* `__meta_openstack_public_ip`: the public IP of the OpenStack instance.
* `__meta_openstack_tag_<tagkey>`: each tag value of the instance.
* `__meta_openstack_user_id`: the user account owning the tenant.

See below for the configuration options for OpenStack discovery:

```yaml
# The information to access the OpenStack API.

# The OpenStack role of entities that should be discovered.
role: <openstack_role>

# The OpenStack Region.
region: <string>

# identity_endpoint specifies the HTTP endpoint that is required to work with
# the Identity API of the appropriate version. While it's ultimately needed by
# all of the identity services, it will often be populated by a provider-level
# function.
[ identity_endpoint: <string> ]

# username is required if using Identity V2 API. Consult with your provider's
# control panel to discover your account's username. In Identity V3, either
# userid or a combination of username and domain_id or domain_name are needed.
[ username: <string> ]
[ userid: <string> ]

# password for the Identity V2 and V3 APIs. Consult with your provider's
# control panel to discover your account's preferred method of authentication.
[ password: <secret> ]

# At most one of domain_id and domain_name must be provided if using username
# with Identity V3. Otherwise, either are optional.
[ domain_name: <string> ]
[ domain_id: <string> ]

# The project_id and project_name fields are optional for the Identity V2 API.
# Some providers allow you to specify a project_name instead of the project_id.
# Some require both. Your provider's authentication policies will determine
# how these fields influence authentication.
[ project_name: <string> ]
[ project_id: <string> ]

# The application_credential_id or application_credential_name fields are
# required if using an application credential to authenticate. Some providers
# allow you to create an application credential to authenticate rather than a
# password.
[ application_credential_name: <string> ]
[ application_credential_id: <string> ]

# The application_credential_secret field is required if using an application
# credential to authenticate.
[ application_credential_secret: <secret> ]

# Whether the service discovery should list all instances for all projects.
# It is only relevant for the 'instance' role and usually requires admin permissions.
[ all_tenants: <boolean> | default: false ]

# Refresh interval to re-read the instance list.
[ refresh_interval: <duration> | default = 60s ]

# The port to scrape metrics from. If using the public IP address, this must
# instead be specified in the relabeling rule.
[ port: <int> | default = 80 ]

# TLS configuration.
tls_config:
  [ <tls_config> ]
```

### tls_config

A `tls_config` allows configuring TLS connections.

```yaml
# CA certificate to validate API server certificate with.
[ ca_file: <filename> ]

# Certificate and key files for client cert authentication to the server.
[ cert_file: <filename> ]
[ key_file: <filename> ]

# ServerName extension to indicate the name of the server.
# https://tools.ietf.org/html/rfc4366#section-3.1
[ server_name: <string> ]

# Disable validation of the server certificate.
[ insecure_skip_verify: <boolean> ]
```

### file_sd_config

File-based service discovery provides a more generic way to configure static
targets and serves as an interface to plug in custom service discovery
mechanisms.

It reads a set of files containing a list of zero or more `<static_config>`s.
Changes to all defined files are detected via disk watches and applied
immediately. Files may be provided in YAML or JSON format. Only changes
resulting in well-formed target groups are applied.

The JSON file must contain a list of static configs, using this format:

```json
[
  {
    "targets": [ "<host>", ... ],
    "labels": {
      "<labelname>": "<labelvalue>", ...
    }
  },
  ...
]
```

As a fallback, the file contents are also re-read periodically at the specified
refresh interval.

Each target has a meta label `__meta_filepath` during the relabeling phase. Its
value is set to the filepath from which the target was extracted.

There is a list of
[integrations](https://prometheus.io/docs/operating/integrations/#file-service-discovery)
with this discovery mechanism.

```yaml
# Patterns for files from which target groups are extracted.
files:
  [ - <filename_pattern> ... ]

# Refresh interval to re-read the files.
[ refresh_interval: <duration> | default = 5m ]
```

Where `<filename_pattern>` may be a path ending in `.json`, `.yml` or `.yaml`. The
last path segment may contain a single `*` that matches any character sequence,
e.g. `my/path/tg_*.json`.

### gce_sd_config

GCE SD configurations allow retrieving scrape targets from GCP GCE instances.
The private IP address is used by default, but may be changed to the public IP
address with relabeling.

The following meta labels are available on targets during relabeling:

* `__meta_gce_instance_id`: the numeric id of the instance
* `__meta_gce_instance_name`: the name of the instance
* `__meta_gce_label_<name>`: each GCE label of the instance
* `__meta_gce_machine_type`: full or partial URL of the machine type of the instance
* `__meta_gce_metadata_<name>`: each metadata item of the instance
* `__meta_gce_network`: the network URL of the instance
* `__meta_gce_private_ip`: the private IP address of the instance
* `__meta_gce_project`: the GCP project in which the instance is running
* `__meta_gce_public_ip`: the public IP address of the instance, if present
* `__meta_gce_subnetwork`: the subnetwork URL of the instance
* `__meta_gce_tags`: comma separated list of instance tags
* `__meta_gce_zone`: the GCE zone URL in which the instance is running

See below for the configuration options for GCE discovery:

```yaml
# The information to access the GCE API.

# The GCP Project
project: <string>

# The zone of the scrape targets. If you need multiple zones use multiple
# gce_sd_configs.
zone: <string>

# Filter can be used optionally to filter the instance list by other criteria
# Syntax of this filter string is described here in the filter query parameter section:
# https://cloud.google.com/compute/docs/reference/latest/instances/list
[ filter: <string> ]

# Refresh interval to re-read the instance list
[ refresh_interval: <duration> | default = 60s ]

# The port to scrape metrics from. If using the public IP address, this must
# instead be specified in the relabeling rule.
[ port: <int> | default = 80 ]

# The tag separator is used to separate the tags on concatenation
[ tag_separator: <string> | default = , ]
```

Credentials are discovered by the Google Cloud SDK default client by looking in the following places, preferring the first location found:

1. a JSON file specified by the `GOOGLE_APPLICATION_CREDENTIALS` environment
   variable
2. a JSON file in the well-known path
   `$HOME/.config/gcloud/application_default_credentials.json`
3. fetched from the GCE metadata server

If Prometheus is running within GCE, the service account associated with the
instance it is running on should have at least read-only permissions to the
compute resources. If running outside of GCE make sure to create an appropriate
service account and place the credential file in one of the expected locations.

### hetzner_sd_config

Hetzner SD configurations allow retrieving scrape targets from [Hetzner
Cloud](https://www.hetzner.com/) API and
[Robot](https://docs.hetzner.com/robot/) API. This service discovery uses the
public IPv4 address by default, but that can be changed with relabeling, as
demonstrated in the [Prometheus hetzner-sd configuration
file](https://github.com/prometheus/prometheus/blob/release-2.22/documentation/examples/prometheus-hetzner.yml).

The following meta labels are available on all targets during relabeling:

- `__meta_hetzner_server_id`: the ID of the server
- `__meta_hetzner_server_name`: the name of the server
- `__meta_hetzner_server_status`: the status of the server
- `__meta_hetzner_public_ipv4`: the public ipv4 address of the server
- `__meta_hetzner_public_ipv6_network`: the public ipv6 network (/64) of the server
- `__meta_hetzner_datacenter`: the datacenter of the server

The labels below are only available for targets with `role` set to `hcloud`:

- `__meta_hetzner_hcloud_image_name`: the image name of the server
- `__meta_hetzner_hcloud_image_description`: the description of the server image
- `__meta_hetzner_hcloud_image_os_flavor`: the OS flavor of the server image
- `__meta_hetzner_hcloud_image_os_version`: the OS version of the server image
- `__meta_hetzner_hcloud_image_description`: the description of the server image
- `__meta_hetzner_hcloud_datacenter_location`: the location of the server
- `__meta_hetzner_hcloud_datacenter_location_network_zone`: the network zone of the server
- `__meta_hetzner_hcloud_server_type`: the type of the server
- `__meta_hetzner_hcloud_cpu_cores`: the CPU cores count of the server
- `__meta_hetzner_hcloud_cpu_type`: the CPU type of the server (shared or dedicated)
- `__meta_hetzner_hcloud_memory_size_gb`: the amount of memory of the server (in GB)
- `__meta_hetzner_hcloud_disk_size_gb`: the disk size of the server (in GB)
- `__meta_hetzner_hcloud_private_ipv4_<networkname>`: the private ipv4 address of the server within a given network
- `__meta_hetzner_hcloud_label_<labelname>`: each label of the server

The labels below are only available for targets with `role` set to `robot`:

- `__meta_hetzner_robot_product`: the product of the server
- `__meta_hetzner_robot_cancelled`: the server cancellation status

```yaml
# The Hetzner role of entities that should be discovered.
# One of robot or hcloud.
role: <string>

# Authentication information used to authenticate to the API server.
# Note that `basic_auth`, `bearer_token` and `bearer_token_file` options are
# mutually exclusive.
# password and password_file are mutually exclusive.

# Optional HTTP basic authentication information, required when role is robot
# Role hcloud does not support basic auth.
basic_auth:
  [ username: <string> ]
  [ password: <secret> ]
  [ password_file: <string> ]

# Optional bearer token authentication information, required when role is hcloud
# Role robot does not support bearer token authentication.
[ bearer_token: <secret> ]

# Optional bearer token file authentication information.
[ bearer_token_file: <filename> ]

# Optional proxy URL.
[ proxy_url: <string> ]

# TLS configuration.
tls_config:
  [ <tls_config> ]

# The port to scrape metrics from.
[ port: <int> | default = 80 ]

# The time after which the servers are refreshed.
[ refresh_interval: <duration> | default = 60s ]
```

### kubernetes_sd_config

Kubernetes SD configurations allow retrieving scrape targets from Kubernetes'
REST API and always staying synchronized with the cluster state.

One of the following role types can be configured to discover targets:

#### node

The `node` role discovers one target per cluster node with the address defaulting
to the Kubelet's HTTP port. The target address defaults to the first existing
address of the Kubernetes node object in the address type order of
`NodeInternalIP`, `NodeExternalIP`, `NodeLegacyHostIP`, and `NodeHostName`.

Available meta labels:

* `__meta_kubernetes_node_name`: The name of the node object.
* `__meta_kubernetes_node_label_<labelname>`: Each label from the node object.
* `__meta_kubernetes_node_labelpresent_<labelname>`: true for each label from the node object.
* `__meta_kubernetes_node_annotation_<annotationname>`: Each annotation from the node object.
* `__meta_kubernetes_node_annotationpresent_<annotationname>`: true for each annotation from the node object.
* `__meta_kubernetes_node_address_<address_type>`: The first address for each node address type, if it exists.

In addition, the `instance` label for the node will be set to the node name as
retrieved from the API server.

#### service

The `service` role discovers a target for each service port for each service. This
is generally useful for blackbox monitoring of a service. The address will be
set to the Kubernetes DNS name of the service and respective service port.

Available meta labels:

* `__meta_kubernetes_namespace`: The namespace of the service object.
* `__meta_kubernetes_service_annotation_<annotationname>`: Each annotation from the service object.
* `__meta_kubernetes_service_annotationpresent_<annotationname>`: "true" for each annotation of the service object.
* `__meta_kubernetes_service_cluster_ip`: The cluster IP address of the service. (Does not apply to services of type ExternalName)
* `__meta_kubernetes_service_external_name`: The DNS name of the service. (Applies to services of type ExternalName)
* `__meta_kubernetes_service_label_<labelname>`: Each label from the service object.
* `__meta_kubernetes_service_labelpresent_<labelname>`: true for each label of the service object.
* `__meta_kubernetes_service_name`: The name of the service object.
* `__meta_kubernetes_service_port_name`: Name of the service port for the target.
* `__meta_kubernetes_service_port_protocol`: Protocol of the service port for the target.

#### pod

The `pod` role discovers all pods and exposes their containers as targets. For
each declared port of a container, a single target is generated. If a container
has no specified ports, a port-free target per container is created for manually
adding a port via relabeling.

Available meta labels:

* `__meta_kubernetes_namespace`: The namespace of the pod object.
* `__meta_kubernetes_pod_name`: The name of the pod object.
* `__meta_kubernetes_pod_ip`: The pod IP of the pod object.
* `__meta_kubernetes_pod_label_<labelname>`: Each label from the pod object.
* `__meta_kubernetes_pod_labelpresent_<labelname>`: truefor each label from the pod object.
* `__meta_kubernetes_pod_annotation_<annotationname>`: Each annotation from the pod object.
* `__meta_kubernetes_pod_annotationpresent_<annotationname>`: true for each annotation from the pod object.
* `__meta_kubernetes_pod_container_init`: true if the container is an InitContainer
* `__meta_kubernetes_pod_container_name`: Name of the container the target address points to.
* `__meta_kubernetes_pod_container_port_name`: Name of the container port.
* `__meta_kubernetes_pod_container_port_number`: Number of the container port.
* `__meta_kubernetes_pod_container_port_protocol`: Protocol of the container port.
* `__meta_kubernetes_pod_ready`: Set to true or false for the pod's ready state.
* `__meta_kubernetes_pod_phase`: Set to Pending, Running, Succeeded, Failed or Unknown in the lifecycle.
* `__meta_kubernetes_pod_node_name`: The name of the node the pod is scheduled onto.
* `__meta_kubernetes_pod_host_ip`: The current host IP of the pod object.
* `__meta_kubernetes_pod_uid`: The UID of the pod object.
* `__meta_kubernetes_pod_controller_kind`: Object kind of the pod controller.
* `__meta_kubernetes_pod_controller_name`: Name of the pod controller.

#### endpoints

The `endpoints` role discovers targets from listed endpoints of a service. For
each endpoint address one target is discovered per port. If the endpoint is
backed by a pod, all additional container ports of the pod, not bound to an
endpoint port, are discovered as targets as well.

Available meta labels:

* `__meta_kubernetes_namespace`: The namespace of the endpoints object.
* `__meta_kubernetes_endpoints_name`: The names of the endpoints object.
* For all targets discovered directly from the endpoints list (those not
  additionally inferred from underlying pods), the following labels are
  attached:
  * `__meta_kubernetes_endpoint_hostname`: Hostname of the endpoint.
  * `__meta_kubernetes_endpoint_node_name`: Name of the node hosting the endpoint.
  * `__meta_kubernetes_endpoint_ready`: Set to true or false for the endpoint's
    ready state.
  * `__meta_kubernetes_endpoint_port_name`: Name of the endpoint port.
  * `__meta_kubernetes_endpoint_port_protocol`: Protocol of the endpoint port.
  * `__meta_kubernetes_endpoint_address_target_kind`: Kind of the endpoint address target.
  * `__meta_kubernetes_endpoint_address_target_name`: Name of the endpoint address target.
* If the endpoints belong to a service, all labels of the `role: service`
  discovery are attached.
* For all targets backed by a pod, all labels of the role: `pod discovery` are
  attached.

#### ingress

The `ingress` role discovers a target for each path of each ingress. This is
generally useful for blackbox monitoring of an ingress. The address will be set
to the host specified in the ingress spec.

Available meta labels:

* `__meta_kubernetes_namespace`: The namespace of the ingress object.
* `__meta_kubernetes_ingress_name`: The name of the ingress object.
* `__meta_kubernetes_ingress_label_<labelname>`: Each label from the ingress object.
* `__meta_kubernetes_ingress_labelpresent_<labelname>`: true for each label from the ingress object.
* `__meta_kubernetes_ingress_annotation_<annotationname>`: Each annotation from the ingress object.
* `__meta_kubernetes_ingress_annotationpresent_<annotationname>`: true for each annotation from the ingress object.
* `__meta_kubernetes_ingress_scheme`: Protocol scheme of ingress, `https` if TLS config is set. Defaults to `http`.
* `__meta_kubernetes_ingress_path`: Path from ingress spec. Defaults to /.

See below for the configuration options for Kubernetes discovery:

```yaml
# The information to access the Kubernetes API.

# The API server addresses. If left empty, Prometheus is assumed to run inside
# of the cluster and will discover API servers automatically and use the pod's
# CA certificate and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/.
[ api_server: <host> ]

# The Kubernetes role of entities that should be discovered.
role: <role>

# Optional authentication information used to authenticate to the API server.
# Note that `basic_auth`, `bearer_token` and `bearer_token_file` options are
# mutually exclusive.
# password and password_file are mutually exclusive.

# Optional HTTP basic authentication information.
basic_auth:
  [ username: <string> ]
  [ password: <secret> ]
  [ password_file: <string> ]

# Optional bearer token authentication information.
[ bearer_token: <secret> ]

# Optional bearer token file authentication information.
[ bearer_token_file: <filename> ]

# Optional proxy URL.
[ proxy_url: <string> ]

# TLS configuration.
tls_config:
  [ <tls_config> ]

# Optional namespace discovery. If omitted, all namespaces are used.
namespaces:
  names:
    [ - <string> ]
```

Where `<role>` must be `endpoints`, `service`, `pod`, `node`, or `ingress`.

### marathon_sd_config

Marathon SD configurations allow retrieving scrape targets using the Marathon
REST API. Prometheus will periodically check the REST endpoint for currently
running tasks and create a target group for every app that has at least one
healthy task.

The following meta labels are available on targets during relabeling:

* `__meta_marathon_app`: the name of the app (with slashes replaced by dashes)
* `__meta_marathon_image`: the name of the Docker image used (if available)
* `__meta_marathon_task`: the ID of the Mesos task
* `__meta_marathon_app_label_<labelname>`: any Marathon labels attached to the app
* `__meta_marathon_port_definition_label_<labelname>`: the port definition labels
* `__meta_marathon_port_mapping_label_<labelname>`: the port mapping labels
* `__meta_marathon_port_index`: the port index number (e.g. 1 for PORT1)

See below for the configuration options for Marathon discovery:

```yaml
# List of URLs to be used to contact Marathon servers.
# You need to provide at least one server URL.
servers:
  - <string>

# Polling interval
[ refresh_interval: <duration> | default = 30s ]

# Optional authentication information for token-based authentication
# https://docs.mesosphere.com/1.11/security/ent/iam-api/#passing-an-authentication-token
# It is mutually exclusive with `auth_token_file` and other authentication mechanisms.
[ auth_token: <secret> ]

# Optional authentication information for token-based authentication
# https://docs.mesosphere.com/1.11/security/ent/iam-api/#passing-an-authentication-token
# It is mutually exclusive with `auth_token` and other authentication mechanisms.
[ auth_token_file: <filename> ]

# Sets the `Authorization` header on every request with the
# configured username and password.
# This is mutually exclusive with other authentication mechanisms.
# password and password_file are mutually exclusive.
basic_auth:
  [ username: <string> ]
  [ password: <string> ]
  [ password_file: <string> ]

# Sets the `Authorization` header on every request with
# the configured bearer token. It is mutually exclusive with `bearer_token_file` and other authentication mechanisms.
# NOTE: The current version of DC/OS marathon (v1.11.0) does not support standard Bearer token authentication. Use `auth_token` instead.
[ bearer_token: <string> ]

# Sets the `Authorization` header on every request with the bearer token
# read from the configured file. It is mutually exclusive with `bearer_token` and other authentication mechanisms.
# NOTE: The current version of DC/OS marathon (v1.11.0) does not support standard Bearer token authentication. Use `auth_token_file` instead.
[ bearer_token_file: /path/to/bearer/token/file ]

# TLS configuration for connecting to marathon servers
tls_config:
  [ <tls_config> ]

# Optional proxy URL.
[ proxy_url: <string> ]
```

By default every app listed in Marathon will be scraped by Prometheus. If not
all of your services provide Prometheus metrics, you can use a Marathon label
and Prometheus relabeling to control which instances will actually be scraped.

By default, all apps will show up as a single job in Prometheus (the one
specified in the configuration file), which can also be changed using
relabeling.

### nerve_sd_config

Nerve SD configurations allow retrieving scrape targets from AirBnB's Nerve
which are stored in Zookeeper.

The following meta labels are available on targets during relabeling:

* `__meta_nerve_path`: the full path to the endpoint node in Zookeeper
* `__meta_nerve_endpoint_host`: the host of the endpoint
* `__meta_nerve_endpoint_port`: the port of the endpoint
* `__meta_nerve_endpoint_name`: the name of the endpoint

```yaml
# The Zookeeper servers.
servers:
  - <host>
# Paths can point to a single service, or the root of a tree of services.
paths:
  - <string>
[ timeout: <duration> | default = 10s ]
```

### serverset_sd_config

Serverset SD configurations allow retrieving scrape targets from Serversets
which are stored in Zookeeper. Serversets are commonly used by Finagle and
Aurora.

The following meta labels are available on targets during relabeling:

* `__meta_serverset_path`: the full path to the serverset member node in Zookeeper
* `__meta_serverset_endpoint_host`: the host of the default endpoint
* `__meta_serverset_endpoint_port`: the port of the default endpoint
* `__meta_serverset_endpoint_host_<endpoint>`: the host of the given endpoint
* `__meta_serverset_endpoint_port_<endpoint>`: the port of the given endpoint
* `__meta_serverset_shard`: the shard number of the member
* `__meta_serverset_status`: the status of the member

```yaml
# The Zookeeper servers.
servers:
  - <host>
# Paths can point to a single serverset, or the root of a tree of serversets.
paths:
  - <string>
[ timeout: <duration> | default = 10s ]
```

### triton_sd_config

Triton SD configurations allow retrieving scrape targets from Container Monitor
discovery endpoints.

The following meta labels are available on targets during relabeling:

* `__meta_triton_groups`: the list of groups belonging to the target joined by a comma separator
* `__meta_triton_machine_alias`: the alias of the target container
* `__meta_triton_machine_brand`: the brand of the target container
* `__meta_triton_machine_id`: the UUID of the target container
* `__meta_triton_machine_image`: the target containers image type
* `__meta_triton_server_id`: the server UUID for the target container

```yaml
# The information to access the Triton discovery API.

# The account to use for discovering new target containers.
account: <string>

# The DNS suffix which should be applied to target containers.
dns_suffix: <string>

# The Triton discovery endpoint (e.g. 'cmon.us-east-3b.triton.zone'). This is
# often the same value as dns_suffix.
endpoint: <string>

# A list of groups for which targets are retrieved. If omitted, all containers
# available to the requesting account are scraped.
groups:
  [ - <string> ... ]

# The port to use for discovery and metric scraping.
[ port: <int> | default = 9163 ]

# The interval which should be used for refreshing target containers.
[ refresh_interval: <duration> | default = 60s ]

# The Triton discovery API version.
[ version: <int> | default = 1 ]

# TLS configuration.
tls_config:
  [ <tls_config> ]
```

### eureka_sd_config

Eureka SD configurations allow retrieving scrape targets using the
[Eureka](https://github.com/Netflix/eureka) REST API. Prometheus will
periodically check the REST endpoint and create a target for every app instance.

The following meta labels are available on targets during relabeling:

- `__meta_eureka_app_name`: the name of the app
- `__meta_eureka_app_instance_id`: the ID of the app instance
- `__meta_eureka_app_instance_hostname`: the hostname of the instance
- `__meta_eureka_app_instance_homepage_url`: the homepage url of the app instance
- `__meta_eureka_app_instance_statuspage_url`: the status page url of the app instance
- `__meta_eureka_app_instance_healthcheck_url`: the health check url of the app instance
- `__meta_eureka_app_instance_ip_addr`: the IP address of the app instance
- `__meta_eureka_app_instance_vip_address`: the VIP address of the app instance
- `__meta_eureka_app_instance_secure_vip_address`: the secure VIP address of the app instance
- `__meta_eureka_app_instance_status`: the status of the app instance
- `__meta_eureka_app_instance_port`: the port of the app instance
- `__meta_eureka_app_instance_port_enabled`: the port enabled of the app instance
- `__meta_eureka_app_instance_secure_port`: the secure port address of the app instance
- `__meta_eureka_app_instance_secure_port_enabled`: the secure port of the app instance
- `__meta_eureka_app_instance_country_id`: the country ID of the app instance
- `__meta_eureka_app_instance_metadata_<metadataname>`: app instance metadata
- `__meta_eureka_app_instance_datacenterinfo_name`: the datacenter name of the app instance
- `__meta_eureka_app_instance_datacenterinfo_<metadataname>`: the datacenter metadata

See below for the configuration options for Eureka discovery:

```yaml
# The URL to connect to the Eureka server.
server: <string>

# Sets the `Authorization` header on every request with the
# configured username and password.
# password and password_file are mutually exclusive.
basic_auth:
  [ username: <string> ]
  [ password: <secret> ]
  [ password_file: <string> ]

# Sets the `Authorization` header on every request with
# the configured bearer token. It is mutually exclusive with `bearer_token_file`.
[ bearer_token: <string> ]

# Sets the `Authorization` header on every request with the bearer token
# read from the configured file. It is mutually exclusive with `bearer_token`.
[ bearer_token_file: <filename> ]

# Configures the scrape request's TLS settings.
tls_config:
  [ <tls_config> ]

# Optional proxy URL.
[ proxy_url: <string> ]

# Refresh interval to re-read the app instance list.
[ refresh_interval: <duration> | default = 30s ]
```

See [the Prometheus eureka-sd configuration
file](https://github.com/prometheus/prometheus/blob/release-2.22/documentation/examples/prometheus-eureka.yml)
for a practical example on how to set up your Eureka app and your Prometheus
configuration.

### static_config

A `static_config` allows specifying a list of targets and a common label set for
them. It is the canonical way to specify static targets in a scrape
configuration.

```yaml
# The targets specified by the static config.
targets:
  [ - '<host>' ]

# Labels assigned to all metrics scraped from the targets.
labels:
  [ <labelname>: <labelvalue> ... ]
```

### relabel_config

Relabeling is a powerful tool to dynamically rewrite the label set of a target
before it gets scraped. Multiple relabeling steps can be configured per scrape
configuration. They are applied to the label set of each target in order of
their appearance in the configuration file.

Initially, aside from the configured per-target labels, a target's job label is
set to the `job_name` value of the respective scrape configuration. The
__address__ label is set to the `<host>:<port>` address of the target. After
relabeling, the `instance` label is set to the value of `__address__` by default if
it was not set during relabeling. The `__scheme__` and `__metrics_path__` labels are
set to the scheme and metrics path of the target respectively. The
`__param_<name>` label is set to the value of the first passed URL parameter
called `<name>`.

Additional labels prefixed with `__meta_` may be available during the relabeling
phase. They are set by the service discovery mechanism that provided the target
and vary between mechanisms.

Labels starting with `__` will be removed from the label set after target
relabeling is completed.

If a relabeling step needs to store a label value only temporarily (as the input
to a subsequent relabeling step), use the `__tmp` label name prefix. This prefix
is guaranteed to never be used by Prometheus itself.

```yaml
# The source labels select values from existing labels. Their content is concatenated
# using the configured separator and matched against the configured regular expression
# for the replace, keep, and drop actions.
[ source_labels: '[' <labelname> [, ...] ']' ]

# Separator placed between concatenated source label values.
[ separator: <string> | default = ; ]

# Label to which the resulting value is written in a replace action.
# It is mandatory for replace actions. Regex capture groups are available.
[ target_label: <labelname> ]

# Regular expression against which the extracted value is matched.
[ regex: <regex> | default = (.*) ]

# Modulus to take of the hash of the source label values.
[ modulus: <uint64> ]

# Replacement value against which a regex replace is performed if the
# regular expression matches. Regex capture groups are available.
[ replacement: <string> | default = $1 ]

# Action to perform based on regex matching.
[ action: <relabel_action> | default = replace ]
```

`<regex>` is any valid RE2 regular expression. It is required for the `replace`,
`keep`, `drop`, `labelmap`, `labeldrop` and `labelkeep` actions. The regex is
anchored on both ends. To un-anchor the regex, use `.*<regex>.*`.

`<relabel_action>` determines the relabeling action to take:

* `replace`: Match regex against the concatenated source_labels. Then, set
  target_label to replacement, with match group references (${1}, ${2}, ...) in
  replacement substituted by their value. If regex does not match, no
  replacement takes place.
* `keep`: Drop targets for which regex does not match the concatenated
  source_labels.
* `drop`: Drop targets for which regex matches the concatenated source_labels.
* `hashmod`: Set target_label to the modulus of a hash of the concatenated
  source_labels.
* `labelmap`: Match regex against all label names. Then copy the values of the
  matching labels to label names given by replacement with match group
  references (${1}, ${2}, ...) in replacement substituted by their value.
* `labeldrop`: Match regex against all label names. Any label that matches will
  be removed from the set of labels.
* `labelkeep`: Match regex against all label names. Any label that does not
  match will be removed from the set of labels.

Care must be taken with `labeldrop` and `labelkeep` to ensure that metrics are still uniquely labeled once the labels are removed.

### remote_write

`write_relabel_configs` is relabeling applied to samples before sending them to
the remote endpoint. Write relabeling is applied after external labels. This
could be used to limit which samples are sent.

```yaml
# The URL of the endpoint to send samples to.
url: <string>

# Timeout for requests to the remote write endpoint.
[ remote_timeout: <duration> | default = 30s ]

# List of remote write relabel configurations.
write_relabel_configs:
  [ - <relabel_config> ... ]

# Sets the `Authorization` header on every remote write request with the
# configured username and password.
# password and password_file are mutually exclusive.
basic_auth:
  [ username: <string> ]
  [ password: <string> ]
  [ password_file: <string> ]

# Sets the `Authorization` header on every remote write request with
# the configured bearer token. It is mutually exclusive with `bearer_token_file`.
[ bearer_token: <string> ]

# Sets the `Authorization` header on every remote write request with the bearer token
# read from the configured file. It is mutually exclusive with `bearer_token`.
[ bearer_token_file: /path/to/bearer/token/file ]

# Configures SigV4 request signing. The default credentials chain will be used,
# documented here:
#
# https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials
#
# When enabled, region must be supplied.
#
# This feature is currently exclusive to the Grafana Cloud Agent and is only
# currently available for remote_write of Prometheus metrics.
sigv4:
  # Enable SigV4 request signing. May not be enabled at the same time as
  # configuring basic auth or bearer_token/bearer_token_file.
  [ enabled: <boolean> | default = false ]

  # Region to use for signing the requests. Should be the region of the AMP
  # workspace. Must be provided when sig4.enabled is true.
  region: <string> 

# Configures the remote write request's TLS settings.
tls_config:
  [ <tls_config> ]

# Optional proxy URL.
[ proxy_url: <string> ]

# Configures the queue used to write to remote storage.
queue_config:
  # Number of samples to buffer per shard before we block reading of more
  # samples from the WAL. It is recommended to have enough capacity in each
  # shard to buffer several requests to keep throughput up while processing
  # occasional slow remote requests.
  [ capacity: <int> | default = 500 ]
  # Maximum number of shards, i.e. amount of concurrency.
  [ max_shards: <int> | default = 1000 ]
  # Minimum number of shards, i.e. amount of concurrency.
  [ min_shards: <int> | default = 1 ]
  # Maximum number of samples per send.
  [ max_samples_per_send: <int> | default = 100]
  # Maximum time a sample will wait in buffer.
  [ batch_send_deadline: <duration> | default = 5s ]
  # Initial retry delay. Gets doubled for every retry.
  [ min_backoff: <duration> | default = 30ms ]
  # Maximum retry delay.
  [ max_backoff: <duration> | default = 100ms ]

# Configures the sending of series metadata to remote storage.
# It is experimental and subject to change at any point.
metadata_config:
  # Whether metric metadata is sent to remote storage or not.
  [ send: <boolean> | default = true ]
  # How frequently metric metadata is sent to remote storage.
  [ send_interval: <duration> | default = 1m ]
```

### loki_config

The `loki_config` block configures how the Agent collects logs and sends them to a Loki push API endpoint. `loki_config` is identical to how Promtail is configured, except deprecated
fields have been removed and the server_config is not supported.

Please refer to the
[Promtail documentation](https://github.com/grafana/loki/tree/master/docs/sources/clients/promtail#client_config)
for the supported values for these fields.

```yaml
clients:
  - [<promtail.client_config>]

[positions: <promtail.position_config>]

scrape_configs:
  - [<promtail.scrape_config>]

[target_config: <promtail.target_config>]
```

### tempo_config

The `tempo_config` block configures how the Agent receives traces and sends them to Tempo.

```yaml
# Attributes options: https://github.com/open-telemetry/opentelemetry-collector/blob/1962d7cd2b371129394b0242b120835e44840192/processor/attributesprocessor
#  This field allows for the general manipulation of tags on spans that pass through this agent.  A common use may be to add an environment or cluster variable.
attributes: [attributes.config]

push_config:
  # host:port to send traces to
  endpoint: <string>

  # Controls whether or not TLS is required.  See https://godoc.org/google.golang.org/grpc#WithInsecure
  [ insecure: <boolean> | default = false ]

  # Disable validation of the server certificate. Only used when insecure is set
  # to false.
  [ insecure_skip_verify: <bool> | default = false ]

  # Sets the `Authorization` header on every trace push with the
  # configured username and password.
  # password and password_file are mutually exclusive.
  basic_auth:
    [ username: <string> ]
    [ password: <secret> ]
    [ password_file: <string> ]

  # Batch options are the same as: https://github.com/open-telemetry/opentelemetry-collector/blob/1962d7cd2b371129394b0242b120835e44840192/processor/batchprocessor
  [ batch: <batch.config> ]

  # sending_queue and retry_on_failure are the same as: https://github.com/open-telemetry/opentelemetry-collector/blob/1962d7cd2b371129394b0242b120835e44840192/exporter/otlpexporter
  [ sending_queue: <otlpexporter.sending_queue> ]
  [ retry_on_failure: <otlpexporter.retry_on_failure> ]

# Receiver configurations are mapped directly into the OpenTelmetry receivers block.
#   At least one receiver is required.
#   https://github.com/open-telemetry/opentelemetry-collector/blob/1962d7cd2b371129394b0242b120835e44840192/receiver/README.md
receivers:

# A list of prometheus scrape configs.  Targets discovered through these scrape configs have their __address__ matched against the ip on incoming spans.
# If a match is found then relabeling rules are applied.
scrape_configs:
  - [<scrape_config>]
```

### integrations_config

The `integrations_config` block configures how the Agent runs integrations that
scrape and send metrics without needing to run specific Prometheus exporters or
manually write `scrape_configs`:

```yaml
# Controls the Agent integration
agent:
  # Enables the Agent integration, allowing the Agent to automatically
  # collect and send metrics about itself.
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the agent integration will be run but not scraped and thus not
  # remote_written. Metrics for the integration will be exposed at
  # /integrations/agent/metrics and can be scraped by an external process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timtout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

# Controls the node_exporter integration
node_exporter: <node_exporter_config>

# Controls the process_exporter integration
process_exporter: <process_exporter_config>

# Controls the mysqld_exporter integration
mysqld_exporter: <mysqld_exporter_config>

# Controls the redis_exporter integration
redis_exporter: <redis_exporter_config>

# Controls the dnsmasq_exporter integration
dnsmasq_exporter: <dnsmasq_exporter_config>

# Controls the memcached_exporter integration
memcached_exporter: <memcached_exporter_config>

# Controls the postgres_exporter integration
postgres_exporter: <postgres_exporter_config>

# Controls the statsd_exporter integration
statsd_exporter: <statsd_exporter_config>

# Controls the consul_exporter integration
consul_exporter: <consul_exporter_config>

# Automatically collect metrics from enabled integrations. If disabled,
# integrations will be run but not scraped and thus not remote_written. Metrics
# for integrations will be exposed at /integrations/<integration_key>/metrics
# and can be scraped by an external process.
[scrape_integrations: <boolean> | default = true]

# When true, replaces the instance label with the hostname of the machine,
# rather than 127.0.0.1:<server.http_listen_port>. Useful when running multiple
# Agents with the same integrations and uniquely identifying where metrics are
# coming from.
#
# The value for the instance label can be replaced by providing custom
# relabel_configs for an integration. The overwritten instance label will be
# available when relabel_configs run.
[replace_instance_label: <boolean> | default = true]

# When true, adds an agent_hostname label to all samples coming from
# integrations. The value of the agent_hostname label will be the
# value of $HOSTNAME (if available) or the machine's hostname.
#
# DEPRECATED. May be removed in a future version. Rely on
# replace_instance_label instead, since it has better compatability
# with existing dashboards.
[use_hostname_label: <boolean> | default = true]

# Extra labels to add to all samples coming from integrations.
labels:
  { <string>: <string> }

# The period to wait before restarting an integration that exits with an
# error.
[integration_restart_backoff: <duration> | default = "5s"]

# A list of remote_write targets. Samples coming from integrations will be
# sent to all addresses specified here.
prometheus_remote_write:
  - [<remote_write>]
```

### node_exporter_config

The `node_exporter_config` block configures the `node_exporter` integration,
which is an embedded version of
[`node_exporter`](https://github.com/prometheus/node_exporter)
and allows for collecting metrics from the UNIX system that `node_exporter` is
running on. It provides a significant amount of collectors that are responsible
for monitoring various aspects of the host system.

Note that if running the Agent in a container, you will need to bind mount
folders from the host system so the integration can monitor them. You can use
the example below, making sure to replace `/path/to/config.yaml` with a path on
your host machine where an Agent configuration file is:

```
docker run \
  --net="host" \
  --pid="host" \
  --cap-add=SYS_TIME \
  -v "/:/host/root:ro,rslave" \
  -v "/sys:/host/sys:ro,rslave" \
  -v "/proc:/host/proc:ro,rslave" \
  -v /tmp/agent:/etc/agent \
  -v /path/to/config.yaml:/etc/agent-config/agent.yaml \
  grafana/agent:v0.9.1 \
  --config.file=/etc/agent-config/agent.yaml
```

Use this configuration file for testing out `node_exporter` support, replacing
the `prometheus_remote_write` settings with settings appropriate for you:

```yaml
server:
  log_level: info
  http_listen_port: 12345

prometheus:
  wal_directory: /tmp/agent
  global:
    scrape_interval: 15s

integrations:
  node_exporter:
    enabled: true
    rootfs_path: /host/root
    sysfs_path: /host/sys
    procfs_path: /host/proc
  prometheus_remote_write:
    - url: https://prometheus-us-central1.grafana.net/api/prom/push
      basic_auth:
        username: user-id
        password: api-token
```

For running on Kubernetes, ensure to set the equivalent mounts and capabilities
there as well:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: agent
spec:
  containers:
  - image: grafana/agent:v0.9.1
    name: agent
    args:
    - --config.file=/etc/agent-config/agent.yaml
    securityContext:
      capabilities:
        add: ["SYS_TIME"]
      priviliged: true
      runAsUser: 0
    volumeMounts:
    - name: rootfs
      mountPath: /host/root
      readOnly: true
    - name: sysfs
      mountPath: /host/sys
      readOnly: true
    - name: procfs
      mountPath: /host/proc
      readOnly: true
  volumes:
  - name: rootfs
    hostPath:
      path: /
  - name: sysfs
    hostPath:
      path: /sys
  - name: procfs
    hostPath:
      path: /proc
```

The manifest and Tanka configs provided by this repository do not have the
mounts or capabilities required for running this integration.

Some collectors only work on specific operating systems, documented in the
table below. Enabling a collector that is not supported by the operating system
the Agent is running on is a no-op.

| Name             | Description | OS | Enabled by default |
| ---------------- | ----------- | -- | ------------------ |
| arp              | Exposes ARP statistics from /proc/net/arp. | Linux | yes |
| bcache           | Exposes bcache statistics from /sys/fs/bcache. | Linux | yes |
| bonding          | Exposes the number of configured and active slaves of Linux bonding interfaces. | Linux | yes |
| boottime         | Exposes system boot time derived from the kern.boottime sysctl. | Darwin, Dragonfly, FreeBSD, NetBSD, OpenBSD, Solaris | yes |
| btrfs            | Exposes statistics on btrfs. | Linux | yes |
| buddyinfo        | Exposes statistics of memory fragments as reported by /proc/buddyinfo. | Linux | no |
| conntrack        | Shows conntrack statistics (does nothing if no /proc/sys/net/netfilter/ present). | Linux | yes |
| cpu              | Exposes CPU statistics. | Darwin, Dragonfly, FreeBSD, Linux, Solaris | yes |
| cpufreq          | Exposes CPU frequency statistics. | Linux, Solaris | yes |
| devstat          | Exposes device statistics. | Dragonfly, FreeBSD | no |
| diskstats        | Exposes disk I/O statistics. | Darwin, Linux, OpenBSD | yes |
| drbd             | Exposes Distributed Replicated Block Device statistics (to version 8.4). | Linux | no |
| edac             | Exposes error detection and correction statistics. | Linux | yes |
| entropy          | Exposes available entropy. | Linux | yes |
| exec             | Exposes execution statistics. | Dragonfly, FreeBSD | yes |
| filefd           | Exposes file descriptor statistics from /proc/sys/fs/file-nr. | Linux | yes |
| filesystem       | Exposes filesystem statistics, such as disk space used. | Darwin, Dragonfly, FreeBSD, Linux, OpenBSD | yes |
| hwmon            | Exposes hardware monitoring and sensor data from /sys/class/hwmon. | Linux | yes |
| infiniband       | Exposes network statistics specific to InfiniBand and Intel OmniPath configurations. | Linux | yes |
| interrupts       | Exposes detailed interrupts statistics. | Linux, OpenBSD | no |
| ipvs             | Exposes IPVS status from /proc/net/ip_vs and stats from /proc/net/ip_vs_stats. | Linux | yes |
| ksmd             | Exposes kernel and system statistics from /sys/kernel/mm/ksm. | Linux | no |
| loadavg          | Exposes load average. | Darwin, Dragonfly, FreeBSD, Linux, NetBSD, OpenBSD, Solaris | yes |
| logind           | Exposes session counts from logind. | Linux | no |
| mdadm            | Exposes statistics about devices in /proc/mdstat (does nothing if no /proc/mdstat present). | Linux | yes |
| meminfo          | Exposes memory statistics. | Darwin, Dragonfly, FreeBSD, Linux, OpenBSD | yes |
| meminfo_numa     | Exposes memory statistics from /proc/meminfo_numa. | Linux | no |
| mountstats       | Exposes filesystem statistics from /proc/self/mountstats. Exposes detailed NFS client statistics. | Linux | no |
| netclass         | Exposes network interface info from /sys/class/net. | Linux | yes |
| netdev           | Exposes network interface statistics such as bytes transferred. | Darwin, Dragonfly, FreeBSD, Linux, OpenBSD | yes |
| netstat          | Exposes network statistics from /proc/net/netstat. This is the same information as netstat -s. | Linux | yes |
| nfs              | Exposes NFS client statistics from /proc/net/rpc/nfs. This is the same information as nfsstat -c. | Linux | yes |
| nfsd             | Exposes NFS kernel server statistics from /proc/net/rpc/nfsd. This is the same information as nfsstat -s. | Linux | yes |
| ntp              | Exposes local NTP daemon helath to check time. | any | no |
| perf             | Exposes perf based metrics (Warning: Metrics are dependent on kernel configuration and settings). | Linux | no |
| powersupplyclass | Collects information on power supplies. | any | yes |
| pressure         | Exposes pressure stall statistics from /proc/pressure/. | Linux (kernel 4.20+ and/or CONFIG_PSI) | yes |
| processes        | Exposes aggregate process statistics from /proc. | Linux | no |
| qdisc            | Exposes queuing discipline statistics. | Linux | no |
| rapl             | Exposes various statistics from /sys/class/powercap. | Linux | yes |
| runit            | Exposes service status from runit. | any | no |
| schedstat        | Exposes task scheduler statistics from /proc/schedstat. | Linux | yes |
| sockstat         | Exposes various statistics from /proc/net/sockstat. | Linux | yes |
| softnet          | Exposes statistics from /proc/net/softnet_stat. | Linux | yes |
| stat             | Exposes various statistics from /proc/stat. This includes boot time, forks and interrupts. | Linux | yes |
| supervisord      | Exposes service status from supervisord. | any | no |
| systemd          | Exposes service and system status from systemd. | Linux | no |
| tcpstat          | Exposes TCP connection status information from /proc/net/tcp and /proc/net/tcp6. (Warning: the current version has potential performance issues in high load situations). | Linux | no |
| textfile         | Collects metrics from files in a directory matching the filename pattern *.prom. The files must be using the text format defined here: https://prometheus.io/docs/instrumenting/exposition_formats/ | any | yes |
| thermal_zone     | Exposes thermal zone & cooling device statistics from /sys/class/thermal. | Linux | yes |
| time             | Exposes the current system time. | any | yes |
| timex            | Exposes selected adjtimex(2) system call stats. | Linux | yes |
| udp_queues       | Exposes UDP total lengths of the rx_queue and tx_queue from /proc/net/udp and /proc/net/udp6. | Linux | yes |
| uname            | Exposes system information as provided by the uname system call. | Darwin, FreeBSD, Linux, OpenBSD | yes |
| vmstat           | Exposes statistics from /proc/vmstat. | Linux | yes |
| wifi             | Exposes WiFi device and station statistics. | Linux | no |
| xfs              | Exposes XFS runtime statistics. | Linux (kernel 4.4+) | yes |
| zfs              | Exposes ZFS performance statistics. | Linux, Solaris | yes |


```yaml
  # Enables the node_exporter integration, allowing the Agent to automatically
  # collect system metrics from the host UNIX system.
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the node_exporter integration will be run but not scraped and thus not remote-written. Metrics for the
  # integration will be exposed at /integrations/node_exporter/metrics and can
  # be scraped by an external process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timtout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <boolean> | default = false]

  # Optionally defines the the list of enabled-by-default collectors.
  # Anything not provided in the list below will be disabled by default,
  # but requires at least one element to be treated as defined.
  #
  # This is useful if you have a very explicit set of collectors you wish
  # to run.
  set_collectors:
    - [<string>]

  # Additional collectors to enable on top of the default set of enabled
  # collectors or on top of the list provided by set_collectors.
  #
  # This is useful if you have a few collectors you wish to run that are
  # not enabled by default, but do not want to explicitly provide an entire
  # list through set_collectors.
  enable_collectors:
    - [<string>]

  # Additional collectors to disable on top of the default set of disabled
  # collectors. Takes precedence over enable_collectors.

  # Additional collectors to disable from the set of enabled collectors.
  # Takes precedence over enabled_collectors.
  #
  # This is useful if you have a few collectors you do not want to run that
  # are enabled by default, but do not want to explicitly provide an entire
  # list through set_collectors.
  disable_collectors:
    - [<string>]

  # procfs mountpoint.
  [procfs_path: <string> | default = "/proc"]

  # sysfs mountpoint.
  [sysfs_path: <string> | default = "/sys"]

  # rootfs mountpoint. If running in docker, the root filesystem of the host
  # machine should be mounted and this value should be changed to the mount
  # directory.
  [rootfs_path: <string> | default = "/"]

  # Enable the cpu_info metric for the cpu collector.
  [enable_cpu_info_metric: <boolean> | default = true]

  # Regexmp of devices to ignore for diskstats.
  [diskstats_ignored_devices: <string> | default = "^(ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$"]

  # Regexp of mount points to ignore for filesystem collector.
  [filesystem_ignored_mount_points: <string> | default = "^/(dev|proc|sys|var/lib/docker/.+)($|/)"]

  # Regexp of filesystem types to ignore for filesystem collector.
  [filesystem_ignored_fs_types: <string> | default = "^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$"]

  # NTP server to use for ntp collector
  [ntp_server: <string> | default = "127.0.0.1"]

  # NTP protocol version
  [ntp_protocol_version: <int> | default = 4]

  # Certify that the server address is not a public ntp server.
  [ntp_server_is_local: <boolean> | default = false]

  # IP TTL to use wile sending NTP query.
  [ntp_ip_ttl: <int> | default = 1]

  # Max accumulated distance to the root.
  [ntp_max_distance: <duration> | default = "3466080us"]

  # Offset between local clock and local ntpd time to tolerate.
  [ntp_local_offset_tolerance: <duration> | default = "1ms"]

  # Regexp of net devices to ignore for netclass collector.
  [netclass_ignored_devices: <string> | default = "^$"]

  # Regexp of net devices to blacklist (mutually exclusive with whitelist)
  [netdev_device_blacklist: <string> | default = ""]

  # Regexp of net devices to whitelist (mutually exclusive with blacklist)
  [netdev_device_whitelist: <string> | default = ""]

  # Regexp of fields to return for netstat collector.
  [netstat_fields: <string> | default = "^(.*_(InErrors|InErrs)|Ip_Forwarding|Ip(6|Ext)_(InOctets|OutOctets)|Icmp6?_(InMsgs|OutMsgs)|TcpExt_(Listen.*|Syncookies.*|TCPSynRetrans)|Tcp_(ActiveOpens|InSegs|OutSegs|PassiveOpens|RetransSegs|CurrEstab)|Udp6?_(InDatagrams|OutDatagrams|NoPorts|RcvbufErrors|SndbufErrors))$"]

  # List of CPUs from which perf metrics should be collected.
  [perf_cpus: <string> | default = ""]

  # Regexp of power supplies to ignore for the powersupplyclass collector.
  [powersupply_ignored_supplies: <string> | default = "^$"]

  # Path to runit service directory.
  [runit_service_dir: <string> | default = "/etc/service"]

  # XML RPC endpoint for the supervisord collector.
  [supervisord_url: <string> | default = "http://localhost:9001/RPC2"]

  # Regexp of systemd units to whitelist. Units must both match whitelist
  # and not match blacklist to be included.
  [systemd_unit_whitelist: <string> | default = ".+"]

  # Regexp of systemd units to blacklist. Units must both match whitelist
  # and not match blacklist to be included.
  [systemd_unit_blacklist: <string> | default = ".+\\.(automount|device|mount|scope|slice)"]

  # Enables service unit tasks metrics unit_tasks_current and unit_tasks_max
  [systemd_enable_task_metrics: <boolean> | default = false]

  # Enables service unit metric service_restart_total
  [systemd_enable_restarts_metrics: <boolean> | default = false]

  # Enables service unit metric unit_start_time_seconds
  [systemd_enable_start_time_metrics: <boolean> | default = false]

  # Directory to read *.prom files from for the textfile collector.
  [textfile_directory: <string> | default = ""]

  # Regexp of fields to return for the vmstat collector.
  [vmstat_fields: <string> | default = "^(oom_kill|pgpg|pswp|pg.*fault).*"]
```

### process_exporter_config

The `process_exporter_config` block configures the `process_exporter` integration,
which is an embedded version of
[`process-exporter`](https://github.com/ncabatoff/process-exporter)
and allows for collection metrics based on the /proc filesystem on Linux
systems. Note that on non-Linux systems, enabling this exporter is a no-op.

Note that if running the Agent in a container, you will need to bind mount
folders from the host system so the integration can monitor them:

```
docker run \
  -v "/proc:/proc:ro" \
  -v /tmp/agent:/etc/agent \
  -v /path/to/config.yaml:/etc/agent-config/agent.yaml \
  grafana/agent:v0.9.1 \
  --config.file=/etc/agent-config/agent.yaml
```

Replace `/path/to/config.yaml` with the appropriate path on your host system
where an Agent config file can be found.

For running on Kubernetes, ensure to set the equivalent mounts and capabilities
there as well:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: agent
spec:
  containers:
  - image: grafana/agent:v0.9.1
    name: agent
    args:
    - --config.file=/etc/agent-config/agent.yaml
    volumeMounts:
    - name: procfs
      mountPath: /proc
      readOnly: true
  volumes:
  - name: procfs
    hostPath:
      path: /proc
```

The manifest and Tanka configs provided by this repository do not have the
mounts or capabilities required for running this integration.

An example config for `process_exporter_config` that tracks all processes is the
following:

```
enabled: true
process_names:
- name: "{{.Comm}}"
  cmdline:
  - '.+'
```

Full reference of options:

```yaml
  # Enables the process_exporter integration, allowing the Agent to automatically
  # collect system metrics from the host UNIX system.
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the process_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/process_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # procfs mountpoint.
  [procfs_path: <string> | default = "/proc"]

  # If a proc is tracked, track with it any children that aren't a part of their
  # own group.
  [track_children: <boolean> | default = true]

  # Report on per-threadname metrics as well.
  [track_threads: <boolean> | default = true]

  # Gather metrics from smaps file, which contains proportional resident memory
  # size.
  [gather_smaps: <boolean> | default = true]

  # Recheck process names on each scrape.
  [recheck_on_scrape: <boolean> | default = false]

  # A collection of matching rules to use for deciding which processes to
  # monitor. Each config can match multiple processes to be tracked as a single
  # process "group."
  process_names:
    [- <process_matcher_config>]
```

#### process_matcher_config

```yaml
# The name to use for identifying the process group name in the metric. By
# default, it uses the base path of the executable.
#
# The following template variables are available:
#
# - {{.Comm}}:      Basename of the original executable from /proc/<pid>/stat
# - {{.ExeBase}}:   Basename of the executable from argv[0]
# - {{.ExeFull}}:   Fully qualified path of the executable
# - {{.Username}}:  Username of the effective user
# - {{.Matches}}:   Map containing all regex capture groups resulting from
#                   matching a process with the cmdline rule group.
# - {{.PID}}:       PID of the process. Note that the PID is copied from the
#                   first executable found.
# - {{.StartTime}}: The start time of the process. This is useful when combined
#                   with PID as PIDS get reused over time.
[name: <string> | default = "{{.ExeBase}}"]

# A list of strings that match the base executable name for a process, truncated
# at 15 characters. It is derived from reading the second field of
# /proc/<pid>/stat minus the parens.
#
# If any of the strings match, the process will be tracked.
comm:
  [- <string>]

# A list of strings that match argv[0] for a process. If there are no slashes,
# only the basename of argv[0] needs to match. Otherwise the name must be an
# exact match. For example, "postgres" may match any postgres binary but
# "/usr/local/bin/postgres" can only match a postgres at that path exactly.
#
# If any of the strings match, the process will be tracked.
exe:
  [- <string>]

# A list of regular expressions applied to the argv of the process. Each
# regex here must match the corresponding argv for the process to be tracked.
# The first element that is matched is argv[1].
#
# Regex Captures are added to the .Matches map for use in the name.
cmdline:
  [- <string>]
```

### mysqld_exporter_config

The `mysqld_exporter_config` block configures the `mysqld_exporter` integration,
which is an embedded version of
[`mysqld_exporter`](https://github.com/prometheus/mysqld_exporter)
and allows for collection metrics from MySQL servers.

Note that currently, an Agent can only collect metrics from a single MySQL
server. If you want to collect metrics from multiple servers, run multiple
Agents and add labels using `relabel_configs` to differentiate between the MySQL
servers:

```yaml
mysqld_exporter:
  enabled: true
  data_source_name: root@(server-a:3306)/
  relabel_configs:
  - source_labels: [__address__]
    target_label: instance
    replacement: server-a
```

Full reference of options:

```yaml
  # Enables the mysqld_exporter integration, allowing the Agent to collect
  # metrics from a MySQL server.
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the mysqld_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/mysqld_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # Data Source Name specifies the MySQL server to connect to. This is REQUIRED
  # but may also be specified by the MYSQLD_EXPORTER_DATA_SOURCE_NAME
  # environment variable. If neither are set, the integration will fail to
  # start.
  #
  # The format of this is specified here: https://github.com/go-sql-driver/mysql#dsn-data-source-name
  #
  # A working example value for a server with no required password
  # authentication is: "root@(localhost:3306)/"
  data_source_name: <string>

  # A list of collector names to enable on top of the default set.
  enable_collectors:
    [ - <string> ]
  # A list of collector names to disable from the default set.
  disable_collectors:
    [ - <string> ]
  # A list of collectors to run. Fully overrides the default set.
  set_collectors:
    [ - <string> ]

  # Set a lock_wait_timeout on the connection to avoid long metadata locking.
  [lock_wait_timeout: <int> | default = 2]
  # Add a low_slow_filter to avoid slow query logging of scrapes. NOT supported
  # by Oracle MySQL.
  [log_slow_filter: <bool> | default = false]

  ## Collector-specific options

  # Minimum time a thread must be in each state to be counted.
  [info_schema_processlist_min_time: <int> | default = 0]
  # Enable collecting the number of processes by user.
  [info_schema_processlist_processes_by_user: <bool> | default = true]
  # Enable collecting the number of processes by host.
  [info_schema_processlist_processes_by_host: <bool> | default = true]
  # The list of databases to collect table stats for. * for all
  [info_schema_tables_databases: <string> | default = "*"]
  # Limit the number of events statements digests by response time.
  [perf_schema_eventsstatements_limit: <int> | default = 250]
  # Limit how old the 'last_seen' events statements can be, in seconds.
  [perf_schema_eventsstatements_time_limit: <int> | default = 86400]
  # Maximum length of the normalized statement text.
  [perf_schema_eventsstatements_digtext_text_limit: <int> | default = 120]
  # Regex file_name filter for performance_schema.file_summary_by_instance
  [perf_schema_file_instances_filter: <string> | default = ".*"]
  # Remove path prefix in performance_schema.file_summary_by_instance
  [perf_schema_file_instances_remove_prefix: <string> | default = "/var/lib/mysql"]
  # Database from where to collect heartbeat data.
  [heartbeat_database: <string> | default = "heartbeat"]
  # Table from where to collect heartbeat data.
  [heartbeat_table: <string> | default = "heartbeat"]
  # Use UTC for timestamps of the current server (`pt-heartbeat` is called with `--utc`)
  [heartbeat_utc: <bool> | default = false]
  # Enable collecting user privileges from mysql.user
  [mysql_user_privileges: <bool> | default = false]
```

The full list of collectors that are supported for `mysqld_exporter` is:

| Name                                             | Description | Enabled by default |
| ------------------------------------------------ | ----------- | ------------------ |
| auto_increment.columns                           | Collect auto_increment columns and max values from information_schema | no |
| binlog_size                                      | Collect the current size of all registered binlog files | no |
| engine_innodb_status                             | Collect from SHOW ENGINE INNODB STATUS | no |
| engine_tokudb_status                             | Collect from SHOW ENGINE TOKUDB STATUS | no |
| global_status                                    | Collect from SHOW GLOBAL STATUS | yes |
| global_variables                                 | Collect from SHOW GLOBAL VARIABLES | yes |
| heartbeat                                        | Collect from heartbeat | no |
| info_schema.clientstats                          | If running with userstat=1, enable to collect client statistics | no |
| info_schema.innodb_cmpmem                        | Collect metrics from information_schema.innodb_cmpmem | yes |
| info_schema.innodb_metrics                       | Collect metrics from information_schema.innodb_metrics | yes |
| info_schema.innodb_tablespaces                   | Collect metrics from information_schema.innodb_sys_tablespaces | no |
| info_schema.processlist                          | Collect current thread state counts from the information_schema.processlist | no |
| info_schema.query_response_time                  | Collect query response time distribution if query_response_time_stats is ON | yes |
| info_schema.replica_host                         | Collect metrics from information_schema.replica_host_status | no |
| info_schema.schemastats                          | If running with userstat=1, enable to collect schema statistics | no |
| info_schema.tables                               | Collect metrics from information_schema.tables | no |
| info_schema.tablestats                           | If running with userstat=1, enable to collect table statistics | no |
| info_schema.userstats                            | If running with userstat=1, enable to collect user statistics | no |
| mysql.user                                       | Collect data from mysql.user | no |
| perf_schema.eventsstatements                     | Collect metrics from performance_schema.events_statements_summary_by_digest | no |
| perf_schema.eventsstatementssum                  | Collect metrics of grand sums from performance_schema.events_statements_summary_by_digest | no |
| perf_schema.eventswaits                          | Collect metrics from performance_schema.events_waits_summary_global_by_event_name | no |
| perf_schema.file_events                          | Collect metrics from performance_schema.file_summary_by_event_name | no |
| perf_schema.file_instances                       | Collect metrics from performance_schema.file_summary_by_instance | no |
| perf_schema.indexiowaits                         | Collect metrics from performance_schema.table_io_waits_summary_by_index_usage | no |
| perf_schema.replication_applier_status_by_worker | Collect metrics from performance_schema.replication_applier_status_by_worker | no |
| perf_schema.replication_group_member_stats       | Collect metrics from performance_schema.replication_group_member_stats | no |
| perf_schema.replication_group_members            | Collect metrics from performance_schema.replication_group_members | no |
| perf_schema.tableiowaits                         | Collect metrics from performance_schema.table_io_waits_summary_by_table | no |
| perf_schema.tablelocks                           | Collect metrics from performance_schema.table_lock_waits_summary_by_table | no |
| slave_hosts                                      | Scrape information from 'SHOW SLAVE HOSTS' | no |
| slave_status                                     | Scrape information from SHOW SLAVE STATUS | yes |


### redis_exporter_config

The `redis_exporter_config` block configures the `redis_exporter` integration, which is an embedded version of [`redis_exporter`](https://github.com/oliver006/redis_exporter). This allows for the collection of metrics from Redis servers.

Note that currently, an Agent can only collect metrics from a single Redis server. If you want to collect metrics from multiple Redis servers, you can run multiple Agents and add labels using `relabel_configs` to differentiate between the Redis servers:

```yaml
redis_exporter:
  enabled: true
  redis_addr: "redis-2:6379"
  relabel_configs:
  - source_labels: [__address__]
    target_label: instance
    replacement: redis-2
```

Full reference of options:
```yaml
  # Enables the redis_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured redis address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the redis_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/redis_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  # exporter-specific configuration options

  # Address of the redis instance.
  redis_addr: <string>

  # User name to use for authentication (Redis ACL for Redis 6.0 and newer).
  [redis_user: <string>]

  # Password of the redis instance.
  [redis_password: <string>]

  # Path of a file containing a passord. If this is defined, it takes precedece
  # over redis_password.
  [redis_password_file: <string>]

  # Namespace for the metrics.
  [namespace: <string> | default = "redis"]

  # What to use for the CONFIG command.
  [config_command: <string> | default = "CONFIG"]

  # Comma separated list of key-patterns to export value and length/size, searched for with SCAN.
  [check_keys: <string>]

  # Comma separated list of LUA regex for grouping keys. When unset, no key
  # groups will be made.
  [check_key_groups: <string>]

  # Check key groups batch size hint for the underlying SCAN.
  [check_key_groups_batch_size: <int> | default = 10000]

  # The maximum number of distinct key groups with the most memory utilization
  # to present as distinct metrics per database. The leftover key groups will be
  # aggregated in the 'overflow' bucket.
  [max_distinct_key_groups: <int> | default = 100]

  # Comma separated list of single keys to export value and length/size.
  [check_single_keys: <string>]

  # Comma separated list of stream-patterns to export info about streams, groups and consumers, searched for with SCAN.
  [check_streams: <string>]

  # Comma separated list of single streams to export info about streams, groups and consumers.
  [check_single_streams: <string>]

  # Comma separated list of individual keys to export counts for.
  [count_keys: <string>]

  # Path to Lua Redis script for collecting extra metrics.
  [script_path: <string>]

  # Timeout for connection to Redis instance (in Golang duration format).
  [connection_timeout: <time.Duration> | default = "15s"]

  # Name of the client key file (including full path) if the server requires TLS client authentication.
  [tls_client_key_file: <string>]

  # Name of the client certificate file (including full path) if the server requires TLS client authentication.
  [tls_client_cert_file: <string>]

  # Name of the CA certificate file (including full path) if the server requires TLS client authentication.
  [tls_ca_cert_file: <string>]

  # Whether to set client name to redis_exporter.
  [set_client_name: <bool>]

  # Whether to scrape Tile38 specific metrics.
  [is_tile38: <bool>]

  # Whether to scrape Client List specific metrics.
  [export_client_list: <bool>]

  # Whether to include the client's port when exporting the client list. Note
  # that including this will increase the cardinality of all redis metrics.
  [export_client_port: <bool>]

  # Whether to also export go runtime metrics.
  [redis_metrics_only: <bool>]

  # Whether to ping the redis instance after connecting.
  [ping_on_connect: <bool>]

  # Whether to include system metrics like e.g. redis_total_system_memory_bytes.
  [incl_system_metrics: <bool>]

  # Whether to to skip TLS verification.
  [skip_tls_verification: <bool>]
```

### dnsmasq_exporter_config

The `dnsmasq_exporter_config` block configures the `dnsmasq_exporter` integration,
which is an embedded version of
[`dnsmasq_exporter`](https://github.com/google/dnsmasq_exporter). This allows for
the collection of metrics from dnsmasq servers.

Note that currently, an Agent can only collect metrics from a single dnsmasq
server. If you want to collect metrics from multiple servers, you can run
multiple Agents and add labels using `relabel_configs` to differentiate between
the servers:

```yaml
dnsmasq_exporter:
  enabled: true
  dnsmasq_address: dnsmasq-a:53
  relabel_configs:
  - source_labels: [__address__]
    target_label: instance
    replacement: dnsmasq-a
```

Full reference of options:

```yaml
  # Enables the dnsmasq_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured dnsmasq server address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the dnsmasq_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/dnsmasq_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  #
  # Exporter-specific configuration options
  #

  # Address of the dnsmasq server in host:port form.
  [dnsmasq_address: <string> | default = "localhost:53"]

  # Path to the dnsmasq leases file. If this file doesn't exist, scraping
  # dnsmasq # will fail with an warning log message.
  [leases_path: <string> | default = "/var/lib/misc/dnsmasq.leases"]
```

### memcached_exporter_config

The `memcached_exporter_config` block configures the `memcached_exporter`
integration, which is an embedded version of
[`memcached_exporter`](https://github.com/prometheus/memcached_exporter). This
allows for the collection of metrics from memcached servers.

Note that currently, an Agent can only collect metrics from a single memcached
server. If you want to collect metrics from multiple servers, you can run
multiple Agents and add labels using `relabel_configs` to differentiate between
the servers:

```yaml
memcached_exporter:
  enabled: true
  memcached_address: memcached-a:53
  relabel_configs:
  - source_labels: [__address__]
    target_label: instance
    replacement: memcached-a
```

Full reference of options:

```yaml
  # Enables the memcached_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured memcached server address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the memcached_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/memcached_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  #
  # Exporter-specific configuration options
  #

  # Address of the memcached server in host:port form.
  [memcached_address: <string> | default = "localhost:53"]

  # Timeout for connecting to memcached.
  [timeout: <duration> | default = "1s"]
```

### postgres_exporter_config

The `postgres_exporter_config` block configures the `postgres_exporter`
integration, which is an embedded version of
[`postgres_exporter`](https://github.com/wrouesnel/postgres_exporter). This
allows for the collection of metrics from Postgres servers.

Full reference of options:

```yaml
  # Enables the postgres_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured postgres server address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the postgres_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/postgres_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  #
  # Exporter-specific configuration options
  #

  # Data Source Names specifies the Postgres server(s) to connect to. This is
  # REQUIRED but may also be specified by the POSTGRES_EXPORTER_DATA_SOURCE_NAME
  # environment variable, where DSNs the environment variable are separated by
  # commas. If neither are set, the integration will fail to start.
  #
  # The format of this is specified here: https://pkg.go.dev/github.com/lib/pq#ParseURL
  #
  # A working example value for a server with a password is:
  # "postgresql://username:passwword@localhost:5432/database?sslmode=disable"
  #
  # Multiple DSNs may be provided here, allowing for scraping from multiple
  # servers.
  data_source_names:
  - <string>

  # Disables collection of metrics from pg_settings.
  [disable_settings_metrics: <boolean> | default = false]

  # Autodiscover databases to collect metrics from. If false, only collects
  # metrics from databases collected from data_source_names.
  [autodiscover_databases: <boolean> | default = false]

  # Excludes specific databases from being collected when autodiscover_databases
  # is true.
  exclude_databases:
  [ - <string> ]
```

### statsd_exporter_config

The `statsd_exporter_config` block configures the `statsd_exporter`
integration, which is an embedded version of
[`statsd_exporter`](https://github.com/prometheus/statsd_exporter). This allows
for the collection of statsd metrics and exposing them as Prometheus metrics.

Full reference of options:

```yaml
  # Enables the statsd_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured statsd server address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the statsd_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/statsd_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  #
  # Exporter-specific configuration options
  #

  # The UDP address on which to receive statsd metric lines. An empty string
  # will disable UDP collection.
  [listen_udp: <string> | default = ":9125"]

  # The TCP address on which to receive statsd metric lines. An empty string
  # will disable TCP collection.
  [listen_tcp: <string> | default = ":9125"]

  # The Unixgram socket path to receive statsd metric lines. An empty string
  # will disable unixgram collection.
  [listen_unixgram: <string> | default = ""]

  # The permission mode of the unixgram socket, when enabled.
  [unix_socket_mode: <string> | default = "755"]

  # An optional mapping config that can translate dot-separated StatsD metrics
  # into labeled Prometheus metrics. For full instructions on how to write this
  # object, see the official documentation from the statsd_exporter:
  #
  # https://github.com/prometheus/statsd_exporter#metric-mapping-and-configuration
  #
  # Note that a SIGHUP will not reload this config.
  [mapping_config: <statsd_exporter.mapping_config>]

  # Size (in bytes) of the operating system's transmit read buffer associated
  # with the UDP or unixgram connection. Please make sure the kernel parameters
  # net.core.rmem_max is set to a value greater than the value specified.
  [read_buffer: <int> | default = 0]

  # Maximum size of your metric mapping cache. Relies on least recently used
  # replacement policy if max size is reached.
  [cache_size: <int> | default = 1000]

  # Metric mapping cache type. Valid values are "lru" and "random".
  [cache_type: <string> | default = "lru"]

  # Size of internal queue for processing events.
  [event_queue_size: <int> | default = 10000]

  # Number of events to hold in queue before flushing.
  [event_flush_threshold: <int> | default = 1000]

  # Number of events to hold in queue before flushing.
  [event_flush_interval: <duration> | default = "200ms"]

  # Parse DogStatsd style tags.
  [parse_dogstatsd_tags: <bool> | default = true]

  # Parse InfluxDB style tags.
  [parse_influxdb_tags: <bool> | default = true]

  # Parse Librato style tags.
  [parse_librato_tags: <bool> | default = true]

  # Parse SignalFX style tags.
  [parse_signalfx_tags: <bool> | default = true]
```

### consul_exporter_config

The `consul_exporter_config` block configures the `consul_exporter`
integration, which is an embedded version of
[`consul_exporter`](https://github.com/prometheus/consul_exporter). This allows
for the collection of consul metrics and exposing them as Prometheus metrics.

Full reference of options:

```yaml
  # Enables the consul_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured consul server address
  [enabled: <boolean> | default = false]

  # Automatically collect metrics from this integration. If disabled,
  # the consul_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/consul_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  #
  # Exporter-specific configuration options
  #

  # Prefix from which to expose key/value pairs.
  [kv_prefix: <string> | default = ""]

  # Regex that determines which keys to expose.
  [kv_filter: <string> | default = ".*"]

  # Generate a health summary for each service instance. Needs n+1 queries to
  # collect all information.
  [generate_health_summary: <bool> | default = true]

  # HTTP API address of a Consul server or agent. Prefix with https:// to
  # connect using HTTPS.
  [server: <string> | default = "http://localhost:8500"]

  # Disable TLS host verification.
  [insecure_skip_verify: <bool> | default = false]

  # File path to a PEM-encoded certificate authority used to validate the
  # authenticity of a server certificate.
  [ca_file: <string> | default = ""]

  # File path to a PEM-encoded certificate used with the private key to verify
  # the exporter's authenticity.
  [cert_file: <string> | default = ""]

  # File path to a PEM-encoded private key used with the certificate to verify
  # the exporter's authenticity.
  [key_file: <string> | default = ""]

  # When provided, this overrides the hostname for the TLS certificate. It can
  # be used to ensure that the certificate name matches the hostname we declare.
  [server_name: <string> | default = ""]

  # Timeout on HTTP requests to the Consul API.
  [timeout: <duration> | default = "500ms"]

  # Limit the maximum number of concurrent requests to consul. 0 means no limit.
  [concurrent_request_limit: <int> | default = 0]

  # Allows any Consul server (non-leader) to service a read.
  [allow_stale: <bool> | default = true]

  # Forces the read to be fully consistent.
  [require_consistent: <bool> | default = false]
```
