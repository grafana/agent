local g = import 'grafana-builder/grafana.libsonnet';

{
  grafanaDashboards+:: {
    'agent-debugging.json':
      g.dashboard('Agent Debugging')
      .addMultiTemplate('job', 'agent_build_info', 'job')
      .addMultiTemplate('instance', 'agent_build_info', 'instance')
      .addRow(
        g.row('Memory')
        .addPanel(
          g.panel('Memory Inuse Avg') +
          { yaxes: g.yaxes('decbytes') } +
          g.queryPanel(
            'avg by (job, instance) (avg_over_time(go_memstats_heap_inuse_bytes{job=~"$job", instance=~"$instance"}[5m]))',
            '{{job}}',
          )
        )
        .addPanel(
          g.panel('Memory / Series') +
          { yaxes: g.yaxes('decbytes') } +
          g.queryPanel(
            |||
              (avg by (job, instance) (avg_over_time(go_memstats_heap_inuse_bytes{job=~"$job", instance=~"$instance"}[5m])))
              /
              (sum by (job, instance) (agent_wal_storage_active_series{job=~"$job", instance=~"$instance"}))
            |||,
            '{{job}}',
          )
        )
        .addPanel(
          g.panel('Active Series') +
          g.queryPanel(
            'agent_wal_storage_active_series{job=~"$job", instance=~"$instance"}',
            '{{job}}',
          )
        ),
      ),
  },
}
