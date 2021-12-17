# Running Grafana Agent

Here are some resources to help you run the Grafana Agent:

- [Windows Installation](#windows-installation)
- [Run the Agent with Docker](#running-the-agent-with-docker)
- [Run the Agent locally](#running-the-agent-locally)
- [Use the example Kubernetes configs](#use-the-example-kubernetes-configs)
- [Grafana Cloud Kubernetes Quickstart Guides](#grafana-cloud-kubernetes-quickstart-guides)
- [Agent Operator Helm Quickstart](#agent-operator-helm-quickstart-guide)
- [Build the Agent from Source](#build-the-agent-from-source)
- [Use our production Tanka configs](#use-our-production-tanka-configs)

## Windows Installation

To run the Windows Installation, download the Windows Installer executable from the [release page](https://github.com/grafana/agent/releases). Then run the installer, this will setup the Agent and run the Agent as a Windows Service. More details can be found in the [Windows Guide](../docs/getting-started/install-agent-on-windows.md)

## Running the Agent with Docker

To run the Agent with Docker, you should have a configuration file on
your local machine ready to bind mount into the container. Then modify
the following command for your environment. Replace `/path/to/config.yaml` with
the full path to your YAML configuration, and replace `/tmp/agent` with the
directory on your host that you want the agent to store its WAL.

```
docker run \
  -v /tmp/agent:/etc/agent/data \
  -v /path/to/config.yaml:/etc/agent/agent.yaml \
  grafana/agent:v0.21.1
```

## Running the Agent locally

Currently, you must provide your own system configuration files to run the
Agent as a long-living process (e.g., write your own systemd unit files).

## Use the example Kubernetes configs

You can find sample deployment manifests in the [Kubernetes](./kubernetes) directory.

## Grafana Cloud Kubernetes quickstart guides

These guides help you get up and running with the Agent and Grafana Cloud, and include sample ConfigMaps.

You can find them in the [Grafana Cloud documentation](https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s/)

## Agent Operator Helm quickstart guide

This guide will show you how to deploy the [Grafana Agent Operator](https://grafana.com/docs/agent/latest/operator/) into your Kubernetes cluster using the [grafana-agent-operator Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/agent-operator).

You'll also deploy the following custom resources (CRs):
- A `GrafanaAgent` resource, which discovers one or more `MetricsInstance` and `LogsInstances` resources.
- A `MetricsInstance` resource that defines where to ship collected metrics.
- A `ServiceMonitor` resource to collect cAdvisor and kubelet metrics.
- A `LogsInstance` resource that defines where to ship collected logs.
- A `PodLogs` resource to collect container logs from Kubernetes Pods.

You can find the guide [here](https://grafana.com/docs/agent/latest/operator/helm-getting-started/).

## Build the Agent from source

Go 1.14 is currently needed to build the agent from source. Run `make agent`
from the root of this repository, and then the build agent binary will be placed
at `./cmd/agent/agent`.

## Use our production Tanka configs

The Tanka configs we use to deploy the agent ourselves can be found in our
[production Tanka directory](./tanka/grafana-agent). These configs are also used
to generate the Kubernetes configs for the install script. To get started with
the tanka configs, do the following:

```
mkdir tanka-agent
cd tanka-agent
tk init --k8s=false
jb install github.com/grafana/agent/production/tanka/grafana-agent

# substitute your target k8s version for "1.16" in the next few commands
jb install github.com/jsonnet-libs/k8s-alpha/1.16
echo '(import "github.com/jsonnet-libs/k8s-alpha/1.16/main.libsonnet")' > lib/k.libsonnet
echo '+ (import "github.com/jsonnet-libs/k8s-alpha/1.16/extensions/kausal-shim.libsonnet")' >> lib/k.libsonnet
```

Then put this in `environments/default/main.jsonnet`:
```
local agent = import 'grafana-agent/grafana-agent.libsonnet';

agent {
  _config+:: {
    namespace: 'grafana-agent'
  },
}
```

If all these steps worked, `tk eval environments/default` should output the
default JSON we use to build our Kubernetes manifests.
