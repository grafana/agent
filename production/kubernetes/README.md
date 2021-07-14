# Kubernetes Config

This directory contains Kubernetes manifest templates and installation scripts
for rendering the templates so they can be applied against Kubernetes.

Manifests:

- Metric collection: [`agent.yaml`](./agent.yaml)
- Log collection: [`agent-loki.yaml`](./agent-loki.yaml)
- Trace collection: [`agent-tempo.yaml`](./agent-tempo.yaml)

Installation script:

- Metric collection: [`install.sh`](./install.sh)
- Log collection: [`install-loki.sh`](./install-loki.sh)
- Trace collection: [`install-tempo.sh`](./install-tempo.sh)

## Install Scripts

There are two installation scripts, one for metrics and the other for logs. Each
install script does the following:

1. Prmopts the user for their remote target credentials (Prometheus remote_write, Loki client).
2. Downloads the manifest template from GitHub
3. Substitutes variables in the template with the provided input from
   step 1.
4. Prints out the final manifest to stdout without applying it.

Here's a script to copy and paste to install the Agent on Kubernetes for
collecting metrics, logs, and traces (requires `envsubst` (GNU gettext)):

```
NAMESPACE="default" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install.sh)" | kubectl apply -f -
NAMESPACE="default" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install-loki.sh)" | kubectl apply -f -
NAMESPACE="default" /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install-tempo.sh)" | kubectl apply -f -
```

## Manually Applying

Since the manifest files are just templates, note that they are *not* ready for
applying out of the box and you will have to manually reroduce the steps that
the installation script does:

1. Download the manifest as `manifest.yaml`.

2. Modify your copy of the manifest, replacing all variables with the
   appropriate values.

3. Apply the modified manifest file: `kubectl -ndefault apply -f manifest.yaml`.

## Rebuilding the manifests

The manifests provided are created using Grafana Labs' production
[Tanka configs](../tanka/grafana-agent) with some default values. If you want to
build the YAML file with some custom values, you will need the following pieces
of software installed:

1. [Tanka](https://github.com/grafana/tanka) >= v0.8
2. [`jsonnet-bundler`](https://github.com/jsonnet-bundler/jsonnet-bundler) >= v0.2.1

See the [`template` Tanka environment](./build/templates) for the current
settings that initialize the Grafana Agent Tanka configs. To build the YAML
file, execute the `./build/build.sh` script or run `make example-kubernetes`
from the project's root directory.
