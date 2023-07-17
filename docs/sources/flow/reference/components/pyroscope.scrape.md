---
title: pyroscope.scrape
labels:
  stage: beta
---

# pyroscope.scrape

{{< docs/shared lookup="flow/stability/beta.md" source="agent" >}}

`pyroscope.scrape` configures a [pprof] scraping job for a given set of
`targets`. The scraped performance profiles are forwarded to the list of receivers passed in
`forward_to`.

Multiple `pyroscope.scrape` components can be specified by giving them different labels.

## Usage

```river
pyroscope.scrape "LABEL" {
  targets    = TARGET_LIST
  forward_to = RECEIVER_LIST
}
```

## Arguments

The component configures and starts a new scrape job to scrape all of the
input targets. Multiple scrape jobs can be spawned for a single input target
when scraping multiple profile types.

The list of arguments that can be used to configure the block is
presented below.

The scrape job name defaults to the component's unique identifier.

Any omitted fields take on their default values. If conflicting
attributes are being passed (e.g., defining both a BearerToken and
BearerTokenFile or configuring both Basic Authorization and OAuth2 at the same
time), the component reports an error.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`targets`                  | `list(map(string))`     | List of targets to scrape. | | yes
`forward_to`               | `list(ProfilesReceiver)` | List of receivers to send scraped profiles to. | | yes
`job_name`                 | `string`   | The job name to override the job label with. | component name | no
`params`                   | `map(list(string))` | A set of query parameters with which the target is scraped. | | no
`scrape_interval`          | `duration` | How frequently to scrape the targets of this scrape config. | `"15s"` | no
`scrape_timeout`           | `duration` | The timeout for scraping targets of this config. | `"15s"` | no
`scheme`                   | `string`   | The URL scheme with which to fetch metrics from targets. | | no
`bearer_token`             | `secret`   | Bearer token to authenticate with. | | no
`bearer_token_file`        | `string`   | File containing a bearer token to authenticate with. | | no
`proxy_url`                | `string`   | HTTP proxy to proxy requests through. | | no
`follow_redirects`         | `bool`     | Whether redirects returned by the server should be followed. | `true` | no
`enable_http2`             | `bool`     | Whether HTTP2 is supported for requests. | `true` | no

 At most one of the following can be provided:
 - [`bearer_token` argument](#arguments).
 - [`bearer_token_file` argument](#arguments).
 - [`basic_auth` block][basic_auth].
 - [`authorization` block][authorization].
 - [`oauth2` block][oauth2].

 [arguments]: #arguments

## Blocks

The following blocks are supported inside the definition of `pyroscope.scrape`:

| Hierarchy                               | Block                    | Description                                                              | Required |
|-----------------------------------------|--------------------------|--------------------------------------------------------------------------|----------|
| basic_auth                              | [basic_auth][]           | Configure basic_auth for authenticating to targets.                      | no       |
| authorization                           | [authorization][]        | Configure generic authorization to targets.                              | no       |
| oauth2                                  | [oauth2][]               | Configure OAuth2 for authenticating to targets.                          | no       |
| oauth2 > tls_config                     | [tls_config][]           | Configure TLS settings for connecting to targets via OAuth2.             | no       |
| tls_config                              | [tls_config][]           | Configure TLS settings for connecting to targets.                        | no       |
| profiling_config                        | [profiling_config][]     | Configure profiling settings for the scrape job.                         | no       |
| profiling_config > profile.memory       | [profile.memory][]       | Collect memory profiles.                                                 | no       |
| profiling_config > profile.block        | [profile.block][]        | Collect profiles on blocks.                                              | no       |
| profiling_config > profile.goroutine    | [profile.goroutine][]    | Collect goroutine profiles.                                              | no       |
| profiling_config > profile.mutex        | [profile.mutex][]        | Collect mutex profiles.                                                  | no       |
| profiling_config > profile.process_cpu  | [profile.process_cpu][]  | Collect CPU profiles.                                                    | no       |
| profiling_config > profile.fgprof       | [profile.fgprof][]       | Collect [fgprof][] profiles.                                             | no       |
| profiling_config > profile.delta_memory | [profile.delta_memory][] | Collect [godeltaprof][] memory profiles.                                 | no       |
| profiling_config > profile.delta_mutex  | [profile.delta_mutex][]  | Collect [godeltaprof][] mutex profiles.                                  | no       |
| profiling_config > profile.delta_block  | [profile.delta_block][]  | Collect [godeltaprof][] block profiles.                                  | no       |
| profiling_config > profile.custom       | [profile.custom][]       | Collect custom profiles.                                                 | no       |
| clustering                              | [clustering][]           | Configure the component for when the Agent is running in clustered mode. | no       |

The `>` symbol indicates deeper levels of nesting. For example,
`oauth2 > tls_config` refers to a `tls_config` block defined inside
an `oauth2` block.

[basic_auth]: #basic_auth-block
[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block
[profiling_config]: #profiling_config-block
[profile.memory]: #profile.memory-block
[profile.block]: #profile.block-block
[profile.goroutine]: #profile.goroutine-block
[profile.mutex]: #profile.mutex-block
[profile.process_cpu]: #profile.process_cpu-block
[profile.fgprof]: #profile.fgprof-block
[profile.delta_memory]: #profile.delta_memory-block
[profile.delta_mutex]: #profile.delta_mutex-block
[profile.delta_block]: #profile.delta_block-block
[profile.custom]: #profile.custom-block
[pprof]: https://github.com/google/pprof/blob/main/doc/README.md
[clustering]: #clustering-beta

[fgprof]: https://github.com/felixge/fgprof
[godeltaprof]: https://github.com/grafana/godeltaprof

### basic_auth block

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" >}}

### authorization block

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" >}}

### oauth2 block

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" >}}

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" >}}

### profiling_config block

The `profiling_config` block configures the profiling settings when scraping
targets.

The block contains the following attributes:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`path_prefix` | `string` | The path prefix to use when scraping targets. | | no

### profile.memory block

The `profile.memory` block collects profiles on memory consumption.

It accepts the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `boolean` | Enable this profile type to be scraped. | `true` | no
`path` | `string` | The path to the profile type on the target. | `"/debug/pprof/memory"` | no
`delta` | `boolean` | Whether to scrape the profile as a delta. | `false` | no

When the `delta` argument is `true`, a `seconds` query parameter is
automatically added to requests.

### profile.block block

The `profile.block` block collects profiles on process blocking.

It accepts the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `boolean` | Enable this profile type to be scraped. | `true` | no
`path` | `string` | The path to the profile type on the target. | `"/debug/pprof/block"` | no
`delta` | `boolean` | Whether to scrape the profile as a delta. | `false` | no

When the `delta` argument is `true`, a `seconds` query parameter is
automatically added to requests.

### profile.goroutine block

The `profile.goroutine` block collects profiles on the number of goroutines.

It accepts the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `boolean` | Enable this profile type to be scraped. | `true` | no
`path` | `string` | The path to the profile type on the target. | `"/debug/pprof/goroutine"` | no
`delta` | `boolean` | Whether to scrape the profile as a delta. | `false` | no

When the `delta` argument is `true`, a `seconds` query parameter is
automatically added to requests.

### profile.mutex block

The `profile.mutex` block collects profiles on mutexes.

It accepts the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `boolean` | Enable this profile type to be scraped. | `true` | no
`path` | `string` | The path to the profile type on the target. | `"/debug/pprof/mutex"` | no
`delta` | `boolean` | Whether to scrape the profile as a delta. | `false` | no

When the `delta` argument is `true`, a `seconds` query parameter is
automatically added to requests.

### profile.process_cpu block

The `profile.process_cpu` block collects profiles on CPU consumption for the
process.

It accepts the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `boolean` | Enable this profile type to be scraped. | `true` | no
`path` | `string` | The path to the profile type on the target. | `"/debug/pprof/profile"` | no
`delta` | `boolean` | Whether to scrape the profile as a delta. | `true` | no

When the `delta` argument is `true`, a `seconds` query parameter is
automatically added to requests.

### profile.fgprof block

The `profile.fgprof` block collects profiles from an [fgprof][] endpoint.

It accepts the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `boolean` | Enable this profile type to be scraped. | `false` | no
`path` | `string` | The path to the profile type on the target. | `"/debug/fgprof"` | no
`delta` | `boolean` | Whether to scrape the profile as a delta. | `true` | no

When the `delta` argument is `true`, a `seconds` query parameter is
automatically added to requests.

### profile.delta_memory block

The `profile.delta_memory` block collects profiles from [godeltaprof][] memory endpoint.

It accepts the following arguments:

| Name      | Type      | Description                                 | Default                     | Required |
|-----------|-----------|---------------------------------------------|-----------------------------|----------|
| `enabled` | `boolean` | Enable this profile type to be scraped.     | `false`                     | no       |
| `path`    | `string`  | The path to the profile type on the target. | `"/debug/pprof/delta_heap"` | no       |

### profile.delta_mutex block

The `profile.delta_mutex` block collects profiles from [godeltaprof][] mutex endpoint.

It accepts the following arguments:

| Name      | Type      | Description                                 | Default                      | Required |
|-----------|-----------|---------------------------------------------|------------------------------|----------|
| `enabled` | `boolean` | Enable this profile type to be scraped.     | `false`                      | no       |
| `path`    | `string`  | The path to the profile type on the target. | `"/debug/pprof/delta_mutex"` | no       |

### profile.delta_block block

The `profile.delta_block` block collects profiles from [godeltaprof][] block endpoint.

It accepts the following arguments:

| Name      | Type      | Description                                 | Default                      | Required |
|-----------|-----------|---------------------------------------------|------------------------------|----------|
| `enabled` | `boolean` | Enable this profile type to be scraped.     | `false`                      | no       |
| `path`    | `string`  | The path to the profile type on the target. | `"/debug/pprof/delta_block"` | no       |


### profile.custom block

The `profile.custom` block allows for collecting profiles from custom
endpoints. Blocks must be specified with a label:

```river
profile.custom "PROFILE_TYPE" {
  enabled = true
  path    = "PROFILE_PATH"
}
```

Multiple `profile.custom` blocks can be specified. Labels assigned to
`profile.custom` blocks must be unique across the component.

The `profile.custom` block accepts the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `boolean` | Enable this profile type to be scraped. | | yes
`path` | `string` | The path to the profile type on the target. | | yes
`delta` | `boolean` | Whether to scrape the profile as a delta. | `false` | no

When the `delta` argument is `true`, a `seconds` query parameter is
automatically added to requests.

### clustering (beta)

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `bool` | Enables sharing targets with other cluster nodes. | `false` | yes

When the agent is [using clustering][], and `enabled` is set to true,
then this `pyroscope.scrape` component instance opts-in to participating in the
cluster to distribute scrape load between all cluster nodes.

Clustering causes the set of targets to be locally filtered down to a unique
subset per node, where each node is roughly assigned the same number of
targets. If the state of the cluster changes, such as a new node joins, then
the subset of targets to scrape per node will be recalculated.

When clustering mode is enabled, all agents participating in the cluster must
use the same configuration file and have access to the same service discovery
APIs.

If the agent is _not_ running in clustered mode, this block is a no-op.

[using clustering]: {{< relref "../../concepts/clustering.md" >}}

## Exported fields

`pyroscope.scrape` does not export any fields that can be referenced by other
components.

## Component health

`pyroscope.scrape` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`pyroscope.scrape` reports the status of the last scrape for each configured
scrape job on the component's debug endpoint.

## Debug metrics

* `pyroscope_fanout_latency` (histogram): Write latency for sending to direct and indirect components.

## Scraping behavior

The `pyroscope.scrape` component borrows the scraping behavior of Prometheus.
Prometheus, and by extension, this component, uses a pull model for scraping
profiles from a given set of _targets_.
Each scrape target is defined as a set of key-value pairs called _labels_.

The set of targets can either be _static_, or dynamically provided periodically
by a service discovery component such as `discovery.kubernetes`. The special
label `__address__` _must always_ be present and corresponds to the
`<host>:<port>` that is used for the scrape request.

The special label `service_name` is required and must always be present. If it's not specified, it is
attempted to be inferred from multiple sources: 
- `__meta_kubernetes_pod_annotation_pyroscope_io_service_name` which is a `pyroscope.io/service_name` pod annotation.
- `__meta_kubernetes_namespace` and `__meta_kubernetes_pod_container_name`
- `__meta_docker_container_name`

If `service_name` is not specified and could not be inferred it is set to `unspecified`.

By default, the scrape job tries to scrape all available targets' `/debug/pprof`
endpoints using HTTP, with a scrape interval of 15 seconds and scrape timeout of
15 seconds. The profile paths, protocol scheme, scrape interval and timeout,
query parameters, as well as any other settings can be configured using the
component's arguments.

The scrape job expects profiles exposed by the endpoint to follow the
[pprof] protobuf format. All profiles are then propagated
to each receiver listed in the component's `forward_to` argument.

Labels coming from targets, that start with a double underscore `__` are
treated as _internal_, and are removed prior to scraping.

The `pyroscope.scrape` component regards a scrape as successful if it
responded with an HTTP `200 OK` status code and returned a body of valid [pprof] profile.

If the scrape request fails, the component's debug UI section contains more
detailed information about the failure, the last successful scrape, as well as
the labels last used for scraping.

The following labels are automatically injected to the scraped profiles and
can help pin down a scrape target.

| Label        | Description                                                                                      |
|--------------|--------------------------------------------------------------------------------------------------|
| job          | The configured job name that the target belongs to. Defaults to the fully formed component name. |
| instance     | The `__address__` or `<host>:<port>` of the scrape target's URL.                                 |
| service_name | The inferred pyroscope service name                                                              |

## Example

The following example sets up the scrape job with certain attributes (profiling config, targets) and lets it scrape two local applications (the Agent itself and Pyroscope).
The exposed profiles are sent over to the provided list of receivers, as defined by other components.

```river
pyroscope.scrape "local" {
  targets    = [
    {"__address__" = "localhost:4100", "service_name"="pyroscope"},
    {"__address__" = "localhost:12345", "service_name"="agent"},
  ]
  forward_to = [pyroscope.write.local.receiver]
  profiling_config {
    profile.fgprof {
      enabled = true
    }
    profile.block {
      enabled = false
    }
    profile.mutex {
      enabled = false
    }
  }
}
```

Here are the endpoints that are being scraped every 15 seconds:

```
http://localhost:4100/debug/pprof/allocs
http://localhost:4100/debug/pprof/goroutine
http://localhost:4100/debug/pprof/profile?seconds=14
http://localhost:4100/debug/fgprof?seconds=14
http://localhost:12345/debug/pprof/allocs
http://localhost:12345/debug/pprof/goroutine
http://localhost:12345/debug/pprof/profile?seconds=14
http://localhost:12345/debug/fgprof?seconds=14
```
