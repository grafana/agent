---
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
* `--cluster.join-addresses`: Comma-separated list of addresses to join the cluster at (default `""`).
* `--cluster.advertise-address`: Address to advertise to other cluster nodes (default `""`).
* `--config.format`: The format of the source file. Supported formats: 'flow', 'prometheus' (default `"flow"`).
* `--config.bypass-conversion-warnings`: Enable bypassing warnings when converting (default `false`).

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

## Clustered mode (experimental)

When the `--cluster.enabled` command-line argument is provided, Grafana Agent will
start in _clustered mode_.

The agent tries to connect over HTTP/2 to one or more peers provided in the
comma-separated `--cluster.join-addresses` list to join an existing cluster.
If no connection can be made or the argument is empty, the agent falls back to
bootstrapping a new cluster of its own.

The agent will advertise its own address as `--cluster.advertise-address` to
other agent nodes; if this is empty it will attempt to find a suitable address
to advertise from a list of default network interfaces. The agent must be
reachable over HTTP on this address as communication happens over the agent's
HTTP server.

## Configuration conversion (beta)

When the `--config.format` command-line argument is provided with a value
other than `flow`, Grafana Agent will convert the config file from the source
format to River and immediately start running with it. This leverages the same
converter API described in the [grafana-agent convert][] docs.

If the `--config.bypass-conversion-warnings` command-line argument is also provided,
Grafana Agent will ignore any warnings provided by the converter. This should
be used with caution since the resulting conversion is not equivalent to the
original config.

[grafana-agent convert]: {{< relref "./convert.md" >}}