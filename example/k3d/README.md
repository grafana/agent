# `k3d` Example

The `k3d` example uses `k3d` and `tanka` to produce a Kubernetes environment
that

## Requirements

- A Unix-y command line (macOS or Linux will do).
- Kubectl
- Docker
- [Tanka >= v0.9.2](https://github.com/grafana/tanka)
- [k3d >= v4.0.0](https://github.com/rancher/k3d)

## Getting Started

Build latest agent images with `make agent-image agentctl-image` in the project root directory if there are local changes to test.

Run the following to create your cluster:

```bash
# Create a new k3d cluster
./scripts/create.bash

# Import images into k3d if they are not available on docker hub
k3d image import -c agent-k3d grafana/agent:latest
k3d image import -c agent-k3d grafana/agentctl:latest

tk apply ./environment

# Navigate to grafana.k3d.localhost:30080 in your browser to view dashboards

# Delete the k3d cluster when you're done with it
k3d cluster delete agent-k3d
```
