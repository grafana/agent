---
title: discovery.kubelet
labels:
  stage: beta
---

# discovery.kubelet

`discovery.kubelet` discovers Kubernetes pods running on the specified Kubelet
and exposes them as scrape targets.

## Usage

```river
discovery.kubelet "LABEL" {
  bearer_token_file = "TOKEN_FILE"
}
```

## Requirements

* The Kubelet must be reachable from the `grafana-agent` pod network.
* Follow the [Kubelet authorization](https://kubernetes.io/docs/reference/access-authn-authz/kubelet-authn-authz/#kubelet-authorization)
  documentation to configure authentication to the Kubelet API.

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`url` | `string` | URL of the Kubelet server. | | no
`bearer_token` | `secret` | Bearer token to authenticate with. | | no
`bearer_token_file` | `string` | File containing a bearer token to authenticate with. | | no
`refresh_interval` | `duration` | How often the Kubelet should be polled for scrape targets | `5s` | no
`namespaces` | `list(string)` | A list of namespaces to extract target pods from | | no

One of the following authentication methods must be provided if kubelet authentication is enabled
 - [`bearer_token` argument](#arguments).
 - [`bearer_token_file` argument](#arguments).
 - [`authorization` block][authorization].

The `namespaces` list limits the namespaces to discover resources in. If
omitted, all namespaces are searched.

## Blocks

The following blocks are supported inside the definition of
`discovery.kubelet`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
authorization | [authorization][] | Configure generic authorization to the endpoint. | no
tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no

[authorization]: #authorization-block
[tls_config]: #tls_config-block

### authorization block

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" >}}

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`targets` | `list(map(string))` | The set of targets discovered from the Kubelet API.

Each target includes the following labels:

* `__address__`: The target address to scrape derived from the pod IP and container port.
* `__meta_kubernetes_namespace`: The namespace of the pod object.
* `__meta_kubernetes_pod_name`: The name of the pod object.
* `__meta_kubernetes_pod_ip`: The pod IP of the pod object.
* `__meta_kubernetes_pod_label_<labelname>`: Each label from the pod object.
* `__meta_kubernetes_pod_labelpresent_<labelname>`: `true` for each label from
  the pod object.
* `__meta_kubernetes_pod_annotation_<annotationname>`: Each annotation from the
  pod object.
* `__meta_kubernetes_pod_annotationpresent_<annotationname>`: `true` for each
  annotation from the pod object.
* `__meta_kubernetes_pod_container_init`: `true` if the container is an
  `InitContainer`.
* `__meta_kubernetes_pod_container_name`: Name of the container the target
  address points to.
* `__meta_kubernetes_pod_container_id`: ID of the container the target address
  points to. The ID is in the form `<type>://<container_id>`.
* `__meta_kubernetes_pod_container_image`: The image the container is using.
* `__meta_kubernetes_pod_container_port_name`: Name of the container port.
* `__meta_kubernetes_pod_container_port_number`: Number of the container port.
* `__meta_kubernetes_pod_container_port_protocol`: Protocol of the container
  port.
* `__meta_kubernetes_pod_ready`: Set to `true` or `false` for the pod's ready
  state.
* `__meta_kubernetes_pod_phase`: Set to `Pending`, `Running`, `Succeeded`, `Failed` or
  `Unknown` in the lifecycle.
* `__meta_kubernetes_pod_node_name`: The name of the node the pod is scheduled
  onto.
* `__meta_kubernetes_pod_host_ip`: The current host IP of the pod object.
* `__meta_kubernetes_pod_uid`: The UID of the pod object.
* `__meta_kubernetes_pod_controller_kind`: Object kind of the pod controller.
* `__meta_kubernetes_pod_controller_name`: Name of the pod controller.

## Component health

`discovery.kubelet` is reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.kubelet` does not expose any component-specific debug information.

### Debug metrics

`discovery.kubelet` does not expose any component-specific debug metrics.

## Examples

### Bearer token file authentication

This example uses a bearer token file to authenticate to the Kubernetes API:

```river
discovery.kubelet "k8s_pods" {
  bearer_token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"
}
```

### Limit searched namespaces

This example limits the namespaces where pods are discovered using the `namespaces` argument:

```river
discovery.kubelet "k8s_pods" {
  bearer_token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"
  namespaces = ["default", "kube-system"]
}
```
