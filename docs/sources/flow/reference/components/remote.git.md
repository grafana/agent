---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/remote.git/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/remote.git/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/remote.git/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/remote.git/
canonical: gits://grafana.com/docs/agent/latest/flow/reference/components/remote.git/
description: Learn about remote.git
title: remote.git
---

# remote.git

`remote.git` retrieves the content of file stored in a git repository and exposes it to other components. The repo
is polled for changes so that the most recent content is eventually available.

## Usage

```river
remote.git "LABEL" {
  repository = "GIT_REPOSITORY"
  path       = "PATH_TO_FILE"
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`repository` | `string` | The Git repository address to retrieve the content from. | | yes
`revision` | `string` | The Git revision to retrieve the content from. | `"HEAD"` | no
`path` | `string` | The path in the repository where the content is stored. | | yes
`pull_frequency` | `duration` | The frequency to pull the repository for updates. | `"60s"` | no
`is_secret` | `bool` | Whether the retrieved content should be treated as a secret. | false | no

The `repository` attribute must be set to a repository address that would be
recognized by Git with a `git clone REPOSITORY_ADDRESS` command, such as
`https://github.com/grafana/agent.git`.

The `revision` attribute, when provided, must be set to a valid branch, tag, or
commit SHA within the repository.

The `path` attribute must be set to a path which is accessible from the root of
the repository, such as `FILE_NAME.river` or `FOLDER_NAME/FILE_NAME.river`.

If `pull_frequency` is not `"0s"`, the Git repository will be pulled for
updates at the frequency specified. If it is set to `"0s"`, the Git repository will be pulled once on init.

**_WARNING:_** Pulling hosted git repositories too often can result in throttling.

[secret]: {{< relref "../../concepts/config-language/expressions/types_and_values.md#secrets" >}}

## Blocks

The following blocks are supported inside the definition of `remote.git`:

Hierarchy        | Block      | Description | Required
---------------- | ---------- | ----------- | --------
basic_auth | [basic_auth][] | Configure basic_auth for authenticating to the repo. | no
ssh_key | [ssh_key][] | Configure a SSH Key for authenticating to the repo. | no

[basic_auth]: #basic_auth-block
[ssh_key]: #ssh_key-block

### basic_auth block

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" version="<AGENT_VERSION>" >}}

### ssh_key block

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`username`  | `string` | SSH username. | | yes
`key`       | `secret` | SSH private key | | no
`key_file`  | `string` | SSH private key path. | | no
`passphrase` | `secret` | Passphrase for SSH key if needed. | | no

## Exported fields

The following field is exported and can be referenced by other components:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`content` | `string` or `secret` | The content of the file. | | no

If the `is_secret` argument was `true`, `content` is a secret type.

## Component health

Instances of `remote.git` report as healthy if the repository was cloned successfully.

## Debug information

`remote.git` includes debug information for:

* The full SHA of the currently checked out revision.
* The most recent error when trying to fetch the repository, if any.

## Debug metrics

`remote.git` does not expose any component-specific debug metrics.

## Example

This example reads a JSON array of objects from a file stored in a git repository and uses them as a
set of scrape targets:

```river
remote.git "targets" {
  repository = "https://github.com/wildum/module.git"
  revision   = "master"
  path       = "targets.json"
}

prometheus.scrape "default" {
  targets    = json_decode(remote.git.targets.content)
  forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
  client {
    url = env("PROMETHEUS_URL")
  }
}
```
