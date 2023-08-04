---
canonical: https://grafana.com/docs/agent/latest/flow/reference/cli/run/
title: grafana-agent run
weight: 100
---

# `grafana-agent run` command

The `grafana-agent run` command runs Grafana Agent Flow in the foreground until an
interrupt is received.

## Usage

Usage: `grafana-agent run [FLAG ...] FILE_NAME`

`grafana-agent run` must be provided an argument which points at the River config file
to use. `grafana-agent run` will immediately exit with an error if the River file
wasn't specified, can't be loaded, or contained errors during the initial load.

Grafana Agent Flow will continue to run if subsequent reloads of the config
file fail, potentially marking components as unhealthy depending on the nature
of the failure. When this happens, Grafana Agent Flow will continue functioning
in the last valid state.

`grafana-agent run` launches an HTTP server for expose metrics about itself and
components. The HTTP server is also used for exposing a UI at `/` for debugging
running components.

The following flags are supported:

* `--server.http.enable-pprof`: Enable /debug/pprof profiling endpoints. (default `true`)
* `--server.http.memory-addr`: Address to listen for [in-memory HTTP traffic][] on
  (default `agent.internal:12345`).
* `--server.http.listen-addr`: Address to listen for HTTP traffic on (default `127.0.0.1:12345`).
* `--server.http.ui-path-prefix`: Base path where the UI will be exposed (default `/`).
* `--storage.path`: Base directory where components can store data (default `data-agent/`).
* `--disable-reporting`: Disable [usage reporting][] of enabled [components][] to Grafana (default `false`).
* `--cluster.enabled`: Start the Agent in clustered mode (default `false`).
* `--cluster.node-name`: The name to use for this node (defaults to the environment's hostname).
* `--cluster.join-addresses`: Comma-separated list of addresses to join the cluster at (default `""`). Mutually exclusive with `--cluster.discover-peers`.
* `--cluster.discover-peers`: List of key-value tuples for discovering peers (default `""`). Mutually exclusive with `--cluster.join-addresses`.
* `--cluster.advertise-address`: Address to advertise to other cluster nodes (default `""`).
* `--config.format`: The format of the source file. Supported formats: `flow`, `prometheus`, `promtail` (default `"flow"`).
* `--config.bypass-conversion-errors`: Enable bypassing errors when converting (default `false`).

[in-memory HTTP traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[usage reporting]: {{< relref "../../../static/configuration/flags.md#report-information-usage" >}}
[components]: {{< relref "../../concepts/components.md" >}}

## Updating the config file

The config file can be reloaded from disk by either:

* Sending an HTTP POST request to the `/-/reload` endpoint.
* Sending a `SIGHUP` signal to the Grafana Agent process.

When this happens, the [component controller][] synchronizes the set of running
components with the latest set of components specified in the config file.
Components that are no longer defined in the config file after reloading are
shut down, and components that have been added to the config file since the
previous reload are created.

All components managed by the component controller are reevaluated after
reloading.

[component controller]: {{< relref "../../concepts/component_controller.md" >}}

## Clustering (beta)

The `--cluster.enabled` command-line argument starts Grafana Agent in
[clustering][] mode. The rest of the `--cluster.*` command-line flags can be
used to configure how nodes discover and connect to one another.

Each cluster member’s name must be unique within the cluster. Nodes which try
to join with a conflicting name are rejected and will fall back to
bootstrapping a new cluster of their own.

Peers communicate over HTTP/2 on the agent's built-in HTTP server. Each node
must be configured to accept connections on `--server.http.listen-addr` and the
address defined or inferred in `--cluster.advertise-address`.

If the `--cluster.advertise-address` flag is not explicitly set, the agent
tries to infer a suitable one from the `eth0` and `en0` local network
interfaces. If the advertised address cannot be determined, the agent will
fail to start. Windows users must explicitly pass a value for
`--cluster.advertise-address`, since Windows does not have interfaces name
`eth0` or `en0`.

The comma-separated list of addresses provided in `--cluster.join-addresses`
can either be IP addresses with an optional port, or DNS records to lookup.
The ports on the list of addresses default to the port used for the HTTP
listener if not explicitly provided. We recommend that you
align the port numbers on as many nodes as possible to simplify the deployment
process.

The `--cluster.discover-peers` command-line flag expects a list of tuples in
the form of `provider=XXX key=val key=val ...`. Clustering makes use of the
[go-discover] package to discover peers and fetch their IP addresses, based
on the chosen provider and the filtering key-values it supports. Clustering
supports the default set of providers available in go-discover and registers
the `k8s` provider on top.

If either the key or the value in a pair contains a space, a backslash or
double quotes, then it must be quoted with double quotes. Within this quoted
string, the backslash can be used to escape double quotes or the backslash
itself.

Discovering peers using the `--cluster.join-addresses` and
`--cluster.discover-peers` flags only happens on startup; after that, cluster
nodes depend on gossiping messages with each other to converge on the cluster's
state.

The first node that is used to bootstrap a new cluster (also known as
the "seed node") can either omit the flag that specifies peers to join or can
try to connect to itself.

### Clustering states

Clustered agents are in one of three states:

* **Viewer**: The agent has a read-only view of the cluster and is not
  participating in workload distribution.

* **Participant**: The agent is participating in workload distribution for
  components which have clustering enabled.

* **Terminating**: The agent is shutting down, and will no longer assign new
  work to itself.

Agents initially join the cluster in the viewer state, and then transition to
the participant state after the processs startup completes. Agents then
transition to the terminating state when shutting down.

The current state of a clustered agent is shown on the clustering page in the
[UI][].

[UI]: {{< relref "../../monitoring/debugging.md#clustering-page" >}}

## Configuration conversion (beta)

When you use the `--config.format` command-line argument with a value
other than `flow`, Grafana Agent converts the configuration file from
the source format to River and immediately starts running with the new
configuration. This conversion uses the converter API described in the
[grafana-agent convert][] docs.

If you also use the `--config.bypass-conversion-errors` command-line argument,
Grafana Agent will ignore any errors from the converter. Use this argument
with caution because the resulting conversion may not be equivalent to the
original configuration.

[grafana-agent convert]: {{< relref "./convert.md" >}}
[clustering]:  {{< relref "../../concepts/clustering.md" >}}
[go-discover]: https://github.com/hashicorp/go-discover
