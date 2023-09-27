---
aliases:
- /docs/agent/shared/flow/reference/components/otelcol-grpc-authority
- /docs/grafana-cloud/agent/shared/flow/reference/components/otelcol-grpc-authority/
- /docs/grafana-cloud/monitor-infrastructure/agent/shared/flow/reference/components/otelcol-grpc-authority/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/shared/flow/reference/components/otelcol-grpc-authority/
description: Shared content, otelcol grpc authority
headless: true
---

The `authority` in gRPC represents the "Host" header for requests. By default, the `authority` is derived from the service URL used for the gRPC call. It's primarily used for virtual hosting, routing decisions, and security validations. Overriding this default value is be done using the [WithAuthority][] dial option. This allows users to simulate production behaviors in development environments, guide traffic in service mesh scenarios, and test server responses for different authority values.

[WithAuthority]: https://pkg.go.dev/google.golang.org/grpc#WithAuthority
