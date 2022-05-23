# Grafana Agent Flow Prototype

> **NOTE**: Grafana Agent Flow is a prototype in active development and
> everything is subject to breaking changes; don't use this in production.

`cmd/agentflow` is the entrypoint for the experimental Grafana Agent Flow
prototype. It is presented as a separate command while the prototype is still
being developed. Support for Flow will eventually be added directly into
`cmd/agent` and this command will be removed.

Grafana Agent Flow is a component-based reimagining of Grafana Agent, where
units of logic are broken up into "components" which can independently
configured and wired together by the user.

Grafana Agent Flow currently uses HCL for its configuration language rather
than the YAML used by the existing project.

See the package-level comments in the [component package][] for information on
how to write new components.

## Running

You can run the Grafana Agent Flow prototype from the root of the repository
with:

```
go run ./cmd/agentflow -config.file ./cmd/agentflow/example-config.flow
```

This starts Grafana Agent Flow with the provided [example config file][].

## Reloading

Agent Flow can reload its config file by sending a `POST` request to
`/-/reload` against Flow's HTTP server.

The default HTTP server address is `http://127.0.0.1:12345` and can be modified
with the `-server.http-listen-addr` flag.

[example config file]: ./example-config.flow
[component package]: ../../component/component.go
