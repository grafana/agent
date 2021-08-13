local utils = import './utils.libsonnet';
local g = import 'grafana-builder/grafana.libsonnet';

{
  grafanaDashboards+:: {
    'agent-operational.json':
      utils.injectUtils(g.dashboard('Agent Operational'))
      .addMultiTemplate('cluster', 'agent_build_info', 'cluster')
      .addMultiTemplate('namespace', 'agent_build_info', 'namespace')
      .addMultiTemplate('container', 'agent_build_info', 'container')
      .addMultiTemplateWithAll('pod', 'agent_build_info{container=~"$container"}', 'pod', all='grafana-agent-.*')
      .addRow(
        g.row('General')
        .addPanel(
          g.panel('GCs') +
          g.queryPanel(
            'rate(go_gc_duration_seconds_count{cluster=~"$cluster", namespace=~"$namespace", container=~"$container", pod=~"$pod"}[5m])',
            '{{pod}}',
          )
        )
        .addPanel(
          g.panel('Go Heap') +
          { yaxes: g.yaxes('decbytes') } +
          g.queryPanel(
            'go_memstats_heap_inuse_bytes{cluster=~"$cluster", namespace=~"$namespace", container=~"$container", pod=~"$pod"}',
            '{{pod}}',
          )
        )
        .addPanel(
          g.panel('Goroutines') +
          g.queryPanel(
            'go_goroutines{cluster=~"$cluster", namespace=~"$namespace", container=~"$container", pod=~"$pod"}',
            '{{pod}}',
          )
        )
        .addPanel(
          g.panel('CPU') +
          g.queryPanel(
            'rate(container_cpu_usage_seconds_total{cluster=~"$cluster", container=~"$container", pod=~"$pod"}[5m])',
            '{{pod}}',
          )
        )
        .addPanel(
          g.panel('WSS') +
          g.queryPanel(
            'container_memory_working_set_bytes{cluster=~"$cluster", container=~"$container", pod=~"$pod"}',
            '{{pod}}',
          )
        )
        .addPanel(
          g.panel('Bad Words') +
          g.queryPanel(
            'rate(promtail_custom_bad_words_total{cluster=~"$cluster", exported_namespace=~"$namespace", exported_job=~"$job"}[5m])',
            '{{job}}',
          )
        )
      )
      .addRow(
        g.row('Network')
        .addPanel(
          g.panel('RX by Pod') +
          g.queryPanel(
            'sum by (pod) (rate(container_network_receive_bytes_total{cluster=~"$cluster", namespace=~"$namespace", pod=~"$pod"}[5m]))',
            '{{pod}}',
          )
        )
        .addPanel(
          g.panel('TX by Pod') +
          g.queryPanel(
            'sum by (pod) (rate(container_network_transmit_bytes_total{cluster=~"$cluster", namespace=~"$namespace", pod=~"$pod"}[5m]))',
            '{{pod}}',
          )
        )
      )
      .addRow(
        g.row('Prometheus Read')
        .addPanel(
          g.panel('Bytes/Series/Pod') +
          { yaxes: g.yaxes('decbytes') } +
          g.queryPanel(
            |||
              (sum by (pod) (avg_over_time(go_memstats_heap_inuse_bytes{cluster=~"$cluster", namespace=~"$namespace", container=~"$container", pod=~"$pod"}[1m])))
              /
              (sum by (pod) (agent_wal_storage_active_series{cluster=~"$cluster", namespace=~"$namespace", container=~"$container", pod=~"$pod"}))
            |||,
            '{{pod}}',
          )
        )
        .addPanel(
          g.panel('Bytes/Series') +
          { yaxes: g.yaxes('decbytes') } +
          g.queryPanel(
            |||
              (sum by (container) (avg_over_time(go_memstats_heap_inuse_bytes{cluster=~"$cluster", namespace=~"$namespace", container=~"$container", pod=~"$pod"}[1m])))
              /
              (sum by (container) (agent_wal_storage_active_series{cluster=~"$cluster", namespace=~"$namespace", container=~"$container", pod=~"$pod"}))
            |||,
            '{{container}}',
          )
        )
        .addPanel(
          g.panel('Series/Pod') +
          g.queryPanel(
            'sum by (pod) (agent_wal_storage_active_series{cluster=~"$cluster", namespace=~"$namespace", container=~"$container", pod=~"$pod"})',
            '{{pod}}',
          )
        )
        .addPanel(
          g.panel('Series/Config') +
          g.queryPanel(
            'sum by (instance_group_name) (agent_wal_storage_active_series{cluster=~"$cluster", namespace=~"$namespace", container=~"$container", pod=~"$pod"})',
            '{{instance_group_name}}',
          )
        )
        .addPanel(
          g.panel('Series') +
          g.queryPanel(
            'sum by (container) (agent_wal_storage_active_series{cluster=~"$cluster", namespace=~"$namespace", container=~"$container", pod=~"$pod"})',
            '{{container}}',
          )
        )
      ),
  },
}
