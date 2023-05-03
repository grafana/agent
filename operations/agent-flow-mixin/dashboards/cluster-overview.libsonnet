local dashboard = import './utils/dashboard.jsonnet';
local panel = import './utils/panel.jsonnet';
local filename = 'agent-cluster-overview.json';

{
  [filename]:
    dashboard.new(name='Grafana Agent Flow / Cluster Overview') +
    dashboard.withDocsLink(
      url='https://grafana.com/docs/agent/latest/flow/reference/cli/run/#clustered-mode-experimental',
      desc='Clustering documentation',
    ) +
    dashboard.withDashboardsLink() +
    dashboard.withUID(std.md5(filename)) +
    dashboard.withTemplateVariablesMixin([
      dashboard.newTemplateVariable('cluster', |||
        label_values(agent_component_controller_running_components, cluster)
      |||),
      dashboard.newTemplateVariable('namespace', |||
        label_values(agent_component_controller_running_components{cluster="$cluster"}, namespace)
      |||),
    ]) +
    dashboard.withPanelsMixin([
      // Nodes
      (
        panel.new('Nodes', 'stat') +
        panel.withDescription(|||
          Node count.
        |||) +
        panel.withPosition({ h: 9, w: 8, x: 0, y: 0 }) +
        panel.withQueries([
          panel.newInstantQuery(
            expr='count(cluster_node_info)'
          ),
        ])
      ),
      // Node table
      (
        panel.new('Node table', 'table') +
        panel.withDescription(|||
          Nodes info.
        |||) +
        panel.withPosition({ h: 18, w: 16, x: 8, y: 0 }) +
        panel.withQueries([
          panel.newInstantQuery(
            expr='cluster_node_info',
            format='table',
          ),
        ]) +
        panel.withTransformations([
          {
            id: 'organize',
            options: {
              excludeByName: {
                Time: true,
                Value: false,
                __name__: true,
                cluster: true,
                namespace: true,
                state: false,
              },
              indexByName: {},
              renameByName: {
                Value: 'Dashboard',
                instance: '',
                state: '',
              },
            },
          },
        ]) +
        panel.withFieldConfigs({
          overrides: [
            {
              matcher: {
                id: 'byName',
                options: 'Dashboard',
              },
              properties: [
                {
                  id: 'mappings',
                  value: [
                    {
                      options: {
                        '1': {
                          index: 0,
                          text: 'Link',
                        },
                      },
                      type: 'value',
                    },
                  ],
                },
                {
                  id: 'links',
                  value: [
                    {
                      targetBlank: false,
                      title: 'Detail dashboard for node',
                      url: '/d/agent-cluster-node/grafana-agent-flow-cluster-node?var-instance=${__data.fields.instance}&var-datasource=${datasource}&var-loki_datasource=${loki_datasource}&var-cluster=${cluster}&var-namespace=${namespace}',
                    },
                  ],
                },
              ],
            },
          ],
        })
      ),
      // Convergance state
      (
        panel.new('Convergance state', 'stat') +
        panel.withDescription(|||
          "Whether the cluster state has converged.

          It is normal for the cluster state to be diverged briefly as gossip events propagate. It is not normal for the cluster state to be diverged for a long period of time.

          This will show one of the following:

          * Converged: Nodes are aware of all other nodes, with the correct states.
          * Not converged: A subset of nodes aren't aware of their peers, or don't have an updated view of peer states."
        |||) +
        panel.withPosition({ h: 9, w: 8, x: 0, y: 9 }) +
        panel.withQueries([
          panel.newInstantQuery(
            expr='clamp((\r\n  sum(stddev by (state) (cluster_node_peers) != 0) or \r\n  (sum(abs(sum without (state) (cluster_node_peers)) - scalar(count(cluster_node_info)) != 0))\r\n), 1, 1)',
            format='time_series'
          ),
        ]) +
        panel.withOptions(
          {
            colorMode: 'background',
            graphMode: 'none',
            justifyMode: 'auto',
            orientation: 'auto',
            reduceOptions: {
              calcs: [
                'lastNotNull',
              ],
              fields: '',
              values: false,
            },
            textMode: 'auto',
          }
        ) +
        panel.withFieldConfigs(
          {
            defaults: {
              color: {
                mode: 'thresholds',
              },
              mappings: [
                {
                  options: {
                    '1': {
                      color: 'red',
                      index: 1,
                      text: 'Not converged',
                    },
                  },
                  type: 'value',
                },
                {
                  options: {
                    match: 'null',
                    result: {
                      color: 'green',
                      index: 0,
                      text: 'Converged',
                    },
                  },
                  type: 'special',
                },
              ],
              thresholds: {
                mode: 'absolute',
                steps: [
                  {
                    color: 'green',
                    value: null,
                  },
                ],
              },
              unit: 'suffix:nodes',
            },
            overrides: [],
          }
        )
      ),
    ]),
}
