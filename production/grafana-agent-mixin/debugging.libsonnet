local g = import 'grafana-builder/grafana.libsonnet';

{
  grafanaDashboards+:: {
    'agent-operational.json':
      g.dashboard('Agent Operational')
      .addMultiTemplate('cluster', 'agent_build_info', 'cluster')
      .addMultiTemplate('namespace', 'agent_build_info', 'namespace')
      .addMultiTemplate('job', 'agent_build_info', 'job')
      .addMultiTemplate('instance', 'agent_build_info', 'instance')
      .addRow(
        g.row('General')
        .addPanel(
          g.panel('GCs') +
          g.queryPanel(
            'rate(go_gc_duration_seconds_count{cluster=~"$cluster", namespace=~"$namespace", job=~"$job"}[5m])',
            '{{job}}',
          )
        )
        .addPanel(
          g.panel('Go Heap') +
          { yaxes: g.yaxes('decbytes') } +
          { stack: 'true' } +
          g.queryPanel(
            'go_memstats_heap_inuse_bytes{cluster=~"$cluster", namespace=~"$namespace", job=~"$job"}',
            '{{job}}',
          )
        )
        .addPanel(
          g.panel('Goroutines') +
          g.queryPanel(
            'go_goroutines{cluster=~"$cluster", namespace=~"$namespace", job=~"$job"}',
            '{{job}}',
          )
        )
        .addPanel(
          g.panel('CPU') +
          g.queryPanel(
            'rate(container_cpu_usage_seconds_total{cluster=~"$cluster", pod_name=~".*grafana-agent.*"}[5m])',
            '{{job}}',
          )
        )
        .addPanel(
          g.panel('WSS') +
          g.queryPanel(
            'container_memory_working_set_bytes{cluster=~"$cluster", pod_name=~".*grafana-agent.*"}',
            '{{job}}',
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
          g.panel('RX') +
          g.queryPanel(
            'rate(container_network_receive_bytes_total{cluster=~"$cluster", namespace=~"$namespace", pod_name=~".*grafana-agent.*"}[5m])',
            '{{job}}',
          )
        )
        .addPanel(
          g.panel('TX') +
          g.queryPanel(
            'rate(container_network_transmit_bytes_total{cluster=~"$cluster", namespace=~"$namespace", pod_name=~".*grafana-agent.*"}[5m])',
            '{{job}}',
          )
        )
      )
      .addRow(
        g.row('Prometheus Read')
        .addPanel(
          g.panel('Bytes/Series/Instance') +
          { yaxes: g.yaxes('decbytes') } +
          { stack: 'true' } +
          g.queryPanel(
            |||
              (sum by (job, instance) (avg_over_time(go_memstats_heap_inuse_bytes{cluster=~"$cluster", job=~"$job", instance=~"$instance"}[1m])))
              /
              (sum by (job, instance) (agent_wal_storage_active_series{cluster=~"$cluster", job=~"$job", instance=~"$instance"}))
            |||,
            '{{instance}}',
          )
        )
        .addPanel(
          g.panel('Bytes/Series') +
          { yaxes: g.yaxes('decbytes') } +
          { stack: 'true' } +
          g.queryPanel(
            |||
              (sum by (job) (avg_over_time(go_memstats_heap_inuse_bytes{cluster=~"$cluster", job=~"$job", instance=~"$instance"}[1m])))
              /
              (sum by (job) (agent_wal_storage_active_series{cluster=~"$cluster", job=~"$job", instance=~"$instance"}))
            |||,
            '{{job}}',
          )
        )
        .addPanel(
          g.panel('Series/Instance') +
          { stack: 'true' } +
          g.queryPanel(
            'sum by (instance) (agent_wal_storage_active_series{cluster=~"$cluster", job=~"$job", instance=~"$instance"})',
            '{{instance}}',
          )
        )
        .addPanel(
          g.panel('Series/Config') +
          { stack: 'true' } +
          g.queryPanel(
            'sum by (instance_name) (agent_wal_storage_active_series{cluster=~"$cluster", job=~"$job", instance=~"$instance"})',
            '{{instance_name}}',
          )
        )
        .addPanel(
          g.panel('Series') +
          { stack: 'true' } +
          g.queryPanel(
            'sum by (job) (agent_wal_storage_active_series{cluster=~"$cluster", job=~"$job", instance=~"$instance"})',
            '{{job}}',
          )
        )
      ),
  },
}
