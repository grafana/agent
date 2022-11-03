{
  grafanaDashboards+:
    (import './dashboards/controller.libsonnet') +
    (import './dashboards/resources.libsonnet') +
    (import './dashboards/prometheus.remote_write.libsonnet'),
}
