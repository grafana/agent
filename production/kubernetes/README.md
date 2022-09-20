# Kubernetes Config

This directory contains Kubernetes manifest templates for rolling out the Agent.

Manifests:

- Metric collection (StatefulSet): [`agent-bare.yaml`](./agent-bare.yaml)
- Log collection (DaemonSet): [`agent-loki.yaml`](./agent-loki.yaml)
- Trace collection (Deployment): [`agent-traces.yaml`](./agent-traces.yaml)

⚠️  **These manifests do not include the Agent's configuration (ConfigMaps)**,
which are necessary to run the Agent.

For sample configurations and detailed installation instructions, please head to:

- [Grafana Agent Metrics Kubernetes Quickstart](https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s/k8s_agent_metrics/)
- [Grafana Agent Logs Kubernetes Quickstart](https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s/k8s_agent_logs/)
- [Grafana Agent Traces Kubernetes Quickstart](https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s/k8s_agent_traces/)

## Manually Applying

Since the manifest files are just templates, note that they are *not* ready for
applying out of the box and you will have to manually perform the following steps:

1. Download the manifest as `manifest.yaml`

2. Modify your copy of the manifest, replacing relevant variables with the appropriate values

3. Apply the modified manifest file: `kubectl -n default apply -f manifest.yaml`.

This directory also contains an `install-bare.sh` script that is used inside of
Grafana Cloud instructions. If using the Grafana Agent outside of Grafana Cloud,
it is recommended to follow the steps above instead of calling this script
directly.

## Rebuilding the manifests

The manifests provided are created using Grafana Labs' production
[Tanka configs](../tanka/grafana-agent) with some default values. If you want to
build the YAML file with some custom values, you will need the following pieces
of software installed:

1. [Tanka](https://github.com/grafana/tanka) >= v0.8
2. [`jsonnet-bundler`](https://github.com/jsonnet-bundler/jsonnet-bundler) >= v0.2.1

See the [`template` Tanka environment](./build/templates) for the current
settings that initialize the Grafana Agent Tanka configs.

To build the YAML files, execute the `./build/build.sh` script or run `make example-kubernetes`
from the project's root directory.
