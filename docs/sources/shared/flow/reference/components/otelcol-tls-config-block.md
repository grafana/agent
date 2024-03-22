---
aliases:
- /docs/agent/shared/flow/reference/components/otelcol-tls-config-block/
- /docs/grafana-cloud/agent/shared/flow/reference/components/otelcol-tls-config-block/
- /docs/grafana-cloud/monitor-infrastructure/agent/shared/flow/reference/components/otelcol-tls-config-block/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/shared/flow/reference/components/otelcol-tls-config-block/
- /docs/grafana-cloud/send-data/agent/shared/flow/reference/components/otelcol-tls-config-block/
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/otelcol-tls-config-block/
description: Shared content, otelcol tls config block
headless: true
---

The following arguments are supported:

Name                           | Type           | Description                                                                                  | Default     | Required
-------------------------------|----------------|----------------------------------------------------------------------------------------------|-------------|---------
`ca_file`                      | `string`       | Path to the CA file.                                                                         |             | no
`ca_pem`                       | `string`       | CA PEM-encoded text to validate the server with.                                             |             | no
`cert_file`                    | `string`       | Path to the TLS certificate.                                                                 |             | no
`cert_pem`                     | `string`       | Certificate PEM-encoded text for client authentication.                                      |             | no
`insecure_skip_verify`         | `boolean`      | Ignores insecure server TLS certificates.                                                    |             | no
`include_system_ca_certs_pool` | `boolean`      | Whether to load the system certificate authorities pool alongside the certificate authority. | `false`     | no
`insecure`                     | `boolean`      | Disables TLS when connecting to the configured server.                                       |             | no
`key_file`                     | `string`       | Path to the TLS certificate key.                                                             |             | no
`key_pem`                      | `secret`       | Key PEM-encoded text for client authentication.                                              |             | no
`max_version`                  | `string`       | Maximum acceptable TLS version for connections.                                              | `"TLS 1.3"` | no
`min_version`                  | `string`       | Minimum acceptable TLS version for connections.                                              | `"TLS 1.2"` | no
`cipher_suites`                | `list(string)` | A list of TLS cipher suites that the TLS transport can use.                                  | `[]`        | no
`reload_interval`              | `duration`     | The duration after which the certificate is reloaded.                                        | `"0s"`      | no
`server_name`                  | `string`       | Verifies the hostname of server certificates when set.                                       |             | no

If the server doesn't support TLS, you must set the `insecure` argument to `true`.

To disable `tls` for connections to the server, set the `insecure` argument to `true`.

If `reload_interval` is set to `"0s"`, the certificate never reloaded.

The following pairs of arguments are mutually exclusive and can't both be set simultaneously:

* `ca_pem` and `ca_file`
* `cert_pem` and `cert_file`
* `key_pem` and `key_file`

If `cipher_suites` is left blank, a safe default list is used.
See the [Go TLS documentation][golang-tls] for a list of supported cipher suites.

[golang-tls]: https://go.dev/src/crypto/tls/cipher_suites.go