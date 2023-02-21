This is release `${VERSION}` of the Grafana Agent.

### Upgrading
Read the [upgrade guide](https://grafana.com/docs/agent/${RELEASE_DOC_TAG}/upgrade-guide) for specific instructions on upgrading from older versions.

### Notable changes:
:warning: **ADD RELEASE NOTES HERE** :warning:


### Installation:
Grafana Agent is currently distributed in plain binary form, Docker container images, a Windows installer, and a Kubernetes install script. Choose whichever fits your use-case best.

#### Kubernetes

Install directions [here.](https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s/)

#### Docker container:

* https://hub.docker.com/r/grafana/agent

Docker containers are published as `grafana/agent:${VERSION}`. For Windows Docker containers, use `grafana/agent:${VERSION}-windows` instead. 

#### Windows installer

The Windows installer is provided as a [release asset](https://github.com/grafana/agent/releases/download/${VERSION}/grafana-agent-installer.exe.zip) for x64 machines.

#### Binary

We provide precompiled binary executables for the most common operating systems. Choose from the assets below for your matching operating system.

Note: ppc64le builds are currently considered secondary release targets and do not have the same level of support and testing as other platforms.

Example for the `linux` operating system on `amd64`:

```bash
# download the binary
curl -O -L "https://github.com/grafana/agent/releases/download/${VERSION}/grafana-agent-linux-amd64.zip"

# extract the binary
unzip "grafana-agent-linux-amd64.zip"

# make sure it is executable
chmod a+x "grafana-agent-linux-amd64"
```

#### `agentctl`

`agentctl`, a tool for helping you interact with the Agent, is available as a Docker image:

Docker containers are published as `grafana/agentctl:${VERSION}`. For Windows Docker containers, use `grafana/agentctl:${VERSION}-windows` instead. 

Or as a binary. Like before, choose the assets below that matches your operating system. For example, with `linux` on `amd64`:

```bash
# download the binary
curl -O -L "https://github.com/grafana/agent/releases/download/${VERSION}/grafana-agentctl-linux-amd64.zip"

# extract the binary
unzip "grafana-agentctl-linux-amd64.zip"

# make sure it is executable
chmod a+x "grafana-agentctl-linux-amd64"
```

#### `agent-operator`

`agent-operator`, a Kubernetes Operator for the Grafana Agent, is available only as a Docker image:

```bash
docker pull "grafana/agent-operator:${VERSION}"
```
