---
aliases:
- /docs/agent/shared/flow/reference/components/otelcol-tls-config-block/
headless: true
---

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`ca_pem` | `string` | CA PEM-encoded text to validate the server with. | | no
`ca_file` | `string` | Path to the CA file. | | no
`cert_pem` | `string` | Certificate PEM-encoded text for client authentication. | | no
`cert_file` | `string` | Path to the TLS certificate. | | no
`key_pem` | `secret` | Key PEM-encoded text for client authentication. | | no
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

The following pairs of arguments are mutually exclusive and cannot both be set
simultaneously:

* `ca_pem` and `ca_file`
* `cert_pem` and `cert_file`
* `key_pem` and `key_file`