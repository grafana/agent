local dashboards = import 'dashboards.libsonnet';

{
  grafanaDashboards+:: std.mapWithKey(function(field, obj) obj {
    grafanaDashboardFolder: 'Grafana Agent',
  }, dashboards.grafanaDashboards),
}
