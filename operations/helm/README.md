# agent-helm-chart

This is an **experimental** repository for creating a Helm chart for Grafana
Agent (in particular, [Grafana Agent Flow][Flow]).

It is not recommended to use this Helm chart in production yet. The intent is
to turn upstream this into grafana/agent once it's ready.

[Flow]: https://grafana.com/docs/agent/latest/flow/

## Testing

The `tests` contains a list of golden templates rendered from the Helm chart.

These manifests are never run directly, but are instead used to validate the
correctness of the templates emitted by the Helm chart. To regenerate this
folder, call `make rebuild-tests`.

`make rebuild-tests` will iterate through the value.yaml files in
`charts/grafana-agent/tests` and generate each one as a separate directory.

When modifying the Helm charts, `make rebuild-tests` must be run before
submitting a PR, as a linter check will ensure that this directory is up to
date.
