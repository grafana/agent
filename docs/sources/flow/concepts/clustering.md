---
title: Clustering
weight: 500
labels:
  stage: beta
---

# Clustering (beta)

Clustering enables Grafana Agent Flow to coordinate a fleet of agents working
together for workload distribution and high availability. It helps create
horizontally scalable deployments with minimal resource and operational
overhead.

To achieve this, Grafana Agent makes use of an eventually consistent model that
assumes all participating Agents are interchangeable and converge on using the
same configuration file.

The behavior of a standalone, non-clustered agent is the same as if it was a
single-node cluster.

## Use cases

[Setting up][] clustering using the command-line arguments is the first step in
making agents aware of one another.
Components in a telemetry pipeline need to explicitly opt-in to participating
to a clustering use case in their River config.

### Target auto-distribution

Target auto-distribution is the most basic use case of clustering; it allows
scraping components running on all peers to distribute scrape load between
themselves. All nodes must have access to the same service discovery APIs, and
the set of targets should converge on a timeline comparable to the scrape
interval.

Whenever a cluster state change is detected, either due to a new node joining
or an existing node going away, all participating components locally
recalculate target ownership and rebalance the number of targets they’re
scraping without explicitly communicating ownership over the network.

As such, target auto-distribution not only allows to dynamically scale the
number of agents to distribute workload during peaks, but also provides
resiliency, since in the event of a node going away, its targets are
automatically picked up by one of their peers. 

The agent makes use of a fully-local consistent hashing algorithm to distribute
targets, meaning that on average only ~1/N of the targets are redistributed.

The components who can make use of target auto-distribution are the following:
- [prometheus.scrape][]
- [pyroscope.scrape][]
- [prometheus.operator.podmonitors][]
- [prometheus.operator.servicemonitors][]

These components can opt-in to auto-distributing targets between nodes by
adding a `clustering` block with the `enabled` argument set to true:
```river
prometheus.scrape "default" {
    clustering {
      enabled = true
    }
    ...
}
```

## Local clustering example

The following commands bootstrap a three-node local cluster
```
# node-a (seed node)
$ ./build/grafana-agent-flow \
    --server.http.listen-addr=0.0.0.0:12345 \
    --cluster.enabled --cluster.node-name=node-a\
    --cluster.advertise-address=localhost:12345 \
    run local-clustering.river

# node-b
$ ./build/grafana-agent-flow \
    --server.http.listen-addr=0.0.0.0:12346 \
    --cluster.enabled --cluster.node-name=node-b \
    --cluster.advertise-address=localhost:12346 \
    --cluster.join-addresses=localhost:12345 \
    run local-clustering.river

# node-c
$ ./build/grafana-agent-flow \
    --server.http.listen-addr=0.0.0.0:12347 \
    --cluster.enabled --cluster.node-name=node-c \
    --cluster.advertise-address=localhost:12347 \
    --cluster.join-addresses=localhost:12345,localhost:12346 \
    run local-clustering.river
```

The clustering UI page of each node will provide information about the status
of its peers and the details page of the each node's `prometheus.scrape`
component will point to which target is being scraped by that node in its
Debug Info.

## Helm chart clustering example

The easiest way to deploy an agent cluster is by making use of our
[Helm chart][]. Here's an example of how to achieve that.

The following `values.yaml` file deploys a StatefulSet for metrics
collection. It makes use of a [headless service][] to retrieve the IPs of the
agent pods for the `--cluster.join-addresses` argument, as well as an
[Horizontal Pod Autoscaler][] (HPA) for dynamically matching demand.

