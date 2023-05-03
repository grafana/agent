local dashboard = import './utils/dashboard.jsonnet';
local panel = import './utils/panel.jsonnet';
local filename = 'agent-cluster-node.json';

{
  [filename]:
    dashboard.new(name='Grafana Agent Flow / Cluster Node') +
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
      dashboard.newTemplateVariable('instance', |||
        label_values(agent_component_controller_running_components{cluster="$cluster", namespace="$namespace"}, instance)
      |||),
    ]) +
    dashboard.withPanelsMixin([
      // Node Info row
      (
        panel.new('Node Info', 'row') +
        panel.withPosition({ h: 1, w: 24, x: 0, y: 0 })
      ),
      // Node Info
      (
        panel.new('Node Info', 'table') +
        panel.withDescription(|||
          Information about a specific cluster node.
        |||) +
        panel.withPosition({ x: 0, y: 1, w: 12, h: 8 }) +
        panel.withQueries([
          panel.newNamedInstantQuery(
            expr='sum(cluster_node_lamport_time{instance="$instance"})',
            refId='Lamport clock time',
            format='table',
          ),
          panel.newNamedInstantQuery(
            expr='sum(cluster_node_update_observers{instance="$instance"})',
            refId='Internal cluster state observers',
            format='table',
          ),
          panel.newNamedInstantQuery(
            expr='sum(cluster_node_gossip_health_score{instance="$instance"})',
            refId='Gossip health score',
            format='table',
          ),
          panel.newNamedInstantQuery(
            expr='sum(cluster_node_gossip_proto_version{instance="$instance"})',
            refId='Gossip protocol version',
            format='table',
          ),
        ]) +
        panel.withTransformations([
          {
            id: 'renameByRegex',
            options: {
              regex: 'Value #(.*)',
              renamePattern: '$1',
            },
          },
          {
            id: 'reduce',
            options: {},
          },
          {
            id: 'organize',
            options: {
              excludeByName: {},
              indexByName: {},
              renameByName: {
                Field: 'Metric',
                Max: 'Value',
              },
            },
          },
        ])
      ),
      // Gossip ops/sec
      (
        panel.new('Gossip ops/s', 'timeseries') +
        panel.withDescription(|||
          Gossip ops/sec for this node.
        |||) +
        panel.withPosition({ x: 12, y: 1, w: 12, h: 8 }) +
        panel.withQueries([
          panel.newQuery(
            expr='rate(cluster_node_gossip_received_events_total{instance="$instance"}[$__rate_interval])',
            legendFormat='{{event}}'
          ),
        ]) +

        panel.withOptions({
          legend: {
            calcs: [],
            displayMode: 'list',
            placement: 'bottom',
            showLegend: true,
          },
          tooltip: {
            mode: 'single',
            sort: 'none',
          },
        })
      ),
      // Known peers
      (
        panel.new('Known peers', 'stat') +
        panel.withDescription(|||
          Known peers to the node (including the local node).
        |||) +
        panel.withPosition({ x: 0, y: 9, w: 12, h: 8 }) +
        panel.withQueries([
          panel.newQuery(
            expr='sum(cluster_node_peers{instance="$instance"})',
          ),
        ]) +
        panel.withFieldConfigs({
          defaults: {
            unit: 'suffix:peers',
          },
        })
      ),
      // Peers by state
      (
        panel.new('Peers by state', 'timeseries') +
        panel.withDescription(|||
          Known peers to the node by state.
        |||) +
        panel.withPosition({ x: 12, y: 9, w: 12, h: 8 }) +
        panel.withQueries([
          panel.newQuery(
            expr='cluster_node_peers{instance="$instance"}',
            legendFormat='{{state}}',
          ),
        ]) +
        panel.withFieldConfigs({
          defaults: {
            unit: 'suffix:nodes',
          },
        })
      ),
      // Gossip Transport row
      (
        panel.new('Gossip Transport', 'row') +
        panel.withPosition({ h: 1, w: 24, x: 0, y: 17 })
      ),
      // Transport bandwidth
      (
        panel.new('Transport bandwidth', 'timeseries') +
        panel.withPosition({
          h: 8,
          w: 8,
          x: 0,
          y: 18,
        }) +
        panel.withQueries([
          panel.newQuery(
            expr='rate(cluster_transport_rx_bytes_total{instance="$instance"}[$__rate_interval])'
          ),
          panel.newQuery(
            expr='-1 * rate(cluster_transport_tx_bytes_total{instance="$instance"}[$__rate_interval])'
          ),
        ]) +
        panel.withFieldConfigs({
          defaults: {
            unit: 'Bps',
          },
          overrides: [],
        })
      ),
      // Packet write success rate
      (
        panel.new('Packet write success rate', 'timeseries') +
        panel.withPosition({
          h: 8,
          w: 8,
          x: 8,
          y: 18,
        }) +
        panel.withQueries([
          panel.newQuery(
            expr='1 - ( \r\n    rate(cluster_transport_tx_packets_failed_total{instance="$instance"}[$__rate_interval]) / \r\n    rate(cluster_transport_tx_packets_total{instance="$instance"}[$__rate_interval])\r\n)',
            legendFormat='Tx success %',
          ),
          panel.newQuery(
            expr='1 - ( \r\n    rate(cluster_transport_rx_packets_failed_total{instance="$instance"}[$__rate_interval]) / \r\n    rate(cluster_transport_rx_packets_total{instance="$instance"}[$__rate_interval])\r\n)',
            legendFormat='Rx success %',
          ),
        ]) +
        panel.withFieldConfigs({
          defaults: {
            unit: 'percentunit',
          },
          overrides: [],
        })
      ),
      // Pending packet queue
      (
        panel.new('Pending packet queue', 'timeseries') +
        panel.withPosition({
          h: 8,
          w: 8,
          x: 16,
          y: 18,
        }) +
        panel.withQueries([
          panel.newQuery(
            expr='cluster_transport_tx_packet_queue_length{instance="$instance"}',
            legendFormat='tx queue',
          ),
          panel.newQuery(
            expr='cluster_transport_rx_packet_queue_length{instance="$instance"}',
            legendFormat='rx queue',
          ),
        ]) +
        panel.withFieldConfigs({
          defaults: {
            unit: 'pkts',
          },
          overrides: [],
        })
      ),
      // Stream bandwidth
      (
        panel.new('Stream bandwidth', 'timeseries') +
        panel.withPosition({
          h: 8,
          w: 8,
          x: 0,
          y: 26,
        }) +
        panel.withQueries([
          panel.newQuery(
            expr='rate(cluster_transport_stream_rx_bytes_total{instance="$instance"}[$__rate_interval])',
            legendFormat='rx',
          ),
          panel.newQuery(
            expr='-1 * rate(cluster_transport_stream_tx_bytes_total{instance="$instance"}[$__rate_interval])',
            legendFormat='tx',
          ),
        ]) +
        panel.withFieldConfigs({
          defaults: {
            unit: 'Bps',
          },
          overrides: [],
        })
      ),
      // Stream write success rate
      (
        panel.new('Stream write success rate', 'timeseries') +
        panel.withPosition({
          h: 8,
          w: 8,
          x: 8,
          y: 26,
        }) +
        panel.withQueries([
          panel.newQuery(
            expr='1 - ( \r\n    rate(cluster_transport_stream_tx_packets_failed_total{instance="$instance"}[$__rate_interval]) / \r\n    rate(cluster_transport_stream_tx_packets_total{instance="$instance"}[$__rate_interval])\r\n)',
            legendFormat='Tx success %'
          ),
          panel.newQuery(
            expr='1 - ( \r\n    rate(cluster_transport_stream_rx_packets_failed_total{instance="$instance"}[$__rate_interval]) / \r\n    rate(cluster_transport_stream_rx_packets_total{instance="$instance"}[$__rate_interval])\r\n)',
            legendFormat='Rx success %'
          ),
        ]) +
        panel.withFieldConfigs({
          defaults: {
            unit: 'percentunit',
          },
          overrides: [],
        })
      ),
      // Open transport streams
      (
        panel.new('Open transport streams', 'timeseries') +
        panel.withPosition({
          h: 8,
          w: 8,
          x: 16,
          y: 26,
        }) +
        panel.withQueries([
          panel.newQuery(
            expr='cluster_transport_streams{instance="$instance"}',
            legendFormat='Open streams'
          ),
        ])
      ),
    ]),
}
