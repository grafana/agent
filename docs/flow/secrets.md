# Secrets

`secret` is a primitive type in Flow. Secrets are string-like values which
contain confidential information like passwords or API keys.

Any `string` value may be assigned to a `secret` field, but not the inverse: a
`secret` value cannot be assigned to a field which expects a `string`.

When a value is a `secret`, its contents are scrubbed from the `/-/config`
endpoint, instead displaying as `(secret)`.

## Secret argument in components

Components which may load secrets (such as API keys) commonly have an argument
named `secret`.

When the `secret` argument is set to `true`, the component will export a
`secret` value instead of a `string`. This hides its value from the `/-/config`
endpoint and restricts use of the export to fields which expect secrets.
