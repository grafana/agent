---
title: discovery.docker
---

# discovery.docker

`discovery.docker` discovers [Docker Engine][] containers and exposes them as targets.

[Docker Engine]: https://docs.docker.com/engine/

## Usage

```river
discovery.docker "LABEL" {
  host = "DOCKER_ENGINE_HOST"
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`host` | `string` | Address of the Docker Daemon to connect to. | | yes
`port` | `number` | Port to use for collecting metrics when containers don't have any port mappings. | `80` | no
`host_networking_host` | `string` | Host to use if the container is in host networking mode. | `"localhost"` | no
`refresh_interval` | `duration` | Frequency to refresh list of containers. | `"1m"` | no

## Blocks

The following blocks are supported inside the definition of
`discovery.docker`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
filter | [filter][] | Filters discoverable resources. | no
http_client_config | [http_client_config][] | HTTP client configuration for docker requests. | no
http_client_config > basic_auth | [basic_auth][] | Configure basic_auth for authenticating to the endpoint. | no
http_client_config > authorization | [authorization][] | Configure generic authorization to the endpoint. | no
http_client_config > oauth2 | [oauth2][] | Configure OAuth2 for authenticating to the endpoint. | no
http_client_config > oauth2 > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no

The `>` symbol indicates deeper levels of nesting. For example,
`http_client_config > basic_auth` refers to a `basic_auth` block defined inside
an `http_client_config` block.

[filter]: #filter-block
[http_client_config]: #http_client_config-block
[basic_auth]: #basic_auth-block
[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block

### filter block

The `filter` block configures a filter to pass to the Docker Engine to limit
the amount of containers returned. The `filter` block can be specified multiple
times to provide more than one filter.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`name` | `string` | Filter name to use. | | yes
`values` | `list(string)` | Values to pass to the filter. | | yes

Refer to [List containers][List containers] from the Docker Engine API
documentation for the list of supported filters and their meaning.

[List containers]: https://docs.docker.com/engine/api/v1.41/#tag/Container/operation/ContainerList

### http_client_config block

The `http_client_config` block configures settings used to connect to the
Docker Engine API server.

{{< docs/shared lookup="flow/reference/components/http-client-config-block.md" source="agent" >}}

### basic_auth block

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" >}}

### authorization block

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" >}}

### oauth2 block

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" >}}

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`targets` | `list(map(string))` | The set of targets discovered from the docker API.

Each target includes the following labels:

* `__meta_docker_container_id`: ID of the container.
* `__meta_docker_container_name`: Name of the container.
* `__meta_docker_container_network_mode`: Network mode of the container.
* `__meta_docker_container_label_<labelname>`: Each label from the container.
* `__meta_docker_network_id`: ID of the Docker network the container is in.
* `__meta_docker_network_name`: Name of the Docker network the container is in.
* `__meta_docker_network_ingress`: Set to `true` if the Docker network is an
  ingress network.
* `__meta_docker_network_internal`: Set to `true` if the Docker network is an
  internal network.
* `__meta_docker_network_label_<labelname>`: Each label from the network the
  container is in.
* `__meta_docker_network_scope`: The scope of the network the container is in.
* `__meta_docker_network_ip`: The IP of the container in the network.
* `__meta_docker_port_private`: The private port on the container.
* `__meta_docker_port_public`: The publicly exposed port from the container,
  if a port mapping exists.
* `__meta_docker_port_public_ip`: The public IP of the container, if a port
  mapping exists.

Each discovered container maps to one target per unique combination of networks
and port mappings used by the container.

## Component health

`discovery.docker` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.docker` does not expose any component-specific debug information.

### Debug metrics

`discovery.docker` does not expose any component-specific debug metrics.

## Examples

### Linux or macOS hosts

This example discovers Docker containers when the host machine is macOS or
Linux:

```river
discovery.docker "containers" {
  host = "unix:///var/run/docker.sock"
}
```

### Windows hosts

This example discovers Docker containers when the host machine is Windows:

```river
discovery.docker "containers" {
  host = "tcp://localhost:2375"
}
```

> **NOTE**: This example requires the "Expose daemon on tcp://localhost:2375
> without TLS" setting to be enabled in the Docker Engine settings.
