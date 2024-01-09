---
aliases:
- ../../configuration/flags/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/flags/
- /docs/grafana-cloud/send-data/agent/static/configuration/flags/
canonical: https://grafana.com/docs/agent/latest/static/configuration/flags/
description: Learn about command-line flags
title: Command-line flags
weight: 100
---

# Command-line flags

Command-line flags are used to configure settings of Grafana Agent which cannot
be updated at runtime.

All flags may be prefixed with either one hyphen or two (i.e., both
`-config.file` and `--config.file` are valid).

> Note: There may be flags returned by `-help` which are not listed here; this
> document only lists flags that do not have an equivalent in the YAML file.

## Basic

* `-version`: Print out version information
* `-help`: Print out help

## Experimental feature flags

Grafana Agent has some experimental features that require being enabled through
an `-enable-features` flag. This flag takes a comma-delimited list of feature
names to enable.

Valid feature names are:

* `remote-configs`: Enable [retrieving][retrieving] config files over HTTP/HTTPS
* `integrations-next`: Enable [revamp][revamp] of the integrations subsystem
* `extra-scrape-metrics`: When enabled, additional time series  are exposed for each metrics instance scrape. See [Extra scrape metrics](https://prometheus.io/docs/prometheus/2.45/feature_flags/#extra-scrape-metrics).
* `agent-management`: Enable support for [agent management][management].

## Report information usage

By default, Grafana Agent sends anonymous, but uniquely-identifiable usage information
from your running Grafana Agent instance to Grafana Labs.
These statistics are sent to `stats.grafana.org`.

Statistics help us better understand how Grafana Agent is used.
This helps us prioritize features and documentation.

The usage information includes the following details:
* A randomly generated and an anonymous unique ID (UUID).
* Timestamp of when the UID was first generated.
* Timestamp of when the report was created (by default, every 4h).
* Version of running Grafana Agent.
* Operating system Grafana Agent is running on.
* System architecture Grafana Agent is running on.
* List of enabled feature flags.
* List of enabled integrations.

This list may change over time. All newly reported data will also be documented in the CHANGELOG.

If you would like to disable the reporting, Grafana Agent provides the flag `-disable-reporting`
to stop the reporting.

## Support bundles
Grafana Agent allows the exporting of 'support bundles' on the `/-/support`
endpoint. Support bundles are zip files containing commonly-used information
that provide a baseline for debugging issues with the Agent.

Support bundles contain all information in plain text, so that they can be
inspected before sharing to verify that no sensitive information has leaked.

Support bundles contain the following data:
* `agent-config.yaml` contains the current agent configuration (when the `-config.enable-read-api` flag is passed).
* `agent-logs.txt` contains the agent logs during the bundle generation.
* `agent-metadata.yaml` contains the agent's build version, operating system, architecture, uptime, plus a string payload defining which extra agent features have been enabled via command-line flags.
* `agent-metrics-instances.json` and `agent-metrics-targets.json` contain the active metric subsystem instances, and the discovered scraped targets for each one.
* `agent-logs-instances.json` and `agent-logs-targets.json` contains the active logs subsystem instances and the discovered log targets for each one.
* `agent-metrics.txt` contains a snapshot of the agent's internal metrics.
* The `pprof/` directory contains Go runtime profiling data (CPU, heap, goroutine, mutex, block profiles) as exported by the pprof package.

To disable the endpoint that exports these support bundles, you can pass in the
`-disable-support-bundle` command-line flag.

## Configuration file

* `-config.file`: Path to the configuration file to load. May be an HTTP(s) URL when the `remote-configs` feature is enabled.
* `-config.file.type`: Type of file which `-config.file` refers to (default `yaml`). Valid values are `yaml` and `dynamic`.
* `-config.expand-env`: Expand environment variables in the loaded configuration file
* `-config.enable-read-api`: Enables the `/-/config` and `/agent/api/v1/configs/{name}` API endpoints to print YAML configuration

### Remote Configuration

These flags require the `remote-configs` feature to be enabled:

`-config.url.basic-auth-user`: Basic Authentication username to use when fetching the remote configuration file
`-config.url.basic-auth-password-file`: File containing a Basic Authentication password to use when fetching the remote configuration file

## Server

* `-server.register-instrumentation`: Expose the `/metrics` and `/debug/pprof/` instrumentation handlers over HTTP (default true)
* `-server.graceful-shutdown-timeout`: Timeout for a graceful server shutdown
* `-server.log.source-ips.enabled`: Whether to log IP addresses of incoming requests
* `-server.log.source-ips.header`: Header field to extract incoming IP requests from (defaults to Forwarded, X-Real-IP, X-Forwarded-For)
* `-server.log.source-ips.regex`: Regex to extract the IP out of the read header, using the first capture group as the IP address
* `-server.http.network`: HTTP server listen network (default `tcp`)
* `-server.http.address`: HTTP server listen:port (default `127.0.0.1:12345`)
* `-server.http.enable-tls`: Enable TLS for the HTTP server
* `-server.http.conn-limit`: Maximum number of simultaneous HTTP connections
* `-server.http.idle-timeout`: HTTP server idle timeout
* `-server.http.read-timeout`: HTTP server read timeout
* `-server.http.write-timeout`: HTTP server write timeout
* `-server.http.in-memory-addr`: Internal address used for the agent to make
  in-memory HTTP connections to itself. (default `agent.internal:12345`) The
  port number specified here is virtual and does not open a real network port.
* `-server.grpc.network` gRPC server listen network (default `grpc`)
* `-server.grpc.address`: gRPC server listen host:port (default `127.0.0.1:12346`)
* `-server.grpc.enable-tls`: Enable TLS for the gRPC server
* `-server.grpc.conn-limit`: Maximum number of simultaneous gRPC connections
* `-server.grpc.keepalive.max-connection-age` Maximum age for any gRPC connection for a graceful shutdown
* `-server.grpc.keepalive.max-connection-age-grace` Grace period to forcibly close connections after a graceful shutdown starts
* `-server.grpc.keepalive.max-connection-idle` Time to wait before closing idle gRPC connections
* `-server.grpc.keepalive.min-time-between-pings` Maximum frequency that clients may send pings at
* `-server.grpc.keepalive.ping-without-stream-allowed` Allow clients to send pings without having a gRPC stream
* `-server.grpc.keepalive.time` Frequency to send keepalive pings from the server
* `-server.grpc.keepalive.timeout` How long to wait for a keepalive pong before closing the connection
* `-server.grpc.max-concurrent-streams` Maximum number of concurrent gRPC streams (0 = unlimited)
* `-server.grpc.max-recv-msg-size-bytes` Maximum size in bytes for received gRPC messages
* `-server.grpc.max-send-msg-size-bytes` Maximum size in bytes for send gRPC messages
* `-server.grpc.in-memory-addr`: Internal address used for the agent to make
  in-memory gRPC connections to itself. (default `agent.internal:12346`). The
  port number specified here is virtual and does not open a real network port.

### TLS Support

TLS support can be enabled with `-server.http.tls-enabled` and
`-server.grpc.tls-enabled` for the HTTP and gRPC servers respectively.

`server.http_tls_config` and `integrations.http_tls_config` must be set in the
YAML configuration when the `-server.http.tls-enabled` flag is used.

`server.grpc_tls_config` must be set in the YAML configuration when the
`-server.grpc.tls-enabled` flag is used.

## Metrics

* `-metrics.wal-directory`: Directory to store the metrics Write-Ahead Log in

{{% docs/reference %}}
[retrieving]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration#remote-configuration-experimental"
[retrieving]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/static/configuration#remote-configuration-experimental"

[revamp]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/integrations/integrations-next/"
[revamp]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/static/configuration/integrations/integrations-next"

[management]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/agent-management"
[management]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/static/configuration/agent-management"
{{% /docs/reference %}}
