local alert = import './utils/alert.jsonnet';

alert.newGroup(
  'clustering',
  [
    // Cluster not converging.
    alert.newRule(
      'ClusterNotConverging',
      'stddev by (cluster, namespace) (sum without (state) (cluster_node_peers)) != 0',
      'Cluster is not converging.',
      '5m',
    ),

    // Cluster has entered a split brain state.
    alert.newRule(
      'ClusterSplitBrain',
      // Assert that the set of known peers (regardless of state) for an
      // agent matches the same number of running agents in the same cluster
      // and namespace.
      |||
        sum without (state) (cluster_node_peers) !=
        on (cluster, namespace) group_left
        count by (cluster, namespace) (cluster_node_info)
      |||,
      'Cluster nodes have entered a split brain state.',
      '5m',
    ),

    // Standard Deviation of Lamport clock time between nodes is too high.
    alert.newRule(
      'ClusterLamportClockDrift',
      'stddev by (cluster, namespace) (cluster_node_lamport_time) > 4 * sqrt(count by (cluster, namespace) (cluster_node_info))',
      "Cluster nodes' lamport clocks are not converging.",
      '5m'
    ),

    // Nodes health score is not zero.
    alert.newRule(
      'ClusterNodeUnhealthy',
      |||
        cluster_node_gossip_health_score > 0
      |||,
      'Cluster node is reporting a health score > 0.',
      '5m',
    ),

    // Lamport clock of a node is not progressing at all.
    //
    // This only checks for nodes that have peers other than themselves; nodes
    // with no external peers will not increase their lamport time because
    // there is no cluster networking traffic.
    alert.newRule(
      'ClusterLamportClockStuck',
      |||
        sum by (cluster, namespace, instance) (rate(cluster_node_lamport_time[2m])) == 0
        and on (cluster, namespace, instance) (cluster_node_peers > 1)
      |||,
      "Cluster nodes's lamport clocks is not progressing.",
      '5m',
    ),

    // Node tried to join the cluster with an already-present node name.
    alert.newRule(
      'ClusterNodeNameConflict',
      'sum by (cluster, namespace) (rate(cluster_node_gossip_received_events_total{event="node_conflict"}[2m])) > 0',
      'A node tried to join the cluster with a name conflicting with an existing peer.',
      '10m',
    ),

    // Node stuck in Terminating state.
    alert.newRule(
      'ClusterNodeStuckTerminating',
      'sum by (cluster, namespace, instance) (cluster_node_peers{state="terminating"}) > 0',
      'Cluster node stuck in Terminating state.',
      '5m',
    ),

    // Nodes are not using the same configuration file.
    alert.newRule(
      'ClusterConfigurationDrift',
      |||
        count without (sha256) (
            max by (cluster, namespace, sha256) (agent_config_hash and on(cluster, namespace) cluster_node_info)
        ) > 1
      |||,
      'Cluster nodes are not using the same configuration file.',
      '5m',
    ),

    // TODO(@tpaschalis) Alert on open transport streams once we investigate
    // their behavior.
  ]
)
