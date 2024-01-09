local utils = import './utils.libsonnet';
local g = import 'grafana-builder/grafana.libsonnet';
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
          g.queryPanel('sum by (pod) (prometheus_sd_discovered_targets{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"})', '{{pod}}') +
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
          g.queryPanel('sum by (job, instance_group_name) (rate(agent_wal_samples_appended_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m]))', '{{job}} {{instance_group_name}}') +
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
              ignoring(url, remote_name) group_right(pod)
              prometheus_remote_storage_queue_highest_sent_timestamp_seconds{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}
            )
          |||,
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local remoteSendLatency =
        graphPanel.new(
          'Latency [1m]',
          datasource='$datasource',
          span=6,
        )
        .addTarget(prometheus.target(
          'rate(prometheus_remote_storage_sent_batch_duration_seconds_sum{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[1m]) / rate(prometheus_remote_storage_sent_batch_duration_seconds_count{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[1m])',
          legendFormat='mean {{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ))
        .addTarget(prometheus.target(
          'histogram_quantile(0.99, rate(prometheus_remote_storage_sent_batch_duration_seconds_bucket{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[1m]))',
          legendFormat='p99 {{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local samplesInRate =
        graphPanel.new(
          'Rate in [5m]',
          datasource='$datasource',
          span=6,
        )
        .addTarget(prometheus.target(
          'rate(agent_wal_samples_appended_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local samplesOutRate =
        graphPanel.new(
          'Rate succeeded [5m]',
          datasource='$datasource',
          span=6,
        )
        .addTarget(prometheus.target(
          'rate(prometheus_remote_storage_succeeded_samples_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m]) or rate(prometheus_remote_storage_samples_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])',
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
          'prometheus_remote_storage_samples_pending{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}',
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
          span=6,
        )
        .addTarget(prometheus.target(
          'rate(prometheus_remote_storage_samples_dropped_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local failedSamples =
        graphPanel.new(
          'Failed Samples',
          datasource='$datasource',
          span=6,
        )
        .addTarget(prometheus.target(
          'rate(prometheus_remote_storage_samples_failed_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local retriedSamples =
        graphPanel.new(
          'Retried Samples',
          datasource='$datasource',
          span=6,
        )
        .addTarget(prometheus.target(
          'rate(prometheus_remote_storage_samples_retried_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      local enqueueRetries =
        graphPanel.new(
          'Enqueue Retries',
          datasource='$datasource',
          span=6,
        )
        .addTarget(prometheus.target(
          'rate(prometheus_remote_storage_enqueue_retries_total{cluster=~"$cluster", namespace=~"$namespace", container=~"$container"}[5m])',
          legendFormat='{{cluster}}:{{pod}}-{{instance_group_name}}-{{url}}',
        ));

      dashboard.new('Agent Prometheus Remote Write', tags=['grafana-agent-mixin'], editable=true, refresh='30s', time_from='now-1h')
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
        .addPanel(remoteSendLatency)
      )
      .addRow(
        row.new('Samples')
        .addPanel(samplesInRate)
        .addPanel(samplesOutRate)
        .addPanel(pendingSamples)
        .addPanel(droppedSamples)
        .addPanel(failedSamples)
        .addPanel(retriedSamples)
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
      )
      .addRow(
        row.new('Segments')
        .addPanel(queueSegment)
      )
      .addRow(
        row.new('Misc. Rates')
        .addPanel(enqueueRetries)
      ),

    'agent-tracing-pipeline.json':
      local acceptedSpans =
        graphPanel.new(
          'Accepted spans',
          datasource='$datasource',
          interval='1m',
          span=3,
          legend_show=false,
          fill=0,
        )
        .addTarget(prometheus.target(
          |||
            rate(traces_receiver_accepted_spans{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod",receiver!="otlp/lb"}[$__rate_interval])
          |||,
          legendFormat='{{ pod }} - {{ receiver }}/{{ transport }}',
        ));

      local refusedSpans =
        graphPanel.new(
          'Refused spans',
          datasource='$datasource',
          interval='1m',
          span=3,
          legend_show=false,
          fill=0,
        )
        .addTarget(prometheus.target(
          |||
            rate(traces_receiver_refused_spans{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod",receiver!="otlp/lb"}[$__rate_interval])
          |||,
          legendFormat='{{ pod }} - {{ receiver }}/{{ transport }}',
        ));

      local sentSpans =
        graphPanel.new(
          'Exported spans',
          datasource='$datasource',
          interval='1m',
          span=3,
          legend_show=false,
          fill=0,
        )
        .addTarget(prometheus.target(
          |||
            rate(traces_exporter_sent_spans{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod",exporter!="otlp"}[$__rate_interval])
          |||,
          legendFormat='{{ pod }} - {{ exporter }}',
        ));

      local exportedFailedSpans =
        graphPanel.new(
          'Exported failed spans',
          datasource='$datasource',
          interval='1m',
          span=3,
          legend_show=false,
          fill=0,
        )
        .addTarget(prometheus.target(
          |||
            rate(traces_exporter_send_failed_spans{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod",exporter!="otlp"}[$__rate_interval])
          |||,
          legendFormat='{{ pod }} - {{ exporter }}',
        ));

      local receivedSpans(receiverFilter, width) =
        graphPanel.new(
          'Received spans',
          datasource='$datasource',
          interval='1m',
          span=width,
          fill=1,
        )
        .addTarget(prometheus.target(
          |||
            sum(rate(traces_receiver_accepted_spans{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod",%s}[$__rate_interval]))
          ||| % receiverFilter,
          legendFormat='Accepted',
        ))
        .addTarget(prometheus.target(
          |||
            sum(rate(traces_receiver_refused_spans{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod",%s}[$__rate_interval]))
          ||| % receiverFilter,
          legendFormat='Refused',
        ));

      local exportedSpans(exporterFilter, width) =
        graphPanel.new(
          'Exported spans',
          datasource='$datasource',
          interval='1m',
          span=width,
          fill=1,
        )
        .addTarget(prometheus.target(
          |||
            sum(rate(traces_exporter_sent_spans{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod",%s}[$__rate_interval]))
          ||| % exporterFilter,
          legendFormat='Sent',
        ))
        .addTarget(prometheus.target(
          |||
            sum(rate(traces_exporter_send_failed_spans{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod",%s}[$__rate_interval]))
          ||| % exporterFilter,
          legendFormat='Send failed',
        ));

      local loadBalancedSpans =
        graphPanel.new(
          'Load-balanced spans',
          datasource='$datasource',
          interval='1m',
          span=3,
          fill=1,
          stack=true,
        )
        .addTarget(prometheus.target(
          |||
            rate(traces_loadbalancer_backend_outcome{cluster=~"$cluster",namespace=~"$namespace",success="true",container=~"$container",pod=~"$pod"}[$__rate_interval])
          |||,
          legendFormat='{{ pod }}',
        ));

      local peersNum =
        graphPanel.new(
          'Number of peers',
          datasource='$datasource',
          interval='1m',
          span=3,
          legend_show=false,
          fill=0,
        )
        .addTarget(prometheus.target(
          |||
            traces_loadbalancer_num_backends{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod"}
          |||,
          legendFormat='{{ pod }}',
        ));

      dashboard.new('Agent Tracing Pipeline', tags=['grafana-agent-mixin'], editable=true, refresh='30s', time_from='now-1h')
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
      .addRow(
        row.new('Write / Read')
        .addPanel(acceptedSpans)
        .addPanel(refusedSpans)
        .addPanel(sentSpans)
        .addPanel(exportedFailedSpans)
        .addPanel(receivedSpans('receiver!="otlp/lb"', 6))
        .addPanel(exportedSpans('exporter!="otlp"', 6))
      )
      .addRow(
        row.new('Load balancing')
        .addPanel(loadBalancedSpans)
        .addPanel(peersNum)
        .addPanel(receivedSpans('receiver="otlp/lb"', 3))
        .addPanel(exportedSpans('exporter="otlp"', 3))
      ),

    'agent-logs-pipeline.json':
      local sumByPodRateCounter(title, metric, format='short') =
        graphPanel.new(
          title,
          datasource='$datasource',
          interval='1m',
          span=6,
          fill=1,
          stack=true,
          format=format
        )
        .addTarget(prometheus.target(
          |||
            sum by($groupBy) (rate(%s{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod"}[$__rate_interval]))
          ||| % [metric],
          legendFormat='{{$groupBy}}',
        ));

      local sumByPodGague(title, metric) =
        graphPanel.new(
          title,
          datasource='$datasource',
          interval='1m',
          span=6,
          fill=1,
          stack=true,
        )
        .addTarget(prometheus.target(
          |||
            sum by($groupBy) (%s{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod"})
          ||| % [metric],
          legendFormat='{{$groupBy}}',
        ));

      local requestSuccessRate() =
        graphPanel.new(
          'Write requests success rate [%]',
          datasource='$datasource',
          interval='1m',
          fill=0,
          span=6,
          format='%',
        )
        .addTarget(prometheus.target(
          |||
            sum by($groupBy) (rate(promtail_request_duration_seconds_bucket{status_code=~"2..", cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod"}[$__rate_interval]))
            /
            sum by($groupBy) (rate(promtail_request_duration_seconds_bucket{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod"}[$__rate_interval]))
            * 100
          |||,
          legendFormat='{{$groupBy}}',
        ));

      local histogramQuantile(title, metric, q) =
        graphPanel.new(
          title,
          datasource='$datasource',
          interval='1m',
          span=6,
          fill=0,
          format='s',
        )
        .addTarget(prometheus.target(
          |||
            histogram_quantile(
              %f, 
              sum by (le, $groupBy)
              (rate(%s{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod"}[$__rate_interval]))
            )
          ||| % [q, metric],
          legendFormat='{{$groupBy}}',
        ));

      local histogramAverage(title, metric) =
        graphPanel.new(
          title,
          datasource='$datasource',
          interval='1m',
          span=6,
          fill=0,
          format='s',
        )
        .addTarget(prometheus.target(
          |||
            (sum by (le, $groupBy) (rate(%s_sum{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod"}[$__rate_interval])))
            /
            (sum by (le, $groupBy) (rate(%s_count{cluster=~"$cluster",namespace=~"$namespace",container=~"$container",pod=~"$pod"}[$__rate_interval])))
          ||| % [metric, metric],
          legendFormat='{{$groupBy}}',
        ));


      dashboard.new('Agent Logs Pipeline', tags=['grafana-agent-mixin'], editable=true, refresh='30s', time_from='now-1h')
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
        template.custom(
          'groupBy',
          'pod,cluster,namespace',
          'pod',
        ),
      )
      .addRow(
        row.new('Errors', height=500)
        .addPanel(sumByPodRateCounter('Dropped bytes rate [B/s]', 'promtail_dropped_bytes_total', format='Bps'))
        .addPanel(requestSuccessRate())
      )
      .addRow(
        row.new('Latencies', height=500)
        .addPanel(histogramQuantile('Write latencies p99 [s]', 'promtail_request_duration_seconds_bucket', 0.99))
        .addPanel(histogramQuantile('Write latencies p90 [s]', 'promtail_request_duration_seconds_bucket', 0.90))
        .addPanel(histogramQuantile('Write latencies p50 [s]', 'promtail_request_duration_seconds_bucket', 0.50))
        .addPanel(histogramAverage('Write latencies average [s]', 'promtail_request_duration_seconds'))
      )
      .addRow(
        row.new('Logs volume', height=500)
        .addPanel(sumByPodRateCounter('Bytes read rate [B/s]', 'promtail_read_bytes_total', format='Bps'))
        .addPanel(sumByPodRateCounter('Lines read rate [lines/s]', 'promtail_read_lines_total'))
        .addPanel(sumByPodGague('Active files count', 'promtail_files_active_total'))
        .addPanel(sumByPodRateCounter('Entries sent rate [entries/s]', 'promtail_sent_entries_total'))
      ),
  },
}
