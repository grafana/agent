# Authenticator - OAuth2 Client Credentials

| Status                   |                       |
| ------------------------ |-----------------------|
| Stability                | [beta]                |
| Distributions            | [contrib]             |


This extension provides OAuth2 Client Credentials flow authenticator for HTTP and gRPC based exporters. The extension
fetches and refreshes the token after expiry automatically. For further details about OAuth2 Client Credentials flow (2-legged workflow)
refer https://datatracker.ietf.org/doc/html/rfc6749#section-4.4.

The authenticator type has to be set to `oauth2client`.

## Configuration

```yaml
extensions:
  oauth2client:
    client_id: someclientid
    client_secret: someclientsecret
    endpoint_params:
      audience: someaudience
    token_url: https://example.com/oauth2/default/v1/token
    scopes: ["api.metrics"]
    # tls settings for the token client
    tls:
      insecure: true
      ca_file: /var/lib/mycert.pem
      cert_file: certfile
      key_file: keyfile
    # timeout for the token client
    timeout: 2s
    
receivers:
  hostmetrics:
    scrapers:
      memory:
  otlp:
    protocols:
      grpc:

exporters:
  otlphttp/withauth:
    endpoint: http://localhost:9000
    auth:
      authenticator: oauth2client
      
  otlp/withauth:
    endpoint: 0.0.0.0:5000
    ca_file: /tmp/certs/ca.pem
    auth:
      authenticator: oauth2client

service:
  extensions: [oauth2client]
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: []
      exporters: [otlphttp/withauth, otlp/withauth]
```

Following are the configuration fields

- [**token_url**](https://datatracker.ietf.org/doc/html/rfc6749#section-3.2) - The resource server's token endpoint URLs.
- [**client_id**](https://datatracker.ietf.org/doc/html/rfc6749#section-2.2) - The client identifier issued to the client.
- [**client_secret**](https://datatracker.ietf.org/doc/html/rfc6749#section-2.3.1) - The secret string associated with above identifier.
- [**endpoint_params**](https://github.com/golang/oauth2/blob/master/clientcredentials/clientcredentials.go#L44) - Additional parameters that are sent to the token endpoint.
- [**scopes**](https://datatracker.ietf.org/doc/html/rfc6749#section-3.3) - **Optional** optional requested permissions associated for the client.
- [**timeout**](https://golang.org/src/net/http/client.go#L90) -  **Optional** specifies the timeout on the underlying client to authorization server for fetching the tokens (initial and while refreshing).
  This is optional and not setting this configuration implies there is no timeout on the client.

For more information on client side TLS settings, see [configtls README](https://github.com/open-telemetry/opentelemetry-collector/tree/main/config/configtls).

[beta]: https://github.com/open-telemetry/opentelemetry-collector#beta
[contrib]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
