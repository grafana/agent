# Running Grafana Agent

There are a few ways to run the Grafana Agent, in order from
easiest to hardest:

- [Use the Install Script for Kubernetes](#install-script-for-kubernetes)
- [Windows Installation](#windows-installation)
- [Run the Agent with Docker](#running-the-agent-with-docker)
- [Run the Agent locally](#running-the-agent-locally)
- [Use the example Kubernetes configs](#use-the-example-kubernetes-configs)
- [Build the Agent from Source](#build-the-agent-from-source)
- [Use our production Tanka configs](#use-our-production-tanka-configs)

## Install Script for Kubernetes

The Grafana Agent repository comes with installation scripts to
configure components and return a Kubernetes manifest that uses our preferred
defaults. To run the script, copy and paste this in your terminal:

```
NAMESPACE="default" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install.sh)" | kubectl apply -f -
NAMESPACE="default" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install-loki.sh)" | kubectl apply -f -
NAMESPACE="default" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install-tempo.sh)" | kubectl apply -f -
```

See the [Kubernetes README](./kubernetes/README.md) for more information.

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
  grafana/agent:v0.16.1
```

## Running the Agent locally

Currently, you must provide your own system configuration files to run the
Agent as a long-living process (e.g., write your own systemd unit files).

## Use the example Kubernetes configs

The install script replaces variable placeholders in the [example Kubernetes
manifest](./kubernetes/agent.yaml) in the Kubernetes directory. Feel free to
examine that file and modify it for your own needs!

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
