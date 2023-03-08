# Grafana Agent Helm chart

![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![Version: 0.9.0](https://img.shields.io/badge/Version-0.9.0-informational?style=flat-square) ![AppVersion: v0.32.1](https://img.shields.io/badge/AppVersion-v0.32.1-informational?style=flat-square)

Helm chart for deploying [Grafana Agent][] to Kubernetes.

[Grafana Agent]: https://grafana.com/docs/agent/latest/

## Usage

### Setup Grafana chart repository

```
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

### Install chart

To install the chart with the release name my-release:

`helm install my-release grafana/grafana-agent`

This chart installs one instance of Grafana Agent into your Kubernetes cluster
using a specific Kubernetes controller. By default, DaemonSet is used. The
`controller.type` value can be used to change the controller to either a
StatefulSet or Deployment.

Creating multiple installations of the Helm chart with different controllers is
useful if just using the default DaemonSet isn't sufficient.

## Flow mode is the default

By default, [Grafana Agent Flow][Flow] is deployed. To opt out of Flow mode and
use the older mode (called "static mode"), set the `agent.mode` value to
`static`.

[Flow]: https://grafana.com/docs/agent/latest/flow/

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| agent.configMap.content | string | `""` | Content to assign to the new ConfigMap.  This is passed into `tpl` allowing for templating from values. |
| agent.configMap.create | bool | `true` | Create a new ConfigMap for the config file. |
| agent.configMap.key | string | `nil` | Key in ConfigMap to get config from. |
| agent.configMap.name | string | `nil` | Name of existing ConfigMap to use. Used when create is false. |
| agent.enableReporting | bool | `true` | Enables sending Grafana Labs anonymous usage stats to help improve Grafana Agent. |
| agent.extraArgs | list | `[]` | Extra args to pass to `agent run`: https://grafana.com/docs/agent/latest/flow/reference/cli/run/ |
| agent.extraEnv | list | `[]` | Extra environment variables to pass to the agent container. |
| agent.extraPorts | list | `[]` | Extra ports to expose on the Agent |
| agent.faroPort | int | `12347` | Port to listen for faro traffic on. |
| agent.listenAddr | string | `"0.0.0.0"` | Address to listen for traffic on. 0.0.0.0 exposes the UI to other containers. |
| agent.listenPort | int | `80` | Port to listen for traffic on. |
| agent.mode | string | `"flow"` | Mode to run Grafana Agent in. Can be "flow" or "static". |
| agent.mounts.dockercontainers | bool | `false` | Mount /var/lib/docker/containers from the host into the container for log collection. |
| agent.mounts.extra | list | `[]` | Extra volume mounts to add into the Grafana Agent container. Does not affect the watch container. |
| agent.mounts.varlog | bool | `false` | Mount /var/log from the host into the container for log collection. |
| agent.resources | object | `{}` | Resource requests and limits to apply to the Grafana Agent container. |
| agent.securityContext | object | `{}` | Security context to apply to the Grafana Agent container. |
| agent.storagePath | string | `"/tmp/agent"` | Path to where Grafana Agent stores data (for example, the Write-Ahead Log). By default, data is lost between reboots. |
| configReloader.customArgs | list | `[]` | Override the args passed to the container. |
| configReloader.enabled | bool | `true` | Enables automatically reloading when the agent config changes. |
| configReloader.image.repository | string | `"jimmidyson/configmap-reload"` | Repository to get config reloader image from. |
| configReloader.image.tag | string | `"v0.8.0"` | Tag of image to use for config reloading. |
| controller.podAnnotations | object | `{}` | Extra pod annotations to add. |
| controller.affinity | object | `{}` | Affinity configuration for pods. |
| controller.podLabels | object | `{}` | Extra pod labels to add. |
| controller.priorityClassName | string | `""` | priorityClassName to apply to Grafana Agent pods. |
| controller.replicas | int | `1` | Number of pods to deploy. Ignored when controller.type is 'daemonset'. |
| controller.tolerations | list | `[]` | Tolerations to apply to Grafana Agent pods. |
| controller.type | string | `"daemonset"` | Type of controller to use for deploying Grafana Agent in the cluster. Must be one of 'daemonset', 'deployment', or 'statefulset'. |
| controller.updateStrategy | object | `{}` | Update strategy for updating deployed Pods. |
| controller.volumeClaimTemplates | list | `[]` | volumeClaimTemplates to add when controller.type is 'statefulset'. |
| controller.volumes.extra | list | `[]` | Extra volumes to add to the Grafana Agent pod. |
| fullnameOverride | string | `nil` | Overrides the chart's computed fullname. Used to change the full prefix of resource names. |
| image.pullPolicy | string | `"IfNotPresent"` | Grafana Agent image pull policy. |
| image.pullSecrets | list | `[]` | Optional set of image pull secrets. |
| image.repository | string | `"grafana/agent"` | Grafana Agent image repository. |
| image.tag | string | `nil` | Grafana Agent image tag. When empty, the Chart's appVersion is used. |
| ingress.annotations | object | `{}` |  |
| ingress.enabled | bool | `false` | Enables ingress for the agent (faro port) |
| ingress.extraPaths | list | `[]` |  |
| ingress.hosts[0] | string | `"chart-example.local"` |  |
| ingress.labels | object | `{}` |  |
| ingress.path | string | `"/"` |  |
| ingress.pathType | string | `"Prefix"` |  |
| ingress.tls | list | `[]` |  |
| nameOverride | string | `nil` | Overrides the chart's name. Used to change the infix in the resource names. |
| rbac.create | bool | `true` | Whether to create RBAC resources for the agent. |
| service.clusterIP | string | `""` | Cluster IP, can be set to None, empty "" or an IP address |
| service.type | string | `"ClusterIP"` | Service type |
| serviceAccount.annotations | object | `{}` | Annotations to add to the created service account. |
| serviceAccount.create | bool | `true` | Whether to create a service account for the Grafana Agent deployment. |
| serviceAccount.name | string | `nil` | The name of the existing service account to use when serviceAccount.create is false. |

### agent.extraArgs

`agent.extraArgs` allows for passing extra arguments to the Grafana Agent
container. The list of available arguments is documented on [agent run][].

> **WARNING**: Using `agent.extraArgs` does not have a stable API. Things may
> break between Chart upgrade if an argument gets added to the template.

[agent run]: https://grafana.com/docs/agent/latest/flow/reference/cli/run/

### agent.extraPorts

`agent.extraPorts` allows for configuring specific open ports.

The detained specification of ports can be found at the [Kubernetes Pod documents](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#ports).

Port numbers specified must be 0 < x < 65535.

| ChartPort | KubePort | Description |
|-----------|----------|-------------|
| targetPort | containerPort | Number of port to expose on the pod's IP address. |
| hostPort | hostPort | (Optional) Number of port to expose on the host. Daemonsets taking traffic might find this useful. |
| name | name | If specified, this must be an `IANA_SVC_NAME` and unique within the pod. Each named port in a pod must have a unique name. Name for the port that can be referred to by services.
| protocol | protocol | Must be UDP, TCP, or SCTP. Defaults to "TCP". |

### agent.listenAddr

`agent.listenAddr` allows for restricting which address the agent listens on
for network traffic on its HTTP server. By default, this is `0.0.0.0` to allow
its UI to be exposed when port-forwarding and to expose its metrics to other
agents in the cluster.

### agent.configMap.config

`agent.configMap.content` holds the Grafana Agent configuration to use.

If `agent.configMap.content` is not provided, a [default configuration file][default-config] is
used. When provided, `agent.configMap.content` must hold a valid River configuration file.

[default-config]: ./config/example.river

### controller.securityContext

`controller.securityContext` sets the securityContext passed to the Grafana
Agent container.

By default, Grafana Agent containers are not able to collect telemetry from the
host node or other specific types of privileged telemetry data. See [Collecting
logs from other containers][#collecting-logs-from-other-containers] and
[Collecting host node telemetry][#collecting-host-node-telemetry] below for
more information on how to enable these capabilities.

### rbac.create

`rbac.create` enables the creation of ClusterRole and ClusterRoleBindings for
the Grafana Agent containers to use. The default permission set allows Flow
components like [discovery.kubernetes][] to work properly.

[discovery.kubernetes]: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.kubernetes/

## Collecting logs from other containers

There are two ways to collect logs from other containers within the cluster
the agent is deployed in.

### Versions >= 0.31.x

The [loki.source.kubernetes][] component introduced in 0.31.0 may be used to
collect logs as an alternative to tailing files from the host. This component
does not require mounting the hosts filesystem into the Agent, nor requires
additional security contexts to work correctly.

However, `loki.source.kubernetes` is experimental and may have issues not
present in the file-based approach.

[loki.source.kubernetes]: https://grafana.com/docs/agent/latest/flow/reference/components/loki.source.kubernetes/

### Versions < 0.31.x

For those running the Agent on versions prior to 0.31.0, the only way to collect logs
from other containers is to mount `/var/lib/docker/containers` from the host and read
the log files directly.

This capability is disabled by default.

To expose logs from other containers to Grafana Agent:

* Set `agent.mounts.dockercontainers` to `true`.
* Set `controller.securityContext` to:
  ```yaml
  privileged: true
  runAsUser: 0
  ```

## Collecting host node telemetry

Telemetry from the host, such as host-specific log files (from `/var/logs`) or
metrics from `/proc` and `/sys` are not accessible to Grafana Agent containers.

To expose this information to Grafana Agent for telemetry collection:

* Set `agent.mounts.dockercontainers` to `true`.
* Mount `/proc` and `/sys` from the host into the container.
* Set `controller.securityContext` to:
  ```yaml
  privileged: true
  runAsUser: 0
  ```
