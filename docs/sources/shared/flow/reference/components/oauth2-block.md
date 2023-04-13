---
aliases:
- /docs/agent/shared/flow/reference/components/oauth2-block/
headless: true
---

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`client_id` | `string` | OAuth2 client ID. | | no
`client_secret` | `secret` | OAuth2 client secret. | | no
`client_secret_file` | `string` | File containing the OAuth2 client secret. | | no
`scopes` | `list(string)` | List of scopes to authenticate with. | | no
`token_url` | `string` | URL to fetch the token from. | | no
`endpoint_params` | `map(string)` | Optional parameters to append to the token URL. | | no
`proxy_url` | `string` | Optional proxy URL for OAuth2 requests. | | no

`client_secret` and `client_secret_file` are mutually exclusive and only one
can be provided inside of an `oauth2` block.

The `oauth2` block may also contain its own separate `tls_config` sub-block.
