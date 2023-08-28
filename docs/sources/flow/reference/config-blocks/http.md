---
canonical: https://grafana.com/docs/agent/latest/flow/reference/config-blocks/http/
title: http
---

# http block

`http` is an optional configuration block used to customize how Grafana Agent's
HTTP server functions. `http` is specified without a label and can only be
provided once per configuration file.

{{% admonition type="note" %}}
While the `http` block can reference component exports, some components that
rely on the HTTP server have a hidden dependency on the `http` block that may
result in a circular dependency error.

Only references to components named `remote.*` or `local.*` are guaranteed to
work without any circular dependency errors.
{{% /admonition %}}

## Example

```river
http {
  tls {
    cert_file = env("TLS_CERT_FILE_PATH")
    key_file  = env("TLS_KEY_FILE_PATH")
  }
}
```

## Arguments

The `http` block supports no arguments and is configured completely through
inner blocks.

## Blocks

The following blocks are supported inside the definition of `http`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
tls | [tls][] | Define TLS settings for the HTTP server. | no

[tls]: #tls-block

### tls block

The `tls` block configures TLS settings for the HTTP server.

{{% admonition type="warning" %}}
If the `tls` block is added and the configuration is reloaded when Grafana
Agent is running, existing connections will continue to communicate over
plaintext. Similarly, if the `tls` block is removed and the configuration is
reloaded when Grafana Agent is running, existing connections will continue to
communicate over TLS.

To ensure all connections use TLS, start Grafana Agent with the `tls` block
already configured.
{{% /admonition %}}

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`cert_pem` | `string` | PEM data of the server TLS certificate. | `""` | conditionally
`cert_file` | `string` | Path to the server TLS certificate on disk. | `""` | conditionally
`key_pem` | `string` | PEM data of the server TLS key. | `""` | conditionally
`key_file` | `string` | Path to the server TLS key on disk. | `""` | conditionally
`client_ca_pem` | `string` | PEM data of the client CA to validate requests against. | `""` | no
`client_ca_file` | `string` | Path to the client CA file on disk to validate requests against. | `""` | no
`client_auth` | `string` | Client authentication to use. | `"NoClientCert"` | no
`cipher_suites` | `list(string)` | Set of cipher suites to use. | `[]` | no
`curve_preferences` | `list(string)` | Set of elliptic curves to use in a handshake. | `[]` | no
`min_version` | `string` | Oldest TLS version to accept from clients. | `""` | no
`max_version` | `string` | Newest TLS version to accept from clients. | `""` | no

When the `tls` block is specified, arguments for the TLS certificate (using
`cert_pem` or `cert_file`) and for the TLS key (using `key_pem` or `key_file`)
are required.

The following pairs of arguments are mutually exclusive, and only one may be
configured at a time:

* `cert_pem` and `cert_file`
* `key_pem` and `key_file`
* `client_ca_pem` and `client_ca_file`

The `client_auth` argument determines whether to validate client certificates.
The default value, `NoClientCert`, indicates that no client certificate
validation is performed.

The following values are accepted for `client_auth`:

* `NoClientCert`: client certificates are neither requested nor verified.
* `RequestClientCert`: requests clients to send an optional certificate.
* `RequireAnyClientCert`: requires at least one certificate from clients that is not checked for validity.
* `VerifyClientCertIfGiven`: requests clients to send an optional certificate. If a certificate is sent, it must be valid.
* `RequireAndVerifyClientCert`: requires clients to send a valid certificate.

The `client_ca_pem` or `client_ca_file` arguments may be used to perform client
certificate validation. These arguments may only be provided when `client_auth`
is not set to `NoClientCert`.

The `cipher_suites` argument determines what cipher suites to use. If not
provided, a default list is used. The set of cipher suites specified may be
from the following:

* `TLS_RSA_WITH_AES_128_CBC_SHA`
* `TLS_RSA_WITH_AES_256_CBC_SHA`
* `TLS_RSA_WITH_AES_128_GCM_SHA256`
* `TLS_RSA_WITH_AES_256_GCM_SHA384`
* `TLS_AES_128_GCM_SHA256`
* `TLS_AES_256_GCM_SHA384`
* `TLS_CHACHA20_POLY1305_SHA256`
* `TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA`
* `TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA`
* `TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA`
* `TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA`
* `TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256`
* `TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384`
* `TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256`
* `TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384`
* `TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256`
* `TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256`

The `curve_preferences` argument determines the set of elliptic curves to
prefer during a handshake in preference order. If not provided, a default list
is used. The set of elliptic curves specified may be from the following:

* `CurveP256`
* `CurveP384`
* `CurveP521`
* `X25519`

The `min_version` and `max_version` arguments determine the oldest and newest
TLS version that is acceptable from clients. If not provided, a default value
is used.

The following versions are recognized:

* `TLS13` for TLS 1.3
* `TLS12` for TLS 1.2
* `TLS11` for TLS 1.1
* `TLS10` for TLS 1.0
