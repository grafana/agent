---
aliases:
- /docs/agent/shared/flow/reference/components/tls-config-block/
headless: true
---

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`ca_file` | `string` | CA certificate to validate the server with. | | no
`cert_file` | `string` | Certificate file for client authentication. | | no
`key_file` | `string` | Key file for client authentication. | | no
`server_name` | `string` | ServerName extension to indicate the name of the server. | | no
`insecure_skip_verify` | `bool` | Disables validation of the server certificate. | | no
`min_version` | `string` | Minimum acceptable TLS version. | | no

When `min_version` is not provided, the minimum acceptable TLS version is
inherited from Go's default minimum version, TLS 1.2. If `min_version` is
provided, it must be set to one of the following strings:

* `"TLS10"` (TLS 1.0)
* `"TLS11"` (TLS 1.1)
* `"TLS12"` (TLS 1.2)
* `"TLS13"` (TLS 1.3)
