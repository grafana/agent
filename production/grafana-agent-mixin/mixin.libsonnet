local dashboards = import 'dashboards.libsonnet';
local debugging = import 'debugging.libsonnet';

{
  grafanaDashboards+:: std.mapWithKey(function(field, obj) obj {
    grafanaDashboardFolder: 'Grafana Agent',
  }, dashboards.grafanaDashboards + debugging.grafanaDashboards),
} + (import 'alerts.libsonnet')
