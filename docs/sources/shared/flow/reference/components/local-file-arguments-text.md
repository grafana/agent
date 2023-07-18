---
aliases:
- /docs/agent/shared/flow/reference/components/local-file-arguments-text/
canonical: https://grafana.com/docs/grafana/agent/latest/shared/flow/reference/components/local-file-arguments-text/
headless: true
---

### File change detectors

File change detectors are used for detecting when the file needs to be re-read
from disk. `local.file` supports two detectors: `fsnotify` and `poll`.

#### fsnotify

The `fsnotify` detector subscribes to filesystem events which indicate when the
watched file had been updated. This requires a filesystem which supports events
at the Operating System level: network-based filesystems like NFS or FUSE won't
work.

When a filesystem event is received, the component will reread the watched
file. This will happen for any filesystem event to the file, including a change
of permissions.

`fsnotify` also polls for changes to the file with the configured
`poll_frequency` as a fallback.

`fsnotify` will stop receiving filesystem events if the watched file has been
deleted, renamed, or moved. The subscription will be re-established on the next
poll once the watched file exists again.

#### poll

The `poll` file change detector will cause the watched file to be reread
every `poll_frequency`, regardless of whether the file changed.