---
aliases:
- /docs/agent/shared/flow/reference/components/otelcol-compression-field
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/otelcol-compression-field/
headless: true
---

By default, requests are compressed with gzip. The `compression` argument
controls which compression mechanism to use. Supported strings are:

* `"gzip"`
* `"zlib"`
* `"deflate"`
* `"snappy"`
* `"zstd"`

If `compression` is set to `"none"` or an empty string `""`, no compression is
used.
