local alert = import './utils/alert.jsonnet';
local filename = 'clustering.json';

{
  [filename]:
    alert.newGroup([
      // Cluster not converging.
      alert.newRule(
        'ClusterNotConverged',
        'stddev by (cluster, namespace) ((sum without (state) (cluster_node_peers))) != 0',
        'Cluster is not converging.',
        '5m',
      ),

      // Clustering has entered a split brain state
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
    ]),
}
