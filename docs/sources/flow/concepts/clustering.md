---
title: Clustering
weight: 500
labels:
  stage: beta
---

# Clustering (beta)

Agent Clustering enables Grafana Agent Flow to coordinate a fleet of agents to
work together for workload distribution and high availability. It makes agent
deployments horizontally scalable by default, and minimizes the operational
overhead of managing your agent infrastructure to handle demand.

To achieve this, Grafana Agent makes use of an eventually consistent model and
the lightweight [grafana/ckit][] framework. All running nodes are assumed to
use the same configuration and have access to the same type of hardware and
network resources such as PVCs and service discovery APIs.

Clustering is built from the ground-up to be the de-facto answer to "how do I
scale my Grafana Agent setup" in the agent's future. As an example, in
comparison to a horizontally-scalable setup using [hashmod sharding][],
clustering with target auto-distribution scales and provides resiliency at
_half_ the cost in resources and without the need of changing configuration
files for resharding.

The behavior of a standalone, non-clustered agent is the same as if it was a
1-node cluster.

## Usage

The following set of command-line flags are used to configure clustering when
starting up Grafana Agent:
- `--server.http.listen-addr`: the address where the agent’s HTTP server
  listens to.
- `--cluster.enabled`: enables cluster awareness.
- `--cluster.node-name`: defines the name used by the cluster node. If not
  provided, it defaults to the machine's hostname.
- `--cluster.advertise-address`: defines the address the agent advertises for
  its peers to connect to. If not provided, it is inferred automatically.
- `--cluster.join-addresses`: accepts a comma-separated list of addresses to
  join the cluster at. These can be IP addresses with an optional port, or a
DNS record to lookup.

Cluster communication happens over HTTP/2 on the agents’ HTTP listener. The
agents must be configured to accept connections from one another on the address
defined by the `--server.http.listen-addr` flag.

Each cluster member’s name must be unique within the cluster. Nodes which try
to join with a conflicting name are rejected.

If the advertised address is not explicitly set, the agent tries to find a
suitable one from the `eth0` and `en0` local network interfaces.

The ports on the join-addresses list default to the port of the node’s HTTP
listener if not explicitly provided; it’s generally recommended to align the
port numbers on as many nodes as possible to simplify the deployment process.

Finally, the first node that is used to bootstrap a new cluster (also known as
the "seed node") can either omit specifying the flag that specifies peers to
join or can try to connect to itself.

## Deploy clustering using the Helm chart

The easiest way to deploy agent clustering is by making use of our
[Helm chart][].

The following `values.yaml` file deploys an agent StatefulSet for metrics
collection. It makes use of a [headless service][] to retrieve the IPs of the
agent pods for the `--cluster.join-addresses` argument, as well as an HPA for
autoscaling.

```yaml
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

rbac:
  # -- Whether to create RBAC resources for the agent.
  create: true

serviceAccount:
  # -- Whether to create a service account for the Grafana Agent deployment.
  create: true

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

First, set up the Grafana chart repository.
```
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

Then, install the chart on your Kubernetes cluster by using the following
command:
```
$ helm install --create-namespace --namespace NAMESPACE INSTALL_NAME . -f VALUES --set-file agent.configMap.content=CONFIG_FILE
```

To upgrade your helm installation with a new configuration file or values
override:
```
$ helm upgrade --install --namespace NAMESPACE INSTALL_NAME . -f VALUES --set-file agent.configMap.content=CONFIG_FILE
```

In the above example we use `grafana-agent` as the install name, which is used
to name all resources, such as the headless service we’re passing in the
`--cluster.join-addresses` flag.

For any other installation name, the resources use Helm’s default name
generation and are called after `INSTALL_NAME-grafana-agent`.

Keep in mind, when using a statefulset, autoscaling with an HPA can lead to up
to `maxReplicas` PVCs leaking when the HPA is scaling down. If you're on
Kubernetes version `>=1.23-0` and your cluster has the
`StatefulSetAutoDeletePVC` feature gate enabled, you can set
`enableStatefulSetAutoDeletePVC` to true to automatically delete stale PVCs.

## Cluster meta-monitoring

A first way to view the state of an agent cluster is through the Flow UI.
The dedicated Clustering page shows the current node, and lists its known peers
with their addresses and current state.

![](../../../assets/ui_clustering_page.png)

For a more production-ready setup, we recommend taking a look at our
[Flow mixin][]; it contains a set of predefined dashboards and alerts for
monitoring clustered agent deployments, allowing to both get an overview of the
current state of the cluster, as well as easily drill down to node-level
information with the a click of a button.

![](../../../assets/clustering_overview_dashboard.png)
![](../../../assets/clustering_node_info_dashboard.png)
![](../../../assets/clustering_node_transport_dashboard.png)

To use the mixin, you first need to install [mixtool][]. Then, clone the
`grafana/agent` repo, and run `make build-mixin` from the repo root. The compiled
mixin is available on the `operations/agent-flow-mixin-compiled`
directory. You can import the JSON dashboards into your Grafana instance and
upload the alerts on Prometheus.

