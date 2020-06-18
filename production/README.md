# Running Grafana Cloud Agent

There are a few ways to run the Grafana Cloud Agent, in order from
easiest to hardest:

- [Use the Install Script for Kubernetes](#install-script-for-kubernetes)
- [Run the Agent with Docker](#running-the-agent-with-docker)
- [Run the Agent locally](#running-the-agent-locally)
- [Use the example Kubernetes configs](#use-the-example-kubernetes-configs)
- [Build the Agent from Source](#build-the-agent-from-source)
- [Use our production Tanka configs](#use-our-production-tanka-configs)

## Install Script for Kubernetes

The Grafana Cloud Agent repository comes with an installation script to
configure remote write and return a Kubernetes manifest that uses our preferred
defaults. To run the script, copy and paste this line in your terminal:

```
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/master/production/kubernetes/install.sh)" | kubectl apply -f -
```

See the [Kubernetes README](./kubernetes/README.md) for more information.

## Running the Agent with Docker

To run the Agent with Docker, you should have a configuration file on
your local machine ready to bind mount into the container. Then modify
the following command for your environment. Replace `/path/to/config.yaml` with
the full path to your YAML configuration, and replace `/tmp/agent` with the
directory on your host that you want the agent to store its WAL.

```
docker run \
  -v /tmp/agent:/etc/agent \
  -v /path/to/config.yaml:/etc/agent-config/agent.yaml \
  --entrypoint "/bin/agent -config.file=/etc/agent-config/agent.yaml -prometheus.wal-directory=/etc/agent/data"
  grafana/agent:v0.4.0
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
to generate the Kubernetes configs for the install script.
