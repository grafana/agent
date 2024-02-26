---
aliases:
- /docs/agent/shared/flow/reference/components/remote-http-fallback-cache-block/
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/remote-http-fallback-cache-block/
description: Shared content, remote http fallback cache
headless: true
---

The `fallback_cache` block configures a local cache that is used as a fallback if
the remote endpoint is unavailable or returns an error.

Cache files are created with `600` permissions and owned by the user running the
agent. If the contents of the response are sensitive, it is not recommended to
enable this feature.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `bool` | Whether to enable the fallback cache. | `false` | no
