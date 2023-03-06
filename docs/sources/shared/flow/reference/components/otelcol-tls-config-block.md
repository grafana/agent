---
aliases:
- /docs/agent/shared/flow/reference/components/otelcol-tls-config-block/
headless: true
---

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`ca_file` | `string` | Path to the CA file. | | no
`cert_file` | `string` | Path to the TLS certificate. | | no
`key_file` | `string` | Path to the TLS certificate key. | | no
`min_version` | `string` | Minimum acceptable TLS version for connections. | `"TLS 1.2"` | no
`max_version` | `string` | Maximum acceptable TLS version for connections. | `"TLS 1.3"` | no
`reload_interval` | `duration` | The duration after which the certificate will be reloaded. | `"0s"` | no
`insecure` | `boolean` | Disables TLS when connecting to the configured server. | | no
`insecure_skip_verify` | `boolean` | Ignores insecure server TLS certificates. | | no
`server_name` | `string` | Verifies the hostname of server certificates when set. | | no

If the server doesn't support TLS, the tls block must be provided with the
`insecure` argument set to `true`. To disable `tls` for connections to the
server, set the `insecure` argument to `true`.

If `reload_interval` is set to `"0s"`, the certificate will never be reloaded.