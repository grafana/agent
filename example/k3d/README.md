# `k3d` Example

The `k3d` example uses `k3d` and `tanka` to produce a Kubernetes environment
that

## Requirements

- A Unix-y command line (macOS or Linux will do).
- Kubectl
- Docker
- [Tanka >= v0.9.2](https://github.com/grafana/tanka)
- [k3d >= v1.5.1 < v3.0.0](https://github.com/rancher/k3d)

## Getting Started

Build latest agent images with `make agent-image agentctl-image` in the project root directory.

Run the following to create your cluster:

```bash
# Create a new k3d cluster
./scripts/create.bash

# Merge the k3d cluster config with your local kubectl config
./scripts/merge_k3d.bash

# Import images into k3d
k3d import-images -n agent-k3d grafana/agent
k3d import-images -n agent-k3d grafana/agentctl

tk apply ./environment

# Navigate to localhost:30080 in your browser to view dashboards

# Take down the cluster after validation
k3d delete -n agent-k3d
```
