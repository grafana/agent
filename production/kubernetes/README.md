# Kubernetes Config

This directory contains an [`agent.yaml`](./agent.yaml) template file
and an [install script](./install.sh) that renders the template
for application against Kubernetes.

## Install Script

The install script does the following:

1. Prmopts the user for their remote write URL, username, and password
2. Downloads `agent.yaml` from GitHub
3. Substitutes variables in the template with the provided input from
   step 1.
4. Prints out the final manifest to stdout without applying it.

Here's a one-line script to copy and paste to install the Agent on
Kubernetes:

```
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/v0.1.0/production/kubernetes/install.sh)" | kubectl apply -f -
```

## Manually Applying

Since the `agent.yaml` file is just a template, note that it is *not* ready for
applying out of the box and you'll have to manually reproduce the steps that the
install script does:

1. Download `agent.yaml` locally.

2. Modify your copy of `agent.yaml`, replacing the following strings with the
   appropriate values:

  1. Replace `${REMOTE_WRITE_URL}` with the full endpoint of the remote
     write API.

  2. Replace `${REMOTE_WRITE_PASSWORD}` with the password of the remote
     write API's authentication. If you do not need authentication to the
     remote write API, remove the entire `basic_auth` section, leaving just
     the URL.

  3. If you did not remove the `basic_auth` section from the previous step,
     replace `${REMOTE_WRITE_USERNAME}` with the username used to connect to
     the remote write API.

3. Apply the modified `agent.yaml` file: `kubectl apply -f agent.yaml`.

## Rebuilding the YAML file

The YAML file provided is created using Grafana Labs' production
[Tanka configs](../tanka/grafana-agent) with some default values. If you want to
build the YAML file with some custom values, you will need the following pieces
of software installed:

1. [Tanka](https://github.com/grafana/tanka) >= v0.8
2. [`jsonnet-bundler`](https://github.com/jsonnet-bundler/jsonnet-bundler) >= v0.2.1

See the [`template` Tanka environment](./build/template) for the current
settings that initialize the Grafana Agent Tanka configs. To build the YAML
file, execute the `./build/build.sh` script or run `make example-kubernetes`
from the project's root directory.
