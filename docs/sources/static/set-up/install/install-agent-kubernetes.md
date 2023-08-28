---
canonical: https://grafana.com/docs/agent/latest/static/set-up/install/install-agent-kubernetes/
menuTitle: Kubernetes
title: Deploy Grafana Agent in static mode on Kubernetes
weight: 300
---

# Deploy Grafana Agent in static mode on Kubernetes

You can deploy Grafana Agent in static mode on Kubernetes.

## Deploy

To deploy Grafana Agent in static mode on Kubernetes, perform the following steps.

1. Download one of the following manifests from GitHub and save it as `manifest.yaml`:

   - Metric collection (StatefulSet): [agent-bare.yaml](https://github.com/grafana/agent/blob/main/production/kubernetes/agent-bare.yaml)
   - Log collection (DaemonSet): [agent-loki.yaml](https://github.com/grafana/agent/blob/main/production/kubernetes/agent-loki.yaml)
   - Trace collection (Deployment): [agent-traces.yaml](https://github.com/grafana/agent/blob/main/production/kubernetes/agent-traces.yaml)

1. Edit the downloaded `manifest.yaml` and replace the placeholders with information relevant to your Kubernetes deployment.

1. Apply the modified manifest file:

   ```shell
   kubectl -n default apply -f manifest.yaml
   ```

{{% admonition type="note" %}}
The manifests do not include the `ConfigMaps` which are necessary to run Grafana Agent.
{{% /admonition %}}

For sample configuration files and detailed instructions, refer to the Grafana Cloud Kubernetes quick start guides:

- [Grafana Agent Metrics Kubernetes Quickstart](https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s/k8s_agent_metrics/)
- [Grafana Agent Logs Kubernetes Quickstart](https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s/k8s_agent_logs/)
- [Grafana Agent Traces Kubernetes Quickstart](https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s/k8s_agent_traces/)

## Rebuild the Kubernetes manifests

The manifests provided are created using Grafana Labs' production Tanka configs with some default values. If you want to build the YAML file with some custom values, you must install the following applications:

- [Tanka](https://github.com/grafana/tanka) version 0.8 or higher
- [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler) version 0.2.1 or higher

Refer to the [`template` Tanka environment](https://github.com/grafana/agent/blob/main/production/kubernetes/build/templates) for the current settings that initialize the Grafana Agent Tanka configurations.

To build the YAML files, run the `/build/build.sh` script, or run `make example-kubernetes` from the project's root directory.
