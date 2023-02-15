---
aliases:
- /docs/agent/shared/flow/reference/components/http-client-config-squashedblock/
headless: true
---
`bearer_token` | `secret` | Bearer token to authenticate with. | | no
`bearer_token_file` | `string` | File containing a bearer token to authenticate with. | | no
`proxy_url` | `string` | HTTP proxy to proxy requests through. | | no
`follow_redirects` | `bool` | Whether redirects returned by the server should be followed. | `true` | no
`enable_http2` | `bool` | Whether HTTP2 is supported for requests. | `true` | no

At most one of the following can be provided:
- [`bearer_token` argument](#Arguments).
- [`bearer_token_file` argument](#Arguments). 
- [`basic_auth` block][basic_auth].
- [`authorization` block][authorization].
- [`oauth2` block][oauth2].