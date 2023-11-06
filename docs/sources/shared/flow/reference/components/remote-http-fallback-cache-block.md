---
aliases:
- /docs/agent/shared/flow/reference/components/remote-http-fallback-cache-block/
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/remote-http-fallback-cache-block/
description: Shared content, remote http fallback cache
headless: true
---

The `fallback_cache` block configures a local cache that is used as a fallback if
the remote endpoint is unavailable or returns an error.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `bool` | Whether to enable the fallback cache. | `false` | no
`max_age` | `duration` | Maximum age of the cached response. | `"0"` | no
`allow_secrets` | `bool` | Whether to allow secrets in the cached response. | `false` | no

A `max_age` of `0` means that the cache will never be considered stale.

If the cache was last updated more than `max_age` ago, the cache is considered
stale and the component will not fall back to it.

If `is_secret` is `true` and `allow_secrets` is `false`, the component will
not write the cached response, even if the cache.
