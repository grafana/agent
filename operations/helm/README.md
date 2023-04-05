# Helm charts

This directory contains Helm charts for Grafana Agent.

## Testing

The `tests` contains a list of golden templates rendered from the Helm chart.

These manifests are never run directly, but are instead used to validate the
correctness of the templates emitted by the Helm chart. To regenerate this
folder, call `make generate-helm-tests` from the root of the repository.

`make generate-helm-tests` will iterate through the value.yaml files in
`charts/grafana-agent/tests` and generate each one as a separate directory.

When modifying the Helm charts, `make generate-helm-tests` must be run before
submitting a PR, as a linter check will ensure that this directory is
up-to-date.