```yaml
--- clustering-values.yaml ---
agent:
  mode: 'flow'
  configMap:
    # -- Create a new ConfigMap for the config file.
    create: true
    # -- Content to assign to the new ConfigMap. This is passed into `tpl` allowing for templating from values.
    content: ''

  # -- Address and port to listen for traffic on.
  # 0.0.0.0 exposes the HTTP server to other containers.
  listenAddr: 0.0.0.0
  listenPort: 80

  # -- Extra args to pass to `agent run`: https://grafana.com/docs/agent/latest/flow/reference/cli/run/
  extraArgs:
    - "--cluster.enabled"
    - "--cluster.join-addresses=grafana-agent"  # Uses the headless service name, which defaults to the installation name.

  # -- The minimum resources required for scheduling; required for using autoscaling.
  resources:
    requests:
      cpu: 100m
      memory: 200Mi

controller:
  type: 'statefulset'

  # -- Whether to enable automatic deletion of stale PVCs due to a scale down operation, when controller.type is 'statefulset'.
  enableStatefulSetAutoDeletePVC: true

  autoscaling:
    # -- Creates a HorizontalPodAutoscaler for controller type deployment.
    enabled: true
    # -- The lower limit for the number of replicas to which the autoscaler can scale down.
    minReplicas: 3
    # -- The upper limit for the number of replicas to which the autoscaler can scale up.
    maxReplicas: 8
    # -- Average Memory utilization across all relevant pods, a percentage of the requested value of the resource for the pods. Setting `targetMemoryUtilizationPercentage` to 0 disables Memory scaling.
    targetMemoryUtilizationPercentage: 80

service:
  # -- Creates a Headless Service for the controller's pods.
  enabled: true
  type: ClusterIP
  clusterIP: 'None'
```

Also, here's a simple River config file
```river
--- clustering.river ---
logging {
	level  = "info"
	format = "logfmt"
}

discovery.kubernetes "pods" {
	role = "pod"
}

prometheus.scrape "pods" {
	clustering {
		enabled = true
	}
	targets    = discovery.kubernetes.pods.targets
	forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default"{
	endpoint {
		url = env("PROMETHEUS_URL")
		basic_auth {
			username = env("PROMETHEUS_USERNAME")
			password = env("PROMETHEUS_API_KEY")
		}
	}
}
```

First, set up the Grafana chart repository.
```
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

Then, install the chart by using:
```
$ helm install --create-namespace --namespace agent grafana-agent. -f clustering-values.yaml --set-file agent.configMap.content=clustering.river
```

To upgrade the Helm installation with a new configuration file or values file:
```
$ helm upgrade --install --namespace agent grafana-agent . -f clustering-values.yaml --set-file agent.configMap.content=clustering.river
```

Use port-forwarding on the pods to see the UI in action. 
```
$ k port-forward grafana-agent-0 8080:80
```

The number of targets being by scraped by the `prometheus.scrape` component on
each pod will be automatically adjusted to share the load.

## Cluster monitoring and troubleshooting

To monitor your cluster status, you can check the Flow UI [clustering page][]
or install our [mixin][] to reuse our set of predefined dashboards and alerts.

The [debugging][] page contains some clues to help pin down clustering issues.


[Setting up]: {{< relref "../reference/cli/run.md#clustering-beta" >}}
[Helm chart]: https://artifacthub.io/packages/helm/grafana/grafana-agent
[Horizontal Pod Autoscaler]: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
[headless service]: https://kubernetes.io/docs/concepts/services-networking/service/#headless-services
[clustering page]: {{< relref "../monitoring/debugging.md#clustering-page" >}}
[mixin]: {{< relref "../monitoring/mixin.md" >}}
[debugging]: {{< relref "../monitoring/debugging.md#debugging-clustering-issues" >}}

[prometheus.scrape]: {{< relref "../reference/components/prometheus.scrape.md#clustering-beta" >}}
[pyroscope.scrape]: {{< relref "../reference/components/pyroscope.scrape.md#clustering-beta" >}}
[prometheus.operator.podmonitors]: {{< relref "../reference/components/prometheus.operator.podmonitors.md#clustering-beta" >}}
[prometheus.operator.servicemonitors]: {{< relref "../reference/components/prometheus.operator.servicemonitors.md#clustering-beta" >}}

