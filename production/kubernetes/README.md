# Kubernetes Config

This directory contains Kubernetes manifest templates and installation scripts
for rendering the templates so they can be applied against Kubernetes.

Manifests:

- Metric collection: [`agent.yaml`](./agent.yaml)
- Log collection: [`agent-loki.yaml`](./agent-loki.yaml)

Installation script:

- Metric collection: [`install.sh`](./install.sh)
- Log collection: [`install-loki.sh`](./install-loki.sh)

## Install Scripts

There are two installation scripts, one for metrics and the other for logs. Each
install script does the following:

1. Prmopts the user for their remote target credentials (Prometheus remote_write, Loki client).
2. Downloads the manifest template from GitHub
3. Substitutes variables in the template with the provided input from
   step 1.
4. Prints out the final manifest to stdout without applying it.

Here's a two-line script to copy and paste to install the Agent on
Kubernetes for collecting metrics and logs (requires `envsubst` (GNU gettext)):

```
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install.sh)" | kubectl -ndefault apply -f -
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/release/production/kubernetes/install-loki.sh)" | kubectl -ndefault apply -f -
```

## Manually Applying

Since the manifest files are just templates, note that they are *not* ready for
applying out of the box and you will have to manually reroduce the steps that
the installation script does:

1. Download the manifest as `manifest.yaml`.

2. Modify your copy of the manifest, replacing all variables with the
   appropriate values:

   1. For the metrics collection manifest, replace `${REMOTE_WRITE_URL}` with
      the full endpoint of the Prometheus remote write API. For logs collection,
      replace `${LOKI_HOSTNAME}` with the hostname of the Loki API. Unlike the
      remote write API, `${LOKI_HOSTNAME}` should _only_ be the hostname, such
      as `localhost` or `logs-us-central1.grafana.net`.

  2. Replace `${REMOTE_WRITE_PASSWORD}` or `${LOKI_PASSWORD}` with the password
     for authentication against the remote API. If you do not need
     authentication, remove the entire authentication section.

  3. If you did not remove the authentication section from the previous step,
     replace `${REMOTE_WRITE_USERNAME}` or `${LOKI_USERNAME}` with the username
     used to connect to the remote API.

3. Apply the modified manifest file: `kubectl -ndefault apply -f manifest.yaml`.

## Rebuilding the manifests

The manifests provided are created using Grafana Labs' production
[Tanka configs](../tanka/grafana-agent) with some default values. If you want to
build the YAML file with some custom values, you will need the following pieces
of software installed:

1. [Tanka](https://github.com/grafana/tanka) >= v0.8
2. [`jsonnet-bundler`](https://github.com/jsonnet-bundler/jsonnet-bundler) >= v0.2.1

See the [`template` Tanka environment](./build/template) for the current
settings that initialize the Grafana Agent Tanka configs. To build the YAML
file, execute the `./build/build.sh` script or run `make example-kubernetes`
from the project's root directory.
