---
aliases:
- /docs/agent/shared/flow/reference/components/local-file-arguments-text/
- /docs/grafana-cloud/agent/shared/flow/reference/components/local-file-arguments-text/
- /docs/grafana-cloud/monitor-infrastructure/agent/shared/flow/reference/components/local-file-arguments-text/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/shared/flow/reference/components/local-file-arguments-text/
- /docs/grafana-cloud/send-data/agent/shared/flow/reference/components/local-file-arguments-text/
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/local-file-arguments-text/
description: Shared content, local file arguments text
headless: true
---

### File change detectors

File change detectors are used to detect when the file needs to be re-read from disk. `local.file` supports two detectors: `fsnotify` and `poll`.

#### fsnotify

The `fsnotify` detector subscribes to filesystem events, which indicate when the watched file is updated.
This detector requires a filesystem that supports events at the operating system level. Network-based filesystems like NFS or FUSE won't work.

The component re-reads the watched file when a filesystem event is received.
This happens for any filesystem event to the file, including a change of permissions.

`fsnotify` also polls for changes to the file with the configured `poll_frequency` as a fallback.

`fsnotify` will stop receiving filesystem events if the watched file has been deleted, renamed, or moved.
The subscription will be re-established on the next poll once the watched file exists again.

#### poll

The `poll` file change detector will cause the watched file to be re-read every `poll_frequency`, regardless of whether the file changed.