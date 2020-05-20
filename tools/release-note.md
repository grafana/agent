This is release `${RELEASE_TAG}` of the Grafana Cloud Agent.

### Notable changes:
:warning: **ADD RELEASE NOTES HERE** :warning:


### Installation:
Grafana Cloud Agent is currently distributed in plain binary form, Docker
container images, and a Kubernetes install script. Choose whichever fits your
use-case best.

#### Kubernetes Install Script

```
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/grafana/agent/${RELEASE_TAG}/production/kubernetes/install.sh)" | kubectl apply -f -
```

#### Docker container:

* https://hub.docker.com/r/grafana/agent

```bash
docker pull "grafana/agent:${RELEASE_TAG}"
```

#### Binary

We provide precompiled binary executables for the most common operating systems.
Choose from the assets below for your matching operating system. Example for the
`linux` operating system on `amd64`:

```bash
# download the binary
curl -O -L "https://github.com/grafana/agent/releases/download/${RELEASE_TAG}/agent-linux-amd64.zip"

# extract the binary
unzip "agent-linux-amd64.zip"

# make sure it is executable
chmod a+x "agent-linux-amd64"
```

#### Agentctl

Agentctl, a tool for helping you interact with the Agent,
is available as a Docker image:

```bash
docker pull "grafana/agentctl:${RELEASE_TAG}"
```

Or as a binary. Like before, choose the assets below that matches your
operating system. For example, with `linux` on `amd64`:

```bash
# download the binary
curl -O -L "https://github.com/grafana/agent/releases/download/${RELEASE_TAG}/agentctl-linux-amd64.zip"

# extract the binary
unzip "agentctl-linux-amd64.zip"

# make sure it is executable
chmod a+x "agentctl-linux-amd64"
```
