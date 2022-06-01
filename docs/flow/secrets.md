# Secrets

`secret` is a primitive type in Flow. Secrets are string-like values which
contain sensitive information.

Any `string` value may be assigned to a `secret` field, but not the inverse: a
`secret` value cannot be assigned to a field which expects a `string`.

When a value is a `secret`, its contents are scrubbed from the `/-/config`
endpoint, instead displaying as `(secret)`.

## Sensitive argument in components

Components which may load sensitive information (such as API keys) commonly
have an argument named `sensitive`.

When the `sensitive` argument is set to `true`, the component will export a
`secret` value instead of a `string`. This hides its value from the `/-/config`
endpoint and restricts use of the export to fields which expect sensitive
information.
