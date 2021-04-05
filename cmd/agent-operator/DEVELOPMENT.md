# Developing the Agent Operator

Create a k3d cluster (depending on k3d v4.x):

```
k3d cluster create agent-operator \
  --port 30080:80@loadbalancer \
  --api-port 50043 \
  --kubeconfig-update-default=true \
  --kubeconfig-switch-context=true \
  --wait
```

Now run the operator:

```
go run ./cmd/agent-operator
```

## Apply the CRDs

Generated CRDs used by the operator can be found in [the Production
folder](../../production/operator/crds). Deploy them from the root of the
repository with:

```
kubectl apply -f production/operator/crds
```

## Apply a GrafanaAgent custom resource

Finally, you can apply an example GrafanaAgent custom resource. One is [provided
for you](./agent-example-config.yaml). From the root of the repository, run:

```
kubectl apply -f ./cmd/agent-operator/agent-example-config.yaml
```

If you are running the operator, you should see it pick up the change and start
mutating the cluster.

# Temporary Steps

The Grafana Agent Operator is a WIP and some extra steps must be performed
manually until code is cleaned up.

### Intalling extra dependencies:

```
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
```

### Generating definitions

Run these from the root of the repository.
Note that CRDs from Prometheus Operator are used since we support (some) of the
same CRDs from that project.

```
pushd ./pkg/operator/api/metrics/v1alpha1
controller-gen object paths=.
controller-gen crd:crdVersions=v1 paths=. output:crd:dir=../../../../../production/operator/crds
popd
pushd ./vendor/github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1
controller-gen crd:crdVersions=v1 paths=. output:crd:dir=../../../../../../../../production/operator/crds
popd
```