```
$ go install github.com/monitoring-mixins/mixtool/cmd/mixtool@main
$ git clone https://github.com/grafana/agent.git
$ cd agent
$ make build-mixin
$ tree operations/agent-flow-mixin-compiled
operations/agent-flow-mixin-compiled
├── alerts.yaml
└── dashboards
    ├── agent-cluster-node.json
    ├── agent-cluster-overview.json
    ├── agent-flow-controller.json
    ├── agent-flow-prometheus.remote_write.json
    └── agent-flow-resources.json
```

The compiled mixin is packaged on `operations/agent-flow-mixin.zip`.

## Cluster troubleshooting

Our Flow mixin contains a set of opinionated dashboards and alerts for
monitoring the status of your clusters to help pin down any issues with
clustering.

Here’s the list of some possible issues and what to keep an eye out for.

- **Cluster not converging**: The cluster peers are not converging on the same
  view of their peers' status. Check the "Gossip Transport" row to verify that
incoming and outgoing network requests are succeeding. Check the "Gossip ops/s"
panel to verify that gossip messages are being exchanged, and the "Peers by
state" panel to understand which nodes are not being picked up. This is most
likely due to network connectivity issues between the cluster nodes.
- **Cluster split brain**: The cluster peers are not aware of one another,
  thinking they’re the only node present. Again, check for network connectivity
issues. Check that the addresses or DNS names given in the comma-separated list
on `--cluster.join-addresses` are correctly formatted and reachable and  that
messages are being exchanged between peers.
- **Configuration drift**: Clustering assumes that all nodes are running with the
  same configuration file and that configuration changes converge in a time
scale comparable to the scrape interval (~1m). Check whether the
`config-reloader` container is working properly, as well as pod logs for any
issues with the reloaded configuration file.
- **Node name conflicts**: A new node tried to join the cluster with a
  conflicting name. Cluster peers need to have unique names; the
`--cluster.node-name` command-line flag defaults to the machine’s hostname but
can be used to override the name of the node. If you’re using a StatefulSet
which reuses pod names, check whether the previous pod has already gone away.
Check the "Peers by state" panel to check when the conflict event was first
seen.
- **Node stuck in terminating state**: The node attempted to gracefully shut
  down, set its state to Terminating but has not completely gone away. Check
the "Peers by state" panel to verify the status of the other cluster peers.
Check whether the reporting node is correctly gossiping messages with its
peers. Check whether the pod itself has gone away or has remained in the
cluster as Terminating.
- **Lamport clock stuck or drifting**: The node is either not receiving new
messages from its peer, or it cannot keep up with the rate of messages being
sent by the rest of the cluster. Check the ""Packet write success rate" and
"Pending packet queue" panels to verify that messages are being decoded
correctly and are decoded in time.

## Use cases

Setting up clustering is the first step of making agents aware of one another.
_Components_ in a telemetry pipeline need to explicitly opt-in to participate
in one or more clustering use cases using the `clustering` block in their
River config.

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

The agent makes use of a fully-local consistent hashing algorithm to distribute
targets, meaning that on average only ~1/N of the targets are redistributed.
This is in contrast to hashmod sharding where up to 100% of the targets could
be reassigned to another node and possibly cause system instability.

As such, target auto-distribution not only allows to dynamically scale the
number of agents to distribute workloads during peaks, but also provides
resiliency, since in the event of a node going away, its targets get
automatically picked up by one of their peers. Again, this is in contrast to
hashmod sharding which requires running multiple replicas of each shard for HA,
leading to increased costs and resource usage.

The components who can make use of target auto-distribution are the following:
- [prometheus.scrape][]
- [pyroscope.scrape][]
- [prometheus.operator.podmonitors][]
- [prometheus.operator.servicemonitors][]

These components can opt-in to participating in clustering and
auto-distributing targets between nodes by defining the `clustering` block. For
example:
```river
prometheus.scrape "default" {
    clustering {
      enabled = true
    }
    ...
}
```

[grafana/ckit]: "https://github.com/grafana/ckit"
[hashmod sharding]: "https://grafana.com/docs/agent/latest/static/operation-guide/#hashmod-sharding-stable"
[Helm chart]: "https://artifacthub.io/packages/helm/grafana/grafana-agent"
[headless service]: "https://kubernetes.io/docs/concepts/services-networking/service/#headless-services"
[Flow mixin]: "https://github.com/grafana/agent/tree/main/operations/agent-flow-mixin"
[mixtool]: "https://github.com/monitoring-mixins/mixtool"

[prometheus.scrape]: {{< relref "../reference/components/prometheus.scrape.md#clustering-beta" >}}
[pyroscope.scrape]: {{< relref "../reference/components/pyroscope.scrape.md#clustering-beta" >}}
[prometheus.operator.podmonitors]: {{< relref "../reference/components/prometheus.operator.podmonitors.md#clustering-beta" >}}
[prometheus.operator.servicemonitors]: {{< relref "../reference/components/prometheus.operator.servicemonitors.md#clustering-beta" >}}

