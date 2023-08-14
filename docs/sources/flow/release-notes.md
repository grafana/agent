---
canonical: https://grafana.com/docs/agent/latest/flow/upgrade-guide/
description: Release notes for Grafana Agent flow mode
title: Release notes for Grafana Agentflow mode
menuTitle: Release notes
aliases:
- ./upgrade-guide/
weight: 999
---

# Release notes

The release notes provide information about deprecations and breaking changes in Grafana Agent flow mode.

For a complete list of changes to Grafana Agent, with links to pull requests and related issues when available, refer to the [Changelog](https://github.com/grafana/agent/blob/main/CHANGELOG.md).

{{% admonition type="note" %}}
These release notes are specific to Grafana Agent flow mode.
Other release notes for the different Grafana Agent variants are contained on separate pages:

* [Static mode release notes][release-notes-static]
* [Static mode Kubernetes operator release notes][release-notes-operator]

[release-notes-static]: {{< relref "../static/release-notes.md" >}}
[release-notes-operator]: {{< relref "../operator/release-notes.md" >}}
{{% /admonition %}}

## v0.33

### Symbolic links in Docker containers removed

We've removed the deprecated symbolic links to `/bin/agent*` in Docker
containers, as planned in v0.31. In case you're setting a custom entrypoint,
use the new binaries that are prefixed with `/bin/grafana*`.

## v0.32

### Breaking change: `http_client_config` Flow blocks merged with parent blocks

To reduce the amount of typing required to write Flow components, the arguments
and subblocks found in `http_client_config` have been merged with their parent
blocks:

- `discovery.docker > http_client_config` is merged into the `discovery.docker` block.
- `discovery.kubernetes > http_client_config` is merged into the `discovery.kubernetes` block.
- `loki.source.kubernetes > client > http_client_config` is merged into the `client` block.
- `loki.source.podlogs > client > http_client_config` is merged into the `client` block.
- `loki.write > endpoint > http_client_config` is merged into the `endpoint` block.
- `mimir.rules.kubernetes > http_client_config` is merged into the `mimir.rules.kubernetes` block.
- `otelcol.receiver.opencensus > grpc` is merged into the `otelcol.receiver.opencensus` block.
- `otelcol.receiver.zipkin > http` is merged into the `otelcol.receiver.zipkin` block.
- `phlare.scrape > http_client_config` is merged into the `phlare.scrape` block.
- `phlare.write > endpoint > http_client_config` is merged into the `endpoint` block.
- `prometheus.remote_write > endpoint > http_client_config` is merged into the `endpoint` block.
- `prometheus.scrape > http_client_config` is merged into the `prometheus.scrape` block.

Old configuration example:

```river
prometheus.remote_write "example" {
  endpoint {
    url = URL

    http_client_config {
      basic_auth {
        username = BASIC_AUTH_USERNAME
        password = BASIC_AUTH_PASSWORD
      }
    }
  }
}
```

New configuration example:

```river
prometheus.remote_write "example" {
  endpoint {
    url = URL

    basic_auth {
      username = BASIC_AUTH_USERNAME
      password = BASIC_AUTH_PASSWORD
    }
  }
}
```

### Breaking change: `loki.process` stage blocks combined into new blocks

Previously, to add a stage to `loki.process`, two blocks were needed: a block
called `stage`, then an inner block for the stage being written. Stage blocks
are now a single block called `stage.STAGENAME`.

Old configuration example:

```river
loki.process "example" {
  forward_to = RECEIVER_LIST

  stage {
    docker {}
  }

  stage {
    json {
      expressions = { output = "log", extra = "" }
    }
  }
}
```

New configuration example:

```river
loki.process "example" {
  forward_to = RECEIVER_LIST

  stage.docker {}

  stage.json {
    expressions = { output = "log", extra = "" }
  }
}
```

### Breaking change: `client_options` block renamed in `remote.s3` component

To synchronize naming conventions between `remote.s3` and `remote.http`, the
`client_options` block has been renamed `client`.

Old configuration example:

```river
remote.s3 "example" {
  path = S3_PATH

  client_options {
    key    = ACCESS_KEY
    secret = KEY_SECRET
  }
}
```

New configuration example:

```river
remote.s3 "example" {
  path = S3_PATH

  client {
    key    = ACCESS_KEY
    secret = KEY_SECRET
  }
}
```

### Breaking change: `prometheus.integration.node_exporter` component name changed

The `prometheus.integration.node_exporter` component has been renamed to
`prometheus.exporter.unix`. `unix` was chosen as a name to approximate the
\*nix-like systems the exporter supports.

Old configuration example:

```river
prometheus.integration.node_exporter { }
```

New configuration example:

```river
prometheus.exporter.unix { }
```

### Breaking change: support for `EXPERIMENTAL_ENABLE_FLOW` environment variable removed

As first announced in v0.30.0, support for using the `EXPERIMENTAL_ENABLE_FLOW`
environment variable to enable Flow mode has been removed.

To enable Flow mode, set the `AGENT_MODE` environment variable to `flow`.

## v0.31

### Breaking change: binary names are now prefixed with `grafana-`

As first announced in v0.29, the `agent` release binary name is now prefixed
with `grafana-`:

- `agent` is now `grafana-agent`.

For the `grafana/agent` Docker container, the entrypoint is now
`/bin/grafana-agent`. A symbolic link from `/bin/agent` to the new binary has
been added.

Symbolic links will be removed in v0.33. Custom entrypoints must be
updated prior to v0.33 to use the new binaries before the symbolic links get
removed.

## v0.30

### Deprecation: `EXPERIMENTAL_ENABLE_FLOW` environment variable changed

As part of graduating Grafana Agent Flow to beta, the
`EXPERIMENTAL_ENABLE_FLOW` environment variable is replaced by setting
`AGENT_MODE` to `flow`.

Setting `EXPERIMENTAL_ENABLE_FLOW` to `1` or `true` is now deprecated and
support for it will be removed for the v0.32 release.

## v0.29

### Deprecation: binary names will be prefixed with `grafana-` in v0.31.0

The binary name `agent` has been deprecated and will be renamed to
`grafana-agent` in the v0.31.0 release.

As part of this change, the Docker containers for the v0.31.0 release will
include symbolic links from the old binary names to the new binary names.

There is no action to take at this time.
