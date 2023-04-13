---
title: module.git
labels:
  stage: beta
---

# module.git

{{< docs/shared lookup="flow/stability/beta.md" source="agent" >}}

`module.git` is a *module loader* component. A module loader is a Grafana Agent Flow
component which retrieves a [module][] and runs the components defined inside of it.

`module.git` retrieves a module source from a file in a Git repository.

[module]: {{< relref "../../concepts/modules.md" >}}

## Usage

```river
module. "LABEL" {
  repository = "GIT_REPOSTORY"
  path       = "PATH_TO_MODULE"

  arguments {
    MODULE_ARGUMENT_1 = VALUE_1
    MODULE_ARGUMENT_2 = VALUE_2
    ...
  }
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`repository` | `string` | The Git repository address to retrieve the module from. | | yes
`revision` | `string` | The Git revision to retrieve the module from. | `"HEAD"` | no
`path` | `string` | The path in the repository where the module is stored. | | yes
`pull_frequency` | `duration` | The frequency to pull the repository for updates. | | `"60s"`

The `repository` attribute must be set to a repository address that would be
recognized by Git with a `git clone REPOSITORY_ADDRESS` command, such as
`htts://github.com/grafana/agent.git`.

The `revision` attribute, when provided, must be set to a valid branch, tag, or
commit SHA within the repository.

The `path` attribute must be set to a path which is accessible from the root of
the repository, such as `FILE_NAME.river` or `FOLDER_NAME/FILE_NAME.river`.

If `pull_frequency` is not `"0s"`, the Git repository will be pulled for
updates at the frequency specified, causing the loaded module to update with
the retrieved changes.

## Blocks

The following blocks are supported inside the definition of `module.git`:

Hierarchy        | Block      | Description | Required
---------------- | ---------- | ----------- | --------
arguments | [arguments][] | Arguments to pass to the module. | no

[arguments]: #arguments-block

### arguments block

The `arguments` block specifies the list of values to pass to the loaded
module.

The attributes provided in the `arguments` block are validated based on the
[argument blocks][] defined in the module source:

* If a module source marks one of its arguments as required, it must be
  provided as an attribute in the `arguments` block of the module loader.

* Attributes in the `argument` block of the module loader will be rejected if
  they are not defined in the module source.

[argument blocks]: {{< relref "../config-blocks/argument.md" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`exports` | `map(any)` | The exports of the Module loader.

`exports` exposes the `export` config block inside a module. It can be accessed
from the parent config via `module.git.COMPONENT_LABEL.exports.EXPORT_LABEL`.

Values in `exports` correspond to [export blocks][] defined in the module
source.

[export blocks]: {{< relref "../config-blocks/export.md" >}}

## Component health

`module.git` is reported as healthy if the repository was cloned successfully
and most recent load of the module was successful.

## Debug information

`module.git` includes debug information for:

* The full SHA of the currently checked out revision.
* The most recent error when trying to fetch the repository, if any.

### Debug metrics

`module.git` does not expose any component-specific debug metrics.

## Example

This example uses a module loaded from a Git repository which adds two numbers:

```river
module.git "add" {
  repository = "https://github.com/rfratto/agent-modules.git"
  revision   = "main"
  path       = "add/module.river"

  arguments {
    a = 15
    b = 45
  }
}
```
