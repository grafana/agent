# Developer's Guide

This document contains maintainer-specific information.

Table of Contents:

1. [Introduction](#introduction)
2. [Updating CRDs](#updating-crds)
3. [Testing Locally](#testing-locally)
4. [Development Architecture](#development-architecture)

## Introduction

Kubernetes Operators are designed to automate the behavior of human operators
for pieces of software. The Grafana Agent Operator, in particular, is based off
of the very popular [Prometheus
Operator](https://github.com/prometheus-operator/prometheus-operator):

1. We use the same v1 CRDs from the official project.
2. We aim to generate the same remote_write and scrape_configs that the
   Prometheus Operator does.

That being said, we're not fully compatible, and the Grafana Agent Operator has
the same trade-offs that the Grafana Agent does: no recording rules, no alerts,
no local storage for querying metrics.

The public [Grafana Agent Operator design
doc](https://docs.google.com/document/d/1nlwhJLspTkkm8vLgrExJgf02b9GCAWv_Ci_a9DliI_s)
goes into more detail about the context and design decisions being made.

## Updating CRDs

The `make generate-crds` command at the root of this repository will generate CRDs and
other code used by the operator. This calls the [generate-crds
script](../../tools/generate-crds.bash) in a container. If you wish to call this
script manually, you must also install `controller-gen` and `gen-crd-api-reference-docs`.
Ensure to keep the version in sync with what's defined in the `Dockerfile`.

```
go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.9.2
go install github.com/ahmetb/gen-crd-api-reference-docs@v0.3.1-0.20220618162802-424739b250f5
```

Use the following to run the script in a container:

```
USE_CONTAINER=1 make generate-crds
```
## Testing Locally

Create a k3d cluster (depending on k3d v4.x):

```
k3d cluster create agent-operator \
  --port 30080:80@loadbalancer \
  --api-port 50043 \
  --kubeconfig-update-default=true \
  --kubeconfig-switch-context=true \
  --wait
```

### Deploy Prometheus

An example Prometheus server is provided in `./example-prometheus.yaml`. Deploy
it with the following, from the root of the repository:

```
kubectl apply -f ./cmd/grafana-agent-operator/example-prometheus.yaml
```

You can view it at http://prometheus.k3d.localhost:30080 once the k3d cluster is
running.

### Apply the CRDs

Generated CRDs used by the operator can be found in [the Production
folder](../../operations/agent-static-operator/crds). Deploy them from the root of the
repository with:

```
kubectl apply -f production/operator/crds
```

### Run the Operator

Now that the CRDs are applied, you can run the operator from the root of the
repository:

```
go run ./cmd/grafana-agent-operator
```

### Apply a GrafanaAgent custom resource

Finally, you can apply an example GrafanaAgent custom resource. One is [provided
for you](../../cmd/grafana-agent-operator/agent-example-config.yaml). From the root of the repository, run:

```
kubectl apply -f ./cmd/grafana-agent-operator/agent-example-config.yaml
```

If you are running the operator, you should see it pick up the change and start
mutating the cluster.

## Development Architecture

This project makes heavy use of the [Kubernetes SIG Controller
Runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime) project. That
project has its own documentation, but for a high level overview of how it
relates to this project:

1. The Grafana Agent Operator is composed of a single _controller_. A
   _controller_ is responsible for responding to changes to Kubernetes resources.

2. Controllers can be notified about changes to:

   1. One Primary resource (i.e., the GrafanaAgent CR)

   2. Any number of secondary resources used to deploy the managed software
      (e.g., ServiceMonitor, PodMonitors). This is done using a custom event
      handler, which we'll detail below.

   3. Any number of resources the Operator deploys (ConfigMaps, Secrets,
      StatefulSets). This is done using
      [ownerReferences](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#owners-and-dependents).

3. Controllers have one _reconciler_. The reconciler handles updating managed
   resources for one specific primary resource. The `GrafanaAgent` CRD is
   the primary resource, and the reconciler will handle updating managed
   resources for all discovered GrafanaAgent CRs. Each reconcile request is for
   a specific CR, such as `agent-1` or `agent-2`.

4. A _manager_ initializes all controllers for a project. It provides a caching
   Kubernetes client and propagates Kubernetes events to controllers.

An `EnqueueRequestForSelector` event handler was added to handle dealing to
changes to secondary resources, which is not a concept in the official
Controller Runtime project. This works by allowing the reconciler to request
events for a given primary resource if one of the secondary resource changes.
This means that multiple primary resources can watch a ServiceMonitor and cause
a reconcile when it changes.

Event handlers are specific to a resource, so there is one
`EnqueueRequestForSelector` handler per secondary resource.

Reconciles are supposed to be idempotent, so deletes, updates, and creates
should be treated the same. All managed resources are deployed with
ownerReferences set, so managed resources will be automatically deleted by
Kubernetes' garbage collector when the primary resource gets deleted by the
user.

### Flow

This section walks through what happens when a user deploys a new GrafanaAgent
CR:

1. A GrafanaAgent CR `default/agent` gets deployed to a cluster

2. The Controller's event handlers get notified about the event and queue a
   reconcile request for `default/agent`.

3. The reonciler discovers all secondary `MetricsInstance` referenced by
   `default/agent`.

4. The reconciler discovers all secondary `ServiceMonitor`, `PodMonitor` and
   `Probe` resources that are referenced by the discovered `MetricsInstance`
   resource.

5. The reconciler informs the appropriate `EnqueueRequestForSelector` event
   handlers that changes to those resources should cause a new reconcile for
   `default/agent`.

6. The reconciler discovers all `Secrets` referenced across all current
   resources. The content of the secrets are held in-memory to statically
   configure Grafana Agent fields that do not support reading in from a file
   (e.g., basic auth username).

7. All the discovered secrets are copied to a new Secret in the `default`
   namespace. This is done in case a `ServiceMonitor` is found in a different
   namespace than where the Agent will be deployed.

8. A new Secret is created for the configuration of the Grafana Agent.

9. A StatefulSet is generated for the Grafana Agent.

When `default/agent` gets deleted, all `EnqueueRequestForSelector` event
handlers get notified to stop sending events for `default/agent`.


