# Changelog

> _Contributors should read our [contributors guide][] for instructions on how
> to update the changelog._

This document contains a historical list of changes between releases. Only
changes that impact end-user behavior are listed; changes to documentation or
internal API changes are not present.

Unreleased
----------

0.26.0 (2023-10-10)
-------------------

### Breaking changes

- The `initContainers` setting has been moved to `controller.initContainers`
  for consistency with other Pod-level settings. (@rfratto)

### Enhancements

- Make CRDs optional through the `crds.create` setting. (@bentonam, @rfratto)

- Update Grafana Agent version to v0.37.1. (@tpaschalis)

0.25.0 (2023-09-22)
-------------------

### Enhancements

- An image's digest can now be used in place of a tag. (@hainenber)

- Add ServiceMonitor support. (@QuentinBisson)

- Update Grafana Agent version to v0.36.2. (@ptodev)

0.24.0 (2023-09-08)
-------------------

### Enhancements

- StatefulSets will now use `podManagementPolicy: Parallel` by default. To
  disable this behavior, set `controller.parallelRollout` to `false`.
  (@rfratto)

0.23.0 (2023-09-06)
-------------------

### Enhancements

- Update Grafana Agent version to v0.36.1. (@erikbaranowski)

- Enable clustering for deployments and daemonsets. (@tpaschalis)

0.22.0 (2023-08-30)
-------------------

- Update Grafana Agent version to v0.36.0. (@thampiotr)

0.21.1 (2023-08-30)
-------------------

- Condition parameter minReadySeconds on StatefulSet, Deployment, and DaemonSet to Kubernetes v1.22 clusters.

0.21.0 (2023-08-15)
-------------------

- Update Grafana Agent version to v0.35.4. (@mattdurham)

0.20.0 (2023-08-09)
-------------------

- Update Grafana Agent version to v0.35.3. (@tpaschalis)

### Enhancements

- Add support for initcontainers in helm chart. (@dwalker-sabiogroup)

0.19.0 (2023-07-27)
-------------------

### Enhancements

- Set hostPID from values. (@korniltsev)

- Set nodeSelector at podlevel. (@Flasheh)

- Update Grafana Agent version to v0.35.2. (@rfratto)

0.18.0 (2023-07-26)
-------------------

### Enhancements

- Update Grafana Agent version to v0.35.1. (@ptodev)

0.17.0 (2023-07-19)
-------------------

### Features

- Add native support for Flow mode clustering with the
  `agent.clustering.enabled` flag. Clustering may only be enabled in Flow mode
  when deploying a StatefulSet. (@rfratto)

### Enhancements

- Set securityContext for configReloader container. (@yanehi)

- Set securityContext at podlevel. (@yanehi)

- Update Grafana Agent version to v0.35.0. (@mattdurham)

0.16.0 (2023-06-20)
-------------------

### Enhancements

- Allow requests to be set on the config reloader container. (@tpaschalis)

- Allow users of the helm chart to configure the image registry either at the image level or globally. (@QuentinBisson)

- Don't specify replica count for StatefulSets when autoscaling is enabled (@captncraig)

- Update Grafana Agent version to v0.34.2. (@captncraig)

### Other changes

- Make the agent and config-reloader container resources required when using
  autoscaling. (@tpaschalis)

0.15.0 (2023-06-08)
-------------------

### Enhancements

- Update Grafana Agent version to v0.34.0. (@captncraig)

- Add HPA support for Deployments and StatefulSets. (@tpaschalis)

- Make the Faro port optional. (@tpaschalis)

- Rename the deprecated `serviceAccount` alias to `serviceAccountName` in
  pod template. This is a no-op change. (@tpaschalis)

### Bugfixes

- Only set the deployment replicas when autoscaling is disabled. (@tiithansen)

- Reorder HPA `spec.metrics` to avoid endless sync loop in ArgoCD. (@tiithansen)

0.14.0 (2023-05-11)
-------------------

### Enhancements

- Add a toggle for enabling/disabling the Service. (@tpaschalis)

- Update Grafana Agent version to v0.33.2. (@rfratto)

0.13.0 (2023-05-01)
-------------------

### Enhancements

- Update Grafana Agent version to v0.33.1. (@spartan0x117)

- Update RBAC rules to permit `node/metrics`. (@yurii-kryvosheia)

0.12.0 (2023-04-25)
-------------------

### Enhancements

- Update Grafana Agent version to v0.33.0. (@rfratto)

0.11.0 (2023-04-24)
-------------------

### Enhancements

- Add support for adding Annotations to Service (@ofirshtrull)
- Add `agent.envFrom` value. (@carlosjgp)
- Add `controller.hostNetwork` value. (@carlosjgp)
- Add `controller.dnsPolicy` value. (@carlosjgp)

### Bugfixes

- Fix issue where `controller.tolerations` setting was ignored. (@carlosjgp)
- Fix YAML indentation of some resources. (@carlosjgp)

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
