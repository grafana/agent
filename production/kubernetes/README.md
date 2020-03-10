# Kubernetes Config

This directory contains an `agent.yaml` file that can be used to deploy the
Grafana Cloud Agent. The Grafana Cloud Agent install script utilizes this
file to create the final manifests to deploy.

Note that without using the install script, the file is *not* ready for applying
out of the box and you will have to manually reproduce the steps that the
install script follows:

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
[Tanka configs](../tanka/grafana-agent) with some default values. If you want to build the YAML file with some custom values, you will need the
following pieces of software installed:

1. Tanka
2. `jsonnet-bundler`

See the [`template` environment](./build/template) for the current settings
that initialize the Grafana Agent tanka configs. To build the YAML file,
execute the `./build/build.sh` script or run `make example-kubernetes` from the
project's root directory.
