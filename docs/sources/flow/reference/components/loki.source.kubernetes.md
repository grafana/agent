---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/loki.source.kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/loki.source.kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/loki.source.kubernetes/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/loki.source.kubernetes/
labels:
  stage: experimental
title: loki.source.kubernetes
description: Learn about loki.source.kubernetes
---

# loki.source.kubernetes

{{< docs/shared lookup="flow/stability/experimental.md" source="agent" version="<AGENT VERSION>" >}}

`loki.source.kubernetes` tails logs from Kubernetes containers using the
Kubernetes API. It has the following benefits over `loki.source.file`:

* It works without a privileged container.
* It works without a root user.
* It works without needing access to the filesystem of the Kubernetes node.
* It doesn't require a DaemonSet to collect logs, so one agent could collect
  logs for the whole cluster.

> **NOTE**: Because `loki.source.kubernetes` uses the Kubernetes API to tail
> logs, it uses more network traffic and CPU consumption of Kubelets than
> `loki.source.file`.

Multiple `loki.source.kubernetes` components can be specified by giving them
different labels.

## Usage

```river
loki.source.kubernetes "LABEL" {
  targets    = TARGET_LIST
  forward_to = RECEIVER_LIST
}
```

## Arguments

The component starts a new reader for each of the given `targets` and fans out
log entries to the list of receivers passed in `forward_to`.

`loki.source.kubernetes` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`targets` | `list(map(string))` | List of files to read from. | | yes
`forward_to` | `list(LogsReceiver)` | List of receivers to send log entries to. | | yes

Each target in `targets` must have the following labels:

* `__meta_kubernetes_namespace` or `__pod_namespace__` to specify the namespace
  of the pod to tail.
* `__meta_kubernetes_pod_name` or `__pod_name__` to specify the name of the pod
  to tail.
* `__meta_kubernetes_pod_container_name` or `__pod_container_name__` to specify
  the container within the pod to tail.
* `__meta_kubernetes_pod_uid` or `__pod_uid__` to specify the UID of the pod to
  tail.

By default, all of these labels are present when the output
`discovery.kubernetes` is used.

A log tailer is started for each unique target in `targets`. Log tailers will
reconnect with exponential backoff to Kubernetes if the log stream returns
before the container has permanently terminated.

## Blocks

The following blocks are supported inside the definition of
`loki.source.kubernetes`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
client | [client][] | Configures Kubernetes client used to tail logs. | no
client > basic_auth | [basic_auth][] | Configure basic_auth for authenticating to the endpoint. | no
client > authorization | [authorization][] | Configure generic authorization to the endpoint. | no
client > oauth2 | [oauth2][] | Configure OAuth2 for authenticating to the endpoint. | no
client > oauth2 > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no
client > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no

The `>` symbol indicates deeper levels of nesting. For example, `client >
basic_auth` refers to a `basic_auth` block defined
inside a `client` block.

[client]: #client-block
[basic_auth]: #basic_auth-block
[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block

### client block

The `client` block configures the Kubernetes client used to tail logs from
containers. If the `client` block isn't provided, the default in-cluster
configuration with the service account of the running Grafana Agent pod is
used.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`api_server` | `string` | URL of the Kubernetes API server. | | no
`kubeconfig_file` | `string` | Path of the `kubeconfig` file to use for connecting to Kubernetes. | | no
`bearer_token` | `secret` | Bearer token to authenticate with. | | no
`bearer_token_file` | `string` | File containing a bearer token to authenticate with. | | no
`proxy_url` | `string` | HTTP proxy to proxy requests through. | | no
`follow_redirects` | `bool` | Whether redirects returned by the server should be followed. | `true` | no
`enable_http2` | `bool` | Whether HTTP2 is supported for requests. | `true` | no

 At most one of the following can be provided:
 - [`bearer_token` argument][client].
 - [`bearer_token_file` argument][client].
 - [`basic_auth` block][basic_auth].
 - [`authorization` block][authorization].
 - [`oauth2` block][oauth2].

### basic_auth block

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" version="<AGENT VERSION>" >}}

### authorization block

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" version="<AGENT VERSION>" >}}

### oauth2 block

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" version="<AGENT VERSION>" >}}

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

`loki.source.kubernetes` does not export any fields.

## Component health

`loki.source.kubernetes` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`loki.source.kubernetes` exposes some target-level debug information per
target:

* The labels associated with the target.
* The full set of labels which were found during service discovery.
* The most recent time a log line was read and forwarded to the next components
  in the pipeline.
* The most recent error from tailing, if any.

## Debug metrics

`loki.source.kubernetes` does not expose any component-specific debug metrics.

## Example

This example collects logs from all Kubernetes pods and forwards them to a
`loki.write` component so they are written to Loki.

```river
discovery.kubernetes "pods" {
  role = "pod"
}

loki.source.kubernetes "pods" {
  targets    = discovery.kubernetes.pods.targets
  forward_to = [loki.write.local.receiver]
}

loki.write "local" {
  endpoint {
    url = env("LOKI_URL")
  }
}
```
