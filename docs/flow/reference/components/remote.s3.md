---
aliases:
- /docs/agent/latest/flow/reference/components/remote.s3
title: remote.s3
---

# remote.s3

`remote.s3` exposes the contents of a file located in an S3 compatible system
to other components. The file will be polled for changes so that the most
recent content is always available.

Multiple `remote.s3` components can be specified by giving them different name
labels. By default, AWS environment variables are used to authenticate against
S3. The `key` and `secret` arguments can be used to provide custom
authentication.

## Example

```river
remote.s3 "data" {
  path = "s3://test-bucket/file.txt"
}
```

## Arguments

The following arguments are supported:

Name | Type | Description                                                             | Default | Required
---- | ---- |-------------------------------------------------------------------------| ------- | --------
`path` | `string` | Path in the format of `"s3://bucket/file"` | | **yes**
`poll_frequency` | `duration` | How often to poll the file for changes, must be greater than 30 seconds | `"10m"` | no
`is_secret` | `bool` | Marks the file as containing a [secret][] | `false` | no

The following subblocks are supported:

Name | Description | Required
---- | ----------- | --------
[`client_options`](#client_options) | Additional options for configuring the S3 client | no

### `client_options` block

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`key` | `string` | Used to override default access key | | no
`secret` | `secret` | Used to override default secret | | no
`endpoint` | `string` | Endpoint specifies a custom url to access, used generally for S3 compatible systems | | no
`disable_ssl` | `bool` | Used to disable SSL, generally used for testing | | no
`use_path_style` | `string` | Path style is a deprecated that is generally enabled for S3 compatible systems | `false` | no
`region` | `string` | Used to override default region | | no

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`content` | `string` or `secret` | The contents of the file | | no

The `content` field will be secret is `is_secret` was set.

## Component health

Instances of `remote.s3` will reported as healthy if the most recent read of
the watched file succeeded.

## Debug information

`remote.s3` does not expose any component-specific debug information.

### Debug metrics

`remote.s3` does not expose any component-specific debug metrics.

[secret]: ../secrets.md#is_secret-argument-in-components
