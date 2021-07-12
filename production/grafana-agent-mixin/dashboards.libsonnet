local g = import 'grafana-builder/grafana.libsonnet';
local utils = import './utils.libsonnet';
local grafana = import 'grafonnet/grafana.libsonnet';

local dashboard = grafana.dashboard;
local row = grafana.row;
local singlestat = grafana.singlestat;
local prometheus = grafana.prometheus;
local graphPanel = grafana.graphPanel;
local tablePanel = grafana.tablePanel;
local template = grafana.template;

{
  grafanaDashboards+:: {
    'agent.json':
      utils.injectUtils(g.dashboard('Agent'))
      .addMultiTemplate('cluster', 'agent_build_info', 'cluster')
      .addMultiTemplate('namespace', 'agent_build_info', 'namespace')
      .addMultiTemplate('container', 'agent_build_info', 'container')
      .addMultiTemplateWithAll('pod', 'agent_build_info{container=~"$container"}', 'pod', all='grafana-agent-.*')
      .addRow(
        g.row('Agent Stats')
        .addPanel(
          g.panel('Agent Stats') +
          g.tablePanel([
            'count by (pod, container, version) (agent_build_info{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"})',
            'max by (pod, container) (time() - process_start_time_seconds{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"})',
          ], {
            pod: { alias: 'Pod' },
            container: { alias: 'Container' },
            version: { alias: 'Version' },
            'Value #A': { alias: 'Count', type: 'hidden' },
            'Value #B': { alias: 'Uptime' },
          })
        )
      )
      .addRow(
        g.row('Prometheus Discovery')
        .addPanel(
          g.panel('Target Sync') +
          g.queryPanel('sum(rate(prometheus_target_sync_length_seconds_sum{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])) by (pod, scrape_job) * 1e3', '{{pod}}/{{scrape_job}}') +
          { yaxes: g.yaxes('ms') }
        )
        .addPanel(
          g.panel('Targets') +
          g.queryPanel('sum(prometheus_sd_discovered_targets{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"})', 'Targets') +
          g.stack
        )
      )
      .addRow(
        g.row('Prometheus Retrieval')
        .addPanel(
          g.panel('Average Scrape Interval Duration') +
          g.queryPanel(|||
            rate(prometheus_target_interval_length_seconds_sum{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])
            /
            rate(prometheus_target_interval_length_seconds_count{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])
            * 1e3
          |||, '{{pod}} {{interval}} configured') +
          { yaxes: g.yaxes('ms') }
        )
        .addPanel(
          g.panel('Scrape failures') +
          g.queryPanel([
            'sum by (job) (rate(prometheus_target_scrapes_exceeded_sample_limit_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[1m]))',
            'sum by (job) (rate(prometheus_target_scrapes_sample_duplicate_timestamp_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[1m]))',
            'sum by (job) (rate(prometheus_target_scrapes_sample_out_of_bounds_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[1m]))',
            'sum by (job) (rate(prometheus_target_scrapes_sample_out_of_order_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[1m]))',
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
          g.queryPanel('sum by (job, instance_group_name) (rate(agent_wal_storage_samples_appended_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m]))', '{{job}} {{instance_group_name}}') +
          g.stack
        )
      ),

    // Remote write specific dashboard.
    'agent-remote-write.json':
      local timestampComparison =
        graphPanel.new(
          'Highest Timestamp In vs. Highest Timestamp Sent',
          datasource='$datasource',
          span=6,
        )
        .addTarget(prometheus.target(
          |||
            (
              prometheus_remote_storage_highest_timestamp_in_seconds{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}
              -
              ignoring(url, instance_group_name, remote_name) group_right(pod)
              prometheus_remote_storage_queue_highest_sent_timestamp_seconds{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}
            )
          |||,
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local timestampComparisonRate =
        graphPanel.new(
          'Rate[5m]',
          datasource='$datasource',
          span=6,
        )
        .addTarget(prometheus.target(
          |||
            (
              rate(prometheus_remote_storage_highest_timestamp_in_seconds{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])
              -
              ignoring(url, instance_group_name, remote_name) group_right(pod)
              rate(prometheus_remote_storage_queue_highest_sent_timestamp_seconds{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])
            )
          |||,
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local samplesRate =
        graphPanel.new(
          'Rate, in vs. succeeded or dropped [5m]',
          datasource='$datasource',
          span=12,
        )
        .addTarget(prometheus.target(
          |||
            rate(
              prometheus_remote_storage_samples_in_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])
            -
              ignoring(url, instance_group_name, remote_name) group_right(pod)
              rate(prometheus_remote_storage_succeeded_samples_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])
            -
              rate(prometheus_remote_storage_dropped_samples_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])
          |||,
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local currentShards =
        graphPanel.new(
          'Current Shards',
          datasource='$datasource',
          span=12,
          min_span=6,
        )
        .addTarget(prometheus.target(
          'prometheus_remote_storage_shards{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local maxShards =
        graphPanel.new(
          'Max Shards',
          datasource='$datasource',
          span=4,
        )
        .addTarget(prometheus.target(
          'prometheus_remote_storage_shards_max{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local minShards =
        graphPanel.new(
          'Min Shards',
          datasource='$datasource',
          span=4,
        )
        .addTarget(prometheus.target(
          'prometheus_remote_storage_shards_min{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local desiredShards =
        graphPanel.new(
          'Desired Shards',
          datasource='$datasource',
          span=4,
        )
        .addTarget(prometheus.target(
          'prometheus_remote_storage_shards_desired{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local shardsCapacity =
        graphPanel.new(
          'Shard Capacity',
          datasource='$datasource',
          span=6,
        )
        .addTarget(prometheus.target(
          'prometheus_remote_storage_shard_capacity{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local pendingSamples =
        graphPanel.new(
          'Pending Samples',
          datasource='$datasource',
          span=6,
        )
        .addTarget(prometheus.target(
          'prometheus_remote_storage_pending_samples{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local queueSegment =
        graphPanel.new(
          'Remote Write Current Segment',
          datasource='$datasource',
          span=6,
          formatY1='none',
        )
        .addTarget(prometheus.target(
          'prometheus_wal_watcher_current_segment{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local droppedSamples =
        graphPanel.new(
          'Dropped Samples',
          datasource='$datasource',
          span=3,
        )
        .addTarget(prometheus.target(
          'rate(prometheus_remote_storage_dropped_samples_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local failedSamples =
        graphPanel.new(
          'Failed Samples',
          datasource='$datasource',
          span=3,
        )
        .addTarget(prometheus.target(
          'rate(prometheus_remote_storage_failed_samples_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local retriedSamples =
        graphPanel.new(
          'Retried Samples',
          datasource='$datasource',
          span=3,
        )
        .addTarget(prometheus.target(
          'rate(prometheus_remote_storage_retried_samples_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local enqueueRetries =
        graphPanel.new(
          'Enqueue Retries',
          datasource='$datasource',
          span=3,
        )
        .addTarget(prometheus.target(
          'rate(prometheus_remote_storage_enqueue_retries_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      dashboard.new('Agent Prometheus Remote Write', tags=['grafana-agent-mixin'], editable=true)
      .addTemplate(
        {
          hide: 0,
          label: null,
          name: 'datasource',
          options: [],
          query: 'prometheus',
          refresh: 1,
          regex: '',
          type: 'datasource',
        },
      )
      .addTemplate(
        template.new(
          'cluster',
          '$datasource',
          'label_values(agent_build_info, cluster)',
          refresh='time',
          current={
            selected: true,
            text: 'All',
            value: '$__all',
          },
          includeAll=true,
        ),
      )
      .addTemplate(
        template.new(
          'namespace',
          '$datasource',
          'label_values(agent_build_info, namespace)',
          refresh='time',
          current={
            selected: true,
            text: 'All',
            value: '$__all',
          },
          includeAll=true,
        ),
      )
      .addTemplate(
        template.new(
          'container',
          '$datasource',
          'label_values(agent_build_info, container)',
          refresh='time',
          current={
            selected: true,
            text: 'All',
            value: '$__all',
          },
          includeAll=true,
        ),
      )
      .addTemplate(
        template.new(
          'pod',
          '$datasource',
          'label_values(agent_build_info{container=~"$container"}, pod)',
          refresh='time',
          current={
            selected: true,
            text: 'All',
            value: '$__all',
          },
          includeAll=true,
        ),
      )
      .addTemplate(
        template.new(
          'url',
          '$datasource',
          'label_values(prometheus_remote_storage_shards{cluster=~"$cluster", pod=~"$pod"}, url)',
          refresh='time',
          includeAll=true,
        )
      )
      .addRow(
        row.new('Timestamps')
        .addPanel(timestampComparison)
        .addPanel(timestampComparisonRate)
      )
      .addRow(
        row.new('Samples')
        .addPanel(samplesRate)
      )
      .addRow(
        row.new('Shards')
        .addPanel(currentShards)
        .addPanel(maxShards)
        .addPanel(minShards)
        .addPanel(desiredShards)
      )
      .addRow(
        row.new('Shard Details')
        .addPanel(shardsCapacity)
        .addPanel(pendingSamples)
      )
      .addRow(
        row.new('Segments')
        .addPanel(queueSegment)
      )
      .addRow(
        row.new('Misc. Rates')
        .addPanel(droppedSamples)
        .addPanel(failedSamples)
        .addPanel(retriedSamples)
        .addPanel(enqueueRetries)
      ),
  },
}
