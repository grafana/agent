---
title: Install Grafana Agent in static mode on Kubernetes
menuTitle: Kubernetes
weight: 300
---

# Install Grafana Agent in static mode on Kubernetes

Deploy Kubernetes manifests from the [`kubernetes` directory](https://github.com/grafana/agent/tree/main/production/kubernetes).
You can manually modify the Kubernetes manifests by downloading them. These manifests do not include Grafana Agent configuration files.

For sample configuration files, refer to the Grafana Cloud Kubernetes quick start guide: https://grafana.com/docs/grafana-cloud/kubernetes/agent-k8s/.

Advanced users can use the Grafana Agent Operator to deploy the Grafana Agent on Kubernetes.

## Next steps

- [Start Grafana Agent]({{< relref "../start-agent/" >}})
- [Configure Grafana Agent]({{< relref "../../configuration/" >}})
