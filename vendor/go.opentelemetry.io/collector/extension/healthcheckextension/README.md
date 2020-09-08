# Health Check

Health Check extension enables an HTTP url that can be probed to check the
status of the the OpenTelemetry Collector. This extension can be used as a
liveness and/or readiness probe on Kubernetes.

The following settings are required:

- `port` (default = 13133): What port to expose HTTP health information.

Example:

```yaml
extensions:
  health_check:
```

The full list of settings exposed for this exporter is documented [here](./config.go)
with detailed sample configurations [here](./testdata/config.yaml).
