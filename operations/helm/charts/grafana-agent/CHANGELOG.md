# Changelog

> _Contributors should read our [contributors guide][] for instructions on how
> to update the changelog._

This document contains a historical list of changes between releases. Only
changes that impact end-user behavior are listed; changes to documentation or
internal API changes are not present.

Unreleased
----------

### Enhancements

- Add support for adding Annotations to Service (@ofirshtrull)

0.10.0 (2023-03-09)
-------------------

### Enhancements

- Add Horizontal Pod Autoscaling for controller type deployment. (@therealmanny)
- Add affinity values. (@therealmanny)

0.9.0 (2023-03-14)
------------------

### Enhancements

- Add PodMonitors, ServiceMonitors, and Probes to the agent ClusterRole. (@captncraig)
- Add podLabels values. (@therealmanny)


0.8.1 (2023-03-06)
------------------

### Enhancements

- Add hostPort specification to extraPorts and extraPort documentation. (@pnathan)
- Selectively template ClusterIP. (@aglees)
- Add priorityClassName value. (@aglees)
- Update Grafana Agent version to v0.32.1. (@erikbaranowski)

0.8.0 (2023-02-28)
------------------

### Enhancements

- Update Grafana Agent version to v0.32.0. (@rfratto)

0.7.1 (2023-02-27)
------------------

### Bugfixes

- Fix issue where `.image.pullPolicy` was not being respected. (@rfratto)

0.7.0 (2023-02-24)
------------------

### Enhancements

- Helm chart: Add support for templates inside of configMap.content (@ts-mini)
- Add the necessary rbac to support eventhandler integration (@nvanheuverzwijn)


0.6.0 (2023-02-13)
------------------

### Enhancements

- Update Grafana Agent version to v0.31.3. (@rfratto)

0.5.0 (2023-02-08)
------------------

### Enhancements

- Helm Chart: Add ingress and support for agent-receiver. (@ts-mini)

### Documentation

- Update Helm Chart documentation to reference new `loki.source.kubernetes` component.

0.4.0 (2023-01-31)
------------------

### Enhancements

- Update Grafana Agent version to v0.31.0. (@rfratto)
- Install PodLogs CRD for the `loki.source.podlogs` Flow component. (@rfratto)
- Update RBAC rules to permit `loki.source.podlogs` and `mimir.rules.kubernetes` to work by default. (@rfratto)

0.3.1 (2023-01-31)
------------------

### Bugfixes

- Fix `podAnnotations` values reference in pod template (should be `controller.podAnnotations`).
- Ensure the service gets a clusterIP assigned by default.

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
