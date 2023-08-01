local dashboard = import './utils/dashboard.jsonnet';
local panel = import './utils/panel.jsonnet';
local filename = 'agent-flow-prometheus-remote-write.json';

local stackedPanelMixin = {
  fieldConfig+: {
    defaults+: {
      custom+: {
        fillOpacity: 20,
        gradientMode: 'hue',
        stacking: { mode: 'normal' },
      },
    },
  },
};

{
  [filename]:
    dashboard.new(name='Grafana Agent Flow / prometheus.remote_write') +
    dashboard.withDocsLink(
      url='https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.remote_write/',
      desc='Component documentation',
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
      dashboard.newMultiTemplateVariable('instance', |||
        label_values(agent_component_controller_running_components{cluster="$cluster", namespace="$namespace"}, instance)
      |||),
      dashboard.newMultiTemplateVariable('component', |||
        label_values(agent_wal_samples_appended_total{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"prometheus\\.remote_write\\..*"}, component_id)
      |||),
      dashboard.newMultiTemplateVariable('url', |||
        label_values(prometheus_remote_storage_sent_batch_duration_seconds_sum{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component"}, url)
      |||),
    ]) +
    // TODO(@tpaschalis) Make the annotation optional.
    dashboard.withAnnotations([
      dashboard.newLokiAnnotation('Deployments', '{cluster="$cluster", container="kube-diff-logger"} | json | namespace_extracted="grafana-agent" | name_extracted=~"grafana-agent.*"', 'rgba(0, 211, 255, 1)'),
    ]) +
    dashboard.withPanelsMixin([
      // WAL delay
      (
        panel.new(title='WAL delay', type='timeseries') +
        panel.withUnit('s') +
        panel.withDescription(|||
          How far behind prometheus.remote_write from samples recently written
          to the WAL.

          Each endpoint prometheus.remote_write is configured to send metrics
          has its own delay. The time shown here is the sum across all
          endpoints for the given component.

          It is normal for the WAL delay to be within 1-3 scrape intervals. If
          the WAL delay continues to increase beyond that amount, try
          increasing the number of maximum shards.
        |||) +
        panel.withPosition({ x: 0, y: 0, w: 6, h: 10 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              sum by (instance, component_id) (
                prometheus_remote_storage_highest_timestamp_in_seconds{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component"}
                - ignoring(url, remote_name) group_right(instance)
                prometheus_remote_storage_queue_highest_sent_timestamp_seconds{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component", url=~"$url"}
              )
            |||,
            legendFormat='{{instance}} / {{component_id}}',
          ),
        ])
      ),

      // Data write throughput
      (
        panel.new(title='Data write throughput', type='timeseries') +
        stackedPanelMixin +
        panel.withUnit('Bps') +
        panel.withDescription(|||
          Rate of data containing samples and metadata sent by
          prometheus.remote_write.
        |||) +
        panel.withPosition({ x: 6, y: 0, w: 6, h: 10 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              sum without (remote_name, url) (
                  rate(prometheus_remote_storage_bytes_total{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component", url=~"$url"}[$__rate_interval]) +
                  rate(prometheus_remote_storage_metadata_bytes_total{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component", url=~"$url"}[$__rate_interval])
              )
            |||,
            legendFormat='{{instance}} / {{component_id}}',
          ),
        ])
      ),

      // Write latency
      (
        panel.new(title='Write latency', type='timeseries') +
        panel.withUnit('s') +
        panel.withDescription(|||
          Latency of writes to the remote system made by
          prometheus.remote_write.
        |||) +
        panel.withPosition({ x: 12, y: 0, w: 6, h: 10 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              histogram_quantile(0.99, sum by (le) (
                rate(prometheus_remote_storage_sent_batch_duration_seconds_bucket{cluster="$cluster",namespace="$namespace",instance=~"$instance", component_id=~"$component", url=~"$url"}[$__rate_interval])
              ))
            |||,
            legendFormat='99th percentile',
          ),
          panel.newQuery(
            expr=|||
              histogram_quantile(0.50, sum by (le) (
                rate(prometheus_remote_storage_sent_batch_duration_seconds_bucket{cluster="$cluster",namespace="$namespace",instance=~"$instance", component_id=~"$component", url=~"$url"}[$__rate_interval])
              ))
            |||,
            legendFormat='50th percentile',
          ),
          panel.newQuery(
            expr=|||
              sum(rate(prometheus_remote_storage_sent_batch_duration_seconds_sum{cluster="$cluster",namespace="$namespace",instance=~"$instance", component_id=~"$component"}[$__rate_interval])) /
              sum(rate(prometheus_remote_storage_sent_batch_duration_seconds_count{cluster="$cluster",namespace="$namespace",instance=~"$instance", component_id=~"$component"}[$__rate_interval]))
            |||,
            legendFormat='Average',
          ),
        ])
      ),

      // Shards
      (
        local minMaxOverride = {
          properties: [{
            id: 'custom.lineStyle',
            value: {
              dash: [10, 15],
              fill: 'dash',
            },
          }, {
            id: 'custom.showPoints',
            value: 'never',
          }, {
            id: 'custom.hideFrom',
            value: {
              legend: true,
              tooltip: false,
              viz: false,
            },
          }],
        };

        panel.new(title='Shards', type='timeseries') {
          fieldConfig+: {
            overrides: [
              minMaxOverride { matcher: { id: 'byName', options: 'Minimum' } },
              minMaxOverride { matcher: { id: 'byName', options: 'Maximum' } },
            ],
          },
        } +
        panel.withUnit('none') +
        panel.withDescription(|||
          Total number of shards which are concurrently sending samples read
          from the Write-Ahead Log.

          Shards are bound to a minimum and maximum, displayed on the graph.
          The lowest minimum and the highest maximum across all clients is
          shown.

          Each client has its own set of shards, minimum shards, and maximum
          shards; filter to a specific URL to display more granular
          information.
        |||) +
        panel.withPosition({ x: 18, y: 0, w: 6, h: 10 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              sum without (remote_name, url) (
                  prometheus_remote_storage_shards{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component", url=~"$url"}
              )
            |||,
            legendFormat='{{instance}} / {{component_id}}',
          ),
          panel.newQuery(
            expr=|||
              min (
                  prometheus_remote_storage_shards_min{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component", url=~"$url"}
              )
            |||,
            legendFormat='Minimum',
          ),
          panel.newQuery(
            expr=|||
              max (
                  prometheus_remote_storage_shards_max{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component", url=~"$url"}
              )
            |||,
            legendFormat='Maximum',
          ),
        ])
      ),


      // Sent samples / second
      (
        panel.new(title='Sent samples / second', type='timeseries') +
        stackedPanelMixin +
        panel.withUnit('cps') +
        panel.withDescription(|||
          Total outgoing samples sent by prometheus.remote_write.
        |||) +
        panel.withPosition({ x: 0, y: 10, w: 8, h: 10 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              sum without (url, remote_name) (
                rate(prometheus_remote_storage_samples_total{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component", url=~"$url"}[$__rate_interval])
              )
            |||,
            legendFormat='{{instance}} / {{component_id}}',
          ),
        ])
      ),

      // Failed samples / second
      (
        panel.new(title='Failed samples / second', type='timeseries') +
        stackedPanelMixin +
        panel.withUnit('cps') +
        panel.withDescription(|||
          Rate of samples which prometheus.remote_write could not send due to
          non-recoverable errors.
        |||) +
        panel.withPosition({ x: 8, y: 10, w: 8, h: 10 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              sum without (url,remote_name) (
                rate(prometheus_remote_storage_samples_failed_total{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component", url=~"$url"}[$__rate_interval])
              )
            |||,
            legendFormat='{{instance}} / {{component_id}}',
          ),
        ])
      ),

      // Retried samples / second
      (
        panel.new(title='Retried samples / second', type='timeseries') +
        stackedPanelMixin +
        panel.withUnit('cps') +
        panel.withDescription(|||
          Rate of samples which prometheus.remote_write attempted to resend
          after receiving a recoverable error.
        |||) +
        panel.withPosition({ x: 16, y: 10, w: 8, h: 10 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              sum without (url,remote_name) (
                rate(prometheus_remote_storage_samples_retried_total{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component", url=~"$url"}[$__rate_interval])
              )
            |||,
            legendFormat='{{instance}} / {{component_id}}',
          ),
        ])
      ),

      // Active series (Total)
      (
        panel.new(title='Active series (total)', type='timeseries') {
          options+: {
            legend+: {
              showLegend: false,
            },
          },
        } +
        panel.withUnit('short') +
        panel.withDescription(|||
          Total number of active series across all components.

          An "active series" is a series that prometheus.remote_write recently
          received a sample for. Active series are garbage collected whenever a
          truncation of the WAL occurs.
        |||) +
        panel.withPosition({ x: 0, y: 20, w: 12, h: 10 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              sum(agent_wal_storage_active_series{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component", url=~"$url"})
            |||,
            legendFormat='Series',
          ),
        ])
      ),

      // Active series (by component)
      (
        panel.new(title='Active series (by component)', type='timeseries') +
        panel.withUnit('short') +
        panel.withDescription(|||
          Total number of active series which are currently being tracked by
          prometheus.remote_write components.

          An "active series" is a series that prometheus.remote_write recently
          received a sample for. Active series are garbage collected whenever a
          truncation of the WAL occurs.
        |||) +
        panel.withPosition({ x: 12, y: 20, w: 12, h: 10 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              agent_wal_storage_active_series{cluster="$cluster", namespace="$namespace", instance=~"$instance", component_id=~"$component", url=~"$url"}
            |||,
            legendFormat='{{instance}} / {{component_id}}',
          ),
        ])
      ),


    ]),
}
