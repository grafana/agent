---
aliases:
- ../../configuration-language/standard-library/discovery_target_decode/
title: discovery_target_decode
---

# discovery_target_decode

The `discovery_target_decode` function decodes a string into an array of
targets matching the exports of `discovery.*` components.

The string must match the JSON format used by Prometheus' HTTP and file service
discovery:

```
[
  {
    "targets": [ "<host:ip>", ... ],
    "labels": {
      "<label name>": "<label value>"
    }
  }
]
```

If the provided string doesn't match the expected JSON format,
`discovery_target_decode` fails to evaluate and marks the component containing
the expression as unhealthy.

Elements specified by the `targets` key are converted into a flat list of
targets. The base set of labels for each target is retrieved from the `labels`
key, and the `__address__` label is received from the target element.

For example, the following JSON file maps to the River objects provided below:

```json
[
  {
    "targets": [ "host-a:80", "host-b:80" ],
    "labels": {
      "cluster": "production",
      "region": "us-west-0"
    }
  },
  {
    "targets": [ "host-c:80" ],
    "labels": {
      "cluster": "development",
      "region": "us-west-0"
    }
  }
]
```

```river
[
  {
    __address__ = "host-a:80",
    cluster     = "production",
    region      = "us-west-0",
  },
  {
    __address__ = "host-b:80",
    cluster     = "production",
    region      = "us-west-0",
  },
  {
    __address__ = "host-c:80",
    cluster     = "development",
    region      = "us-west-0",
  },
]
```

## Example pipeline

```river
local.file "example" {
  filename = env("TARGETS_FILE")
}

prometheus.scrape "default" {
  targets = discovery_target_decode(local.file.example.content)
}
```

[`local.file`]: {{< relref "../components/local.file.md" >}}
