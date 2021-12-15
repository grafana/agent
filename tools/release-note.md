This is release `${RELEASE_TAG}` of the Grafana Agent.

### Upgrading
Read the [migration guide](https://github.com/grafana/agent/blob/${RELEASE_TAG}/docs/upgrade-guide/_index.md) for specific instructions on upgrading from older versions.

### Notable changes:
:warning: **ADD RELEASE NOTES HERE** :warning:


### Installation:
Grafana Agent is currently distributed in plain binary form, Docker container images, a Windows installer, and a Kubernetes install script. Choose whichever fits your use-case best.

#### Kubernetes 

Install directions [here.](https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s/)

#### Docker container:

* https://hub.docker.com/r/grafana/agent

```bash
docker pull "grafana/agent:${RELEASE_TAG}"
```

#### Windows installer

The Windows installer is provided as a [release asset](https://github.com/grafana/agent/releases/download/${RELEASE_TAG}/grafana-agent-installer.exe) for x64 machines.

#### Binary

We provide precompiled binary executables for the most common operating systems. Choose from the assets below for your matching operating system. Example for the `linux` operating system on `amd64`:

```bash
# download the binary
curl -O -L "https://github.com/grafana/agent/releases/download/${RELEASE_TAG}/agent-linux-amd64.zip"

# extract the binary
unzip "agent-linux-amd64.zip"

# make sure it is executable
chmod a+x "agent-linux-amd64"
```

#### `agentctl`

`agentctl`, a tool for helping you interact with the Agent, is available as a Docker image:

```bash
docker pull "grafana/agentctl:${RELEASE_TAG}"
```

Or as a binary. Like before, choose the assets below that matches your operating system. For example, with `linux` on `amd64`:

```bash
# download the binary
curl -O -L "https://github.com/grafana/agent/releases/download/${RELEASE_TAG}/agentctl-linux-amd64.zip"

# extract the binary
unzip "agentctl-linux-amd64.zip"

# make sure it is executable
chmod a+x "agentctl-linux-amd64"
```

#### `agent-operator`

`agent-operator`, a Kubernetes Operator for the Grafana Agent, is available only as a Docker image:

```bash
docker pull "grafana/agent-operator:${RELEASE_TAG}"
```
