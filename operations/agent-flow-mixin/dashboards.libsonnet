{
  grafanaDashboards+:
    (import './dashboards/controller.libsonnet') +
    (import './dashboards/resources.libsonnet'),
}
