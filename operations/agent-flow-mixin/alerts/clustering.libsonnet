local alert = import './utils/alert.jsonnet';
local filename = 'clustering.json';

{
  spec: {
    rules: [
      {
        alert: 'ClusterNotConverged',
        expr: 'stddev(cluster_node_peers) != 0',
        'for': '5m',
        labels: {
          severity: 'critical',
        },
        annotations: {
          message: 'Cluster is not converging',
        },
      },
      {
        alert: 'ClusterSplitBrain',
        expr: '(sum without (state) (cluster_node_peers)) != count(cluster_node_info)',
        'for': '5m',
        labels: {
          severity: 'critical',
        },
        annotations: {
          message: 'Cluster nodes have entered a split brain state',
        },

      },
    ],
  },
}
