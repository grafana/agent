# Changelog

> _Contributors should read our [contributors guide][] for instructions on how
> to update the changelog._

This document contains a historical list of changes between releases. Only
changes that impact end-user behavior are listed; changes to documentation or
internal API changes are not present.

Unreleased
----------

### Bugfixes

- Fix `podAnnotations` values reference in pod template (should be `controller.podAnnotations`)

0.3.0 (2023-01-23)
------------------

### Security

- Change config reloader image to `jimmidyson/configmap-reload:v0.8.0` to resolve security scanner report. (@rfratto)

0.2.3 (2023-01-17)
------------------

### Bugfixes

- Sets correct arguments for starting the agent when static mode is selected.

0.2.2 (2023-01-17)
------------------

### Bugfixes

- Updated configmap template to use correct variable for populating configmap content

0.2.1 (2023-01-12)
------------------

### Other changes

- Updated documentation to remove warning about the chart not being ready for
  use.

0.2.0 (2023-01-12)
------------------

### Features

- Introduce supporting extra ports on the Grafana Agent created by Helm Chart.

0.1.0 (2023-01-11)
------------------

### Features

- Introduce a Grafana Agent Helm chart which supports Grafana Agent Flow. (@rfratto)

[contributors guide]: ../../../../docs/developer/contributing.md
