+++
title  = "Command-line flags"
weight = 100
+++

# Command-line flags

Command-line flags are used to configure settings of Grafana Agent which cannot
be updated at runtime.

All flags may be prefixed with either one hypen or two (i.e., both
`-config.file` and `--config.file` are valid).

> Note: There may be flags returned by `-help` which are not listed here; this
> document only lists flags that do not have an equivalent in the YAML file.

## Basic

* `-version`: Print out version information
* `-help`: Print out help

## Experiemntal feature flags

Grafana Agent has some experimental features that require being enabled through
an `-enable-features` flag. This flag takes a comma-delimited list of feature
names to enable.

Valid feature names are:

* `remote-configs`: Enable [retrieving]({{< relref "./_index.md#remote-configuration-experimental" >}}) config files over HTTP/HTTPS
* `integrations-next`: Enable [revamp]({{< relref "./integrations/integrations-next/" >}}) of the integrations subsystem
* `dynamic-config`: Enable support for [dynamic configuration]({{< relref "./dynamic-config" >}})

## Configuration file

* `-config.file`: Path to the configuration file to load. May be an HTTP(s) URL when the `remote-configs` feature is enabled
* `-config.file.type`: Type of file which `-config.file` refers to (default `yaml`). Valid values are `yaml` and `dynamic`.
* `-config.expand-env`: Expand environment variables in the loaded configuration file
* `-config.enable-read-api`: Enables the `/-/config` and `/agent/api/v1/configs/{name}` API endpoints to print YAML configuration

### Remote Configuration

These flags require the `remote-configs` feature to be enabled:

`-config.url.basic-auth-user`: Basic Authentication username to use when fetching the remote configuration file
`-config.url.basic-auth-password-file`: File containing a Basic Authentication password to use when fetching the remote configuration file

### Dynamic Configuration

The `dynamic-config` and `integrations-next` features must be enabled when
`-config.file.type` is set to `dynamic`.

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
* `-server.grpc.network` gRPC server listen network (default `grpc`)
* `-server.grpc.address`: gRPC server listen host:port (default `127.0.0.1:12346`)
* `-server.grpc.enable-tls`: Enable TLS for the gRPC server
* `-server.grpc.conn-limit`: Maximum number of simultaneous gRPC connections
* `-server.grpc.keepalive.max-connection-age` Maximum age for any gRPC connection for a graceful shutdown
* `-server.grpc.keepalive.max-connection-age-grace` Grace period to forceibly close connections after a graceful shutdown starts
* `-server.grpc.keepalive.max-connection-idle` Time to wait before closing idle gRPC connections
* `-server.grpc.keepalive.min-time-between-pings` Maximum frequency that clients may send pings at
* `-server.grpc.keepalive.ping-without-stream-allowed` Allow clients to send pings without having a gRPC stream
* `-server.grpc.keepalive.time` Frequency to send keepalive pings from the server
* `-server.grpc.keepalive.timeout` How long to wait for a keepalive pong before closing the connection
* `-server.grpc.max-concurrent-streams` Maximum number of concurrent gRPC streams (0 = unlimited)
* `-server.grpc.max-recv-msg-size-bytes` Maximum size in bytes for received gRPC messages
* `-server.grpc.max-send-msg-size-bytes` Maximum size in bytes for send gRPC messages

### TLS Support

TLS support can be enabled with `-server.http.tls-enabled` and
`-server.grpc.tls-enabled` for the HTTP and gRPC servers respectively.

`server.http_tls_config` and `integrations.http_tls_config` must be set in the
YAML configuration when the `-server.http.tls-enabled` flag is used.

`server.grpc_tls_config` must be set in the YAML configuration when the
`-server.grpc.tls-enabled` flag is used.

## Metrics

* `-metrics.wal-directory`: Directory to store the metrics Write-Ahead Log in
