local alert = import './utils/alert.jsonnet';
local filename = 'clustering.json';

{
  [filename]:
    alert.newGroup([
      // Cluster not converging.
      alert.newRule(
        'ClusterNotConverged',
        'stddev((sum without (state) (cluster_node_peers))) != 0',
        'Cluster is not converging',
        '2m',
      ),

      // Clustering has entered a split brain state
      alert.newRule(
        'ClusterSplitBrain',
        '(sum without (state) (cluster_node_peers)) != count(cluster_node_info)',
        'Cluster nodes have entered a split brain state',
        '2m',
      ),
    ]),
}
