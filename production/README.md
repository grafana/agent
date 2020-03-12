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

## Running the Agent locally

Currently, you must provide your own system configuration files to run the
Agent as a long-living process (e.g., write your own systemd unit files).

## Use the example Kubernetes configs

## Build the Agent from source

## Use our production Tanka configs


