---
aliases:
- ./upgrade-guide/
- /docs/grafana-cloud/agent/operator/release-notes/
- /docs/grafana-cloud/monitor-infrastructure/agent/operator/release-notes/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/operator/release-notes/
- /docs/grafana-cloud/send-data/agent/operator/release-notes/
canonical: https://grafana.com/docs/agent/latest/operator/release-notes/
description: Release notes for Grafana Agent Operator
menuTitle: Release notes
title: Release notes for Grafana Agent Operator
weight: 999
---

# Release notes for Grafana Agent Operator

The release notes provide information about deprecations and breaking changes in Grafana Agent static mode Kubernetes operator.

For a complete list of changes to Grafana Agent, with links to pull requests and related issues when available, refer to the [Changelog](https://github.com/grafana/agent/blob/main/CHANGELOG.md).

> **Note:** These release notes are specific to the Static mode Kubernetes Operator.
> Other release notes for the different Grafana Agent variants are contained on separate pages:
>
> - [Static mode release notes][release-notes-static]
> - [Flow mode release notes][release-notes-flow]

{{% docs/reference %}}
[release-notes-static]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/release-notes"
[release-notes-static]: "/docs/agent/ -> /docs/grafana-cloud/send-data/agent/static/release-notes"

[release-notes-flow]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/release-notes"
[release-notes-flow]: "/docs/grafana-cloud/ -> /docs/agent/<AGENT_VERSION>/flow/release-notes"
{{% /docs/reference %}}

## v0.33

### Symbolic links in Docker containers removed

We've removed the deprecated symbolic links to `/bin/agent*` in Docker
containers, as planned in v0.31. In case you're setting a custom entrypoint,
use the new binaries that are prefixed with `/bin/grafana*`.

## v0.31

### Breaking change: binary names are now prefixed with `grafana-`

As first announced in v0.29, the `grafana-operator` release binary names is now
prefixed with `grafana-`:

- `agent-operator` is now `grafana-agent-operator`.

For the `grafana/agent-operator` Docker container, the entrypoint is now
`/bin/grafana-agent-operator`. A symbolic link from `/bin/agent-operator` to
the new binary has been added.

Symbolic links will be removed in v0.33. Custom entrypoints must be
updated prior to v0.33 to use the new binaries before the symbolic links get
removed.

## v0.29

### Deprecation: binary names will be prefixed with `grafana-` in v0.31.0

The `agent-operator` binary name has been deprecated and will be renamed to
`grafana-agent-operator` in the v0.31.0 release.

As part of this change, the Docker container for the v0.31.0 release will
include symbolic links from the old binary names to the new binary names.

There is no action to take at this time.

## v0.24

### Breaking change: Grafana Agent Operator supported Agent versions

The v0.24.0 release of Grafana Agent Operator can no longer deploy versions of
Grafana Agent prior to v0.24.0.

## v0.19

### Rename of Prometheus to Metrics (Breaking change)

As a part of the deprecation of "Prometheus," all Operator CRDs and fields with
"Prometheus" in the name have changed to "Metrics."

This includes:

- The `PrometheusInstance` CRD is now `MetricsInstance` (referenced by
  `metricsinstances` and not `metrics-instances` within ClusterRoles).
- The `Prometheus` field of the `GrafanaAgent` resource is now `Metrics`
- `PrometheusExternalLabelName` is now `MetricsExternalLabelName`

This is a hard breaking change, and all fields must change accordingly for the
operator to continue working.

Note that old CRDs with the old hyphenated names must be deleted (`kubectl
delete crds/{grafana-agents,prometheus-instances}`) for ClusterRoles to work
correctly.

To do a zero-downtime upgrade of the Operator when there is a breaking change,
refer to the new `agentctl operator-detatch` command: this will iterate through
all of your objects and remove any OwnerReferences to a CRD, allowing you to
delete your Operator CRDs or CRs.

### Rename of CRD paths (Breaking change)

`prometheus-instances` and `grafana-agents` have been renamed to
`metricsinstances` and `grafanaagents` respectively. This is to remain
consistent with how Kubernetes names multi-word objects.

As a result, you will need to update your ClusterRoles to change the path of
resources.

To do a zero-downtime upgrade of the Operator when there is a breaking change,
refer to the new `agentctl operator-detatch` command: this will iterate through
all of your objects and remove any OwnerReferences to a CRD, allowing you to
delete your Operator CRDs or CRs.


Example old ClusterRole:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: grafana-agent-operator
rules:
- apiGroups: [monitoring.grafana.com]
  resources:
  - grafana-agents
  - prometheus-instances
  verbs: [get, list, watch]
```

Example new ClusterRole:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: grafana-agent-operator
rules:
- apiGroups: [monitoring.grafana.com]
  resources:
  - grafanaagents
  - metricsinstances
  verbs: [get, list, watch]
```
