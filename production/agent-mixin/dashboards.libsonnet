local g = import 'grafana-builder/grafana.libsonnet';

{
  grafanaDashboards+:: {
    'agent.json':
      g.dashboard('Prometheus')
      .addMultiTemplate('job', 'agent_build_info', 'job')
      .addMultiTemplate('instance', 'agent_build_info', 'instance')
      .addRow(
        g.row('Prometheus Stats')
        .addPanel(
          g.panel('Prometheus Stats') +
          g.tablePanel([
            'count by (job, instance, version) (agent_build_info{job=~"$job", instance=~"$instance"})',
            'max by (job, instance) (time() - process_start_time_seconds{job=~"$job", instance=~"$instance"})',
          ], {
            job: { alias: 'Job' },
            instance: { alias: 'Instance' },
            version: { alias: 'Version' },
            'Value #A': { alias: 'Count', type: 'hidden' },
            'Value #B': { alias: 'Uptime' },
          })
        )
      )
      .addRow(
        g.row('Discovery')
        .addPanel(
          g.panel('Target Sync') +
          g.queryPanel('sum(rate(prometheus_target_sync_length_seconds_sum{job=~"$job",instance=~"$instance"}[5m])) by (scrape_job) * 1e3', '{{scrape_job}}') +
          { yaxes: g.yaxes('ms') }
        )
        .addPanel(
          g.panel('Targets') +
          g.queryPanel('sum(prometheus_sd_discovered_targets{job=~"$job",instance=~"$instance"})', 'Targets') +
          g.stack
        )
      )
      .addRow(
        g.row('Retrieval')
        .addPanel(
          g.panel('Average Scrape Interval Duration') +
          g.queryPanel('rate(prometheus_target_interval_length_seconds_sum{job=~"$job",instance=~"$instance"}[5m]) / rate(prometheus_target_interval_length_seconds_count{job=~"$job",instance=~"$instance"}[5m]) * 1e3', '{{interval}} configured') +
          { yaxes: g.yaxes('ms') }
        )
        .addPanel(
          g.panel('Scrape failures') +
          g.queryPanel([
            'sum by (job) (rate(prometheus_target_scrapes_exceeded_sample_limit_total[1m]))',
            'sum by (job) (rate(prometheus_target_scrapes_sample_duplicate_timestamp_total[1m]))',
            'sum by (job) (rate(prometheus_target_scrapes_sample_out_of_bounds_total[1m]))',
            'sum by (job) (rate(prometheus_target_scrapes_sample_out_of_order_total[1m]))',
          ], [
            'exceeded sample limit: {{job}}',
            'duplicate timestamp: {{job}}',
            'out of bounds: {{job}}',
            'out of order: {{job}}',
          ]) +
          g.stack
        )
        .addPanel(
          g.panel('Appended Samples') +
          g.queryPanel('rate(prometheus_tsdb_head_samples_appended_total{job=~"$job",instance=~"$instance"}[5m])', '{{job}} {{instance}}') +
          g.stack
        )
      ),
    // Remote write specific dashboard.
    'agent-prometheus-remote-write.json':
      g.dashboard('Agent Prometheus Remote Write')
      .addMultiTemplate('instance', 'agent_build_info', 'instance')
      .addMultiTemplate('cluster', 'agent_build_info', 'cluster')  // NOTE: this used to use kube_pod_container_info, might need to change it back
      .addRow(
        g.row('Timestamps')
        .addPanel(
          g.panel('Highest Timestamp In vs. Highest Timestamp Sent') +
          g.queryPanel('prometheus_remote_storage_highest_timestamp_in_seconds{cluster=~"$cluster", instance=~"$instance"} - ignoring(queue) group_right(instance) prometheus_remote_storage_queue_highest_sent_timestamp_seconds{cluster=~"$cluster", instance=~"$instance"}', '{{cluster}}:{{instance}}-{{queue}}') +
          { yaxes: g.yaxes('s') }
        )
        .addPanel(
          g.panel('Rate[5m]') +
          g.queryPanel('rate(prometheus_remote_storage_highest_timestamp_in_seconds{cluster=~"$cluster", instance=~"$instance"}[5m])  - ignoring (queue) group_right(instance) rate(prometheus_remote_storage_queue_highest_sent_timestamp_seconds{cluster=~"$cluster", instance=~"$instance"}[5m])', '{{cluster}}:{{instance}}-{{queue}}')
        )
      )
      .addRow(
        g.row('Samples')
        .addPanel(
          g.panel('Rate, in vs. succeeded or dropped [5m]') +
          g.queryPanel('rate(prometheus_remote_storage_samples_in_total{cluster=~"$cluster", instance=~"$instance"}[5m])- ignoring(queue) group_right(instance) rate(prometheus_remote_storage_succeeded_samples_total{cluster=~"$cluster", instance=~"$instance"}[5m]) - rate(prometheus_remote_storage_dropped_samples_total{cluster=~"$cluster", instance=~"$instance"}[5m])', '{{cluster}}:{{instance}}-{{queue}}')
        )
      )
      .addRow(
        g.row('Shards')
        .addPanel(
          g.panel('Num. Shards') +
          g.queryPanel('prometheus_remote_storage_shards{cluster=~"$cluster", instance=~"$instance"}', '{{cluster}}:{{instance}}-{{queue}}')
        )
        .addPanel(
          g.panel('Capacity') +
          g.queryPanel('prometheus_remote_storage_shard_capacity{cluster=~"$cluster", instance=~"$instance"}', '{{cluster}}:{{instance}}-{{queue}}')
        )
      )
      .addRow(
        g.row('Misc Rates.')
        .addPanel(
          g.panel('Dropped Samples') +
          g.queryPanel('rate(prometheus_remote_storage_dropped_samples_total{cluster=~"$cluster", instance=~"$instance"}[5m])', '{{cluster}}:{{instance}}-{{queue}}')
        )
        .addPanel(
          g.panel('Failed Samples') +
          g.queryPanel('rate(prometheus_remote_storage_failed_samples_total{cluster=~"$cluster", instance=~"$instance"}[5m])', '{{cluster}}:{{instance}}-{{queue}}')
        )
        .addPanel(
          g.panel('Retried Samples') +
          g.queryPanel('rate(prometheus_remote_storage_retried_samples_total{cluster=~"$cluster", instance=~"$instance"}[5m])', '{{cluster}}:{{instance}}-{{queue}}')
        )
        .addPanel(
          g.panel('Enqueue Retries') +
          g.queryPanel('rate(prometheus_remote_storage_enqueue_retries_total{cluster=~"$cluster", instance=~"$instance"}[5m])', '{{cluster}}:{{instance}}-{{queue}}')
        )
      ),
  },
}
