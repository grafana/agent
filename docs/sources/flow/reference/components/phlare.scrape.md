---
title: phlare.scrape
---

# phlare.scrape

`phlare.scrape` configures a [pprof] scraping job for a given set of
`targets`. The scraped performance profiles are forwarded to the list of receivers passed in
`forward_to`.

Multiple `phlare.scrape` components can be specified by giving them different labels.

## Usage

```
phlare.scrape "LABEL" {
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

## Blocks

The following blocks are supported inside the definition of `phlare.scrape`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
http_client_config | [http_client_config][] | HTTP client settings when connecting to targets. | no
http_client_config > basic_auth | [basic_auth][] | Configure basic_auth for authenticating to targets. | no
http_client_config > authorization | [authorization][] | Configure generic authorization to targets. | no
http_client_config > oauth2 | [oauth2][] | Configure OAuth2 for authenticating to targets. | no
http_client_config > oauth2 > tls_config | [tls_config][] | Configure TLS settings for connecting to targets via OAuth2. | no
http_client_config > tls_config | [tls_config][] | Configure TLS settings for connecting to targets. | no
profiling_config | [profiling_config][] | Configure profiling settings for the scrape job. | no
profiling_config > profile.memory | [profile.memory][] | Collect memory profiles. | no
profiling_config > profile.block | [profile.block][] | Collect profiles on blocks. | no
profiling_config > profile.goroutine | [profile.goroutine][] | Collect goroutine profiles. | no
profiling_config > profile.mutex | [profile.mutex][] | Collect mutex profiles. | no
profiling_config > profile.process_cpu | [profile.process_cpu][] | Collect CPU profiles. | no
profiling_config > profile.fgprof | [profile.fgprof][] | Collect [fgprof][] profiles. | no
profiling_config > profile.custom | [profile.custom][] | Collect custom profiles. | no

The `>` symbol indicates deeper levels of nesting. For example,
`http_client_config > basic_auth` refers to a `basic_auth` block defined inside
an `http_client_config` block.

[http_client_config]: #http_client_config-block
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
[profile.custom]: #profile.custom-block
[pprof]: https://github.com/google/pprof/blob/main/doc/README.md

[fgprof]: https://github.com/felixge/fgprof

### http_client_config block

The `http_client_config` block configures settings used to connect to
endpoints.

{{< docs/shared lookup="flow/reference/components/http-client-config-block.md" source="agent" >}}

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

## Exported fields

`phlare.scrape` does not export any fields that can be referenced by other
components.

## Component health

`phlare.scrape` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`phlare.scrape` reports the status of the last scrape for each configured
scrape job on the component's debug endpoint.

## Debug metrics

* `phlare_fanout_latency` (histogram): Write latency for sending to direct and indirect components.

## Scraping behavior

The `phlare.scrape` component borrows the scraping behavior of Prometheus.
Prometheus, and by extension, this component, uses a pull model for scraping
profiles from a given set of _targets_.
Each scrape target is defined as a set of key-value pairs called _labels_.

The set of targets can either be _static_, or dynamically provided periodically
by a service discovery component such as `discovery.kubernetes`. The special
label `__address__` _must always_ be present and corresponds to the
`<host>:<port>` that is used for the scrape request.

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

The `phlare.scrape` component regards a scrape as successful if it
responded with an HTTP `200 OK` status code and returned a body of valid [pprof] profile.

If the scrape request fails, the component's debug UI section contains more
detailed information about the failure, the last successful scrape, as well as
the labels last used for scraping.

The following labels are automatically injected to the scraped profiles and
can help pin down a scrape target.

Label                 | Description
--------------------- | ----------
job                   | The configured job name that the target belongs to. Defaults to the fully formed component name.
instance              | The `__address__` or `<host>:<port>` of the scrape target's URL.

## Example

The following example sets up the scrape job with certain attributes (profiling config, targets) and lets it scrape two local applications (the Agent itself and Phlare).
The exposed profiles are sent over to the provided list of receivers, as defined by other components.

```river
phlare.scrape "local" {
  targets    = [
    {"__address__" = "localhost:4100", "app"="phlare"},
    {"__address__" = "localhost:12345", "app"="agent"},
  ]
  forward_to = [phlare.write.local.receiver]
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

Here are the the endpoints that are being scraped every 15 seconds:

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
