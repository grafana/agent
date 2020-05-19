local dashboards = import 'dashboards.libsonnet';

{
  grafanaDashboards: std.mapWithKey(function(field, obj) obj {
    grafanaDashboardFolder: 'Agent',
  }, dashboards.grafanaDashboards),
}
