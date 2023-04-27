---
title: agent run
weight: 100
---

# `agent run` command

The `agent run` command runs Grafana Agent Flow in the foreground until an
interrupt is received.

## Usage

Usage: `agent run [FLAG ...] FILE_NAME`

`agent run` must be provided an argument which points at the River config file
to use. `agent run` will immediately exit with an error if the River file
wasn't specified, can't be loaded, or contained errors during the initial load.

Grafana Agent Flow will continue to run if subsequent reloads of the config
file fail, potentially marking components as unhealthy depending on the nature
of the failure. When this happens, Grafana Agent Flow will continue functioning
in the last valid state.

`agent run` launches an HTTP server for expose metrics about itself and
components. The HTTP server is also used for exposing a UI at `/` for debugging
running components.

The following flags are supported:

* `--server.http.listen-addr`: Address to listen for HTTP traffic on (default `127.0.0.1:12345`).
* `--server.http.ui-path-prefix`: Base path where the UI will be exposed (default `/`).
* `--storage.path`: Base directory where components can store data (default `data-agent/`).
* `--disable-reporting`: Disable [usage reporting][] of enabled [components][] to Grafana (default `false`).

[usage reporting]: {{< relref "../../../configuration/flags.md/#report-information-usage" >}}
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
