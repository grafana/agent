# remote.s3

The `remote.s3` component exposes the contents of a file located in an S3 compatible system to other components. The file will be polled for changes so that the most recent content is always available.

Multiple `remote.s3` components can be specified by giving them different name labels. By default the component will use AWS environment vars or profiles that are setup. If those are not setup or overrides are required then `key` and `secret` can be used.

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
`path` | `string` | Path in the format of `"s3://bucket/file"`                              | | **yes**
`poll_frequency` | `duration` | How often to poll the file for changes, must be greater than 30 seconds | `"10m"` | no
`is_secret` | `bool` | Marks the file as containing a [secret][]                               | `false` | no
`client_options` | `client_options` | Additional [options](#client_options) for configuring the client      | | no

### client_options

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`key` | `string` | Used to override default access key | | no
`secret` | `secret` | Used to override default secret | | no
`endpoint` | `string` | Endpoint specifies a custom url to access, used generally for S3 compatible systems | | no
`disable_ssl` | `bool` | Used to disable SSL, generally used for testing | | no
`use_path_style` | `string` | Path style is a deprecated that is generally enabled for S3 compatible systems | `false` | no
`region` | `string` | Used to override default region | | no

## Exported fields

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`content` | `string` or `secret` | The contents of the file | | no

The `content` field will be secret is `is_secret` was set.

## Component health

Any `remote.s3` component will report as healthy if the watched file was read successfully.

Failed to read the file will cause the component to be unhealthy but the last good read will kept. The error will be exposed on the health information

## Debug information

`remote.s3` does not expose any component-specific debug information

[secret]: ../secrets.md#is_secret-argument-in-components
