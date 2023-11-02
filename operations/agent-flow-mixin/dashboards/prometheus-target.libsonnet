local dashboard = import './utils/dashboard.jsonnet';
local panel = import './utils/panel.jsonnet';
local filename = 'agent-flow-prometheus-target.json';

{
  [filename]:
    dashboard.new(name='Grafana Agent Flow / Prometheus Target') +
    dashboard.withDocsLink(
      url='https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.scrape/',
      desc='Component documentation',
    ) +
    dashboard.withDashboardsLink() +
    dashboard.withUID(std.md5(filename)) +
    dashboard.withTemplateVariablesMixin([
      dashboard.newTemplateVariable('cluster', |||
        label_values(agent_component_controller_running_components, cluster)
      |||),
      dashboard.newMultiTemplateVariable('namespace', |||
        label_values(agent_component_controller_running_components{cluster="$cluster"}, namespace)
      |||),
      dashboard.newMultiTemplateVariable('job', |||
        label_values(agent_component_controller_running_components{cluster="$cluster", namespace=~"$namespace"}, job)
      |||),
      dashboard.newMultiTemplateVariable('instance', |||
        label_values(agent_component_controller_running_components{cluster="$cluster", namespace=~"$namespace"}, instance)
      |||),
      dashboard.newMultiTemplateVariable('scrape_job', |||
        label_values(prometheus_target_scrape_pool_targets{namespace=~"$namespace"}, scrape_job)
      |||),
    ]) +
    dashboard.withPanelsMixin([
      // Prometheus targets
      (
        panel.new(title='Targets', type='timeseries') +
        panel.withUnit('short') +
        panel.withDescription(|||
          Discovered targets by prometheus service discovery.
        |||) +
        panel.withPosition({ h: 9, w: 12, x: 0, y: 0 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              sum by (job) (
                prometheus_target_scrape_pool_targets{job=~"$job", namespace=~"$namespace", instance=~"$instance"}
              )
            |||,
            legendFormat='{{job}}',
          ),
        ])
      ),
      // Appended samples
      (
        panel.new(title='Appended Samples', type='timeseries') +
        panel.withUnit('short') +
        panel.withDescription(|||
          Total number of samples appended to the WAL.
        |||) +
        panel.withPosition({ h: 9, w: 12, x: 12, y: 0 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              sum by (job, instance_group_name) (
                rate(agent_wal_samples_appended_total{job=~"$job", namespace=~"$namespace", instance=~"$instance"}[$__rate_interval])
              )
            |||,
            legendFormat='{{job}} {{instance_group_name}}',
          ),
        ])
      ),
      // Scrape interval duration
      (
        panel.new(title='Average Scrape Interval Duration', type='timeseries') +
        panel.withUnit('short') +
        panel.withDescription(|||
          Actual intervals between scrapes.
        |||) +
        panel.withPosition({ h: 9, w: 8, x: 0, y: 9 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              rate(prometheus_target_interval_length_seconds_sum{job=~"$job", namespace=~"$namespace", instance=~"$instance"}[$__rate_interval]) / 
              rate(prometheus_target_interval_length_seconds_count{job=~"$job", namespace=~"$namespace", instance=~"$instance"}[$__rate_interval])
            |||,
            legendFormat='{{instance}} {{interval}} configured',
          ),
        ])
      ),
      // Scrape failures
      (
        panel.new(title='Scrape Failures', type='timeseries') +
        panel.withUnit('short') +
        panel.withDescription(|||
          Shows all scrape failures (sample limit exceeded, duplicate, out of bounds, out of order).
        |||) +
        panel.withPosition({ h: 9, w: 8, x: 8, y: 9 }) +
        panel.withQueries([
          panel.newQuery(
            expr=|||
              sum by (job) (
                rate(prometheus_target_scrapes_exceeded_sample_limit_total{job=~"$job", namespace=~"$namespace", instance=~"$instance"}[$__rate_interval])
              )
            |||,
            legendFormat='exceeded sample limit: {{job}}',
          ),
          panel.newQuery(
            expr=|||
              sum by (job) (
                rate(prometheus_target_scrapes_sample_duplicate_timestamp_total{job=~"$job", namespace=~"$namespace", instance=~"$instance"}[$__rate_interval])
              )
            |||,
            legendFormat='duplicate timestamp: {{job}}',
          ),
          panel.newQuery(
            expr=|||
              sum by (job) (
                rate(prometheus_target_scrapes_sample_out_of_bounds_total{job=~"$job", namespace=~"$namespace", instance=~"$instance"}[$__rate_interval])
              )
            |||,
            legendFormat='out of bounds: {{job}}',
          ),
          panel.newQuery(
            expr=|||
              sum by (job) (
                rate(prometheus_target_scrapes_sample_out_of_order_total{job=~"$job", namespace=~"$namespace", instance=~"$instance"}[$__rate_interval])
              )
            |||,
            legendFormat='out of order: {{job}}',
          ),
        ])
      ),
      // HTTP Scrape failures
      (
        panel.new(title='HTTP Scrape Failures', type='bargauge') {
          options: {
            orientation: 'vertical',
            showUnfilled: true,
          },
          fieldConfig: {
            defaults: {
              color: {
                mode: "fixed",
                fixedColor: "semi-dark-red",
              },
              mappings: [
                {
                  options: {
                    "0": {
                      color: "green",
                      index: 0
                    }
                  },
                  "type": "value"
                }
              ],
              min: 0,
              noValue: "0",
              thresholds: {
                mode: "absolute",
                steps: [
                  {
                    color: "green",
                    value: null
                  }
                ]
              }
            },
          },
        } +
        panel.withDescription(|||
          Breakdown of HTTP Scrape failures

          * Refused
          * Resolution
          * Timeout
        |||) +
        panel.withPosition({ h: 9, w: 8, x: 16, y: 9 }) +
        panel.withQueries([
          panel.newInstantQuery(
            legendFormat='Refused',
            expr='count by (reason) (
                rate(
                  net_conntrack_dialer_conn_failed_total{dialer_name=~"$scrape_job", job=~"$job", namespace=~"$namespace", instance=~"$instance", reason="refused"}[2m]
                ) > 0
              ) or label_replace(vector(0), "reason", "refused", "", "")',
          ),
          panel.newInstantQuery(
            legendFormat='Timeout',
            expr='count by (reason) (
                rate(
                  net_conntrack_dialer_conn_failed_total{dialer_name=~"$scrape_job", job=~"$job", namespace=~"$namespace", instance=~"$instance", reason="timeout"}[2m]
                ) > 0
              ) or label_replace(vector(0), "reason", "timeout", "", "")',
          ),
          panel.newInstantQuery(
            legendFormat='Resolution',
            expr='count by (reason) (
                rate(
                  net_conntrack_dialer_conn_failed_total{dialer_name=~"$scrape_job", job=~"$job", namespace=~"$namespace", instance=~"$instance", reason="resolution"}[2m]
                ) > 0
              ) or label_replace(vector(0), "reason", "resolution", "", "")',
          ),

        ])        
      ),
      // HTTP Scrape failures - table
      (
        panel.new('HTTP Scrape Failures', 'table') +
        panel.withDescription(|||
          Breakdown of HTTP Scrape failures

          * Refused
          * Resolution
          * Timeout
        |||) +
        panel.withPosition({ h: 9, w: 24, x: 0, y: 18 }) +
        panel.withQueries([
          panel.newInstantQuery(
            expr='count by (dialer_name, instance, job, namespace, reason) (
                rate(
                  net_conntrack_dialer_conn_failed_total{dialer_name=~"$scrape_job", job=~"$job", namespace=~"$namespace", instance=~"$instance", reason!~"unknown"}[2m]
                ) > 0
              )',
            format='table',
          ),
        ])
      )      
    ]),      
}
