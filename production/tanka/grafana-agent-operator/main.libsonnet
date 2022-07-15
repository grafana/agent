{
  new(name='grafana-agent-operator', namespace='')::
    (import './internal/base.libsonnet')(name, namespace),

  withGrafanaAgent(name='grafana-agent', namespace='')::
    (import './internal/grafana-agent.libsonnet')(name, namespace),

  withMetricsInstance(name='grafana-agent-metrics', namespace='', external_labels={})::
    (import './internal/metricsinstance.libsonnet')(name, namespace, external_labels),

  withLogsInstance(name='grafana-agent-logs', namespace='', external_labels={})::
    (import './internal/logsinstance.libsonnet')(name, namespace, external_labels),

  withK8sMetrics(namespace='', allowlist=false, allowlistmetrics=[])::
    (import './internal/k8smetrics.libsonnet')(namespace, allowlist, allowlistmetrics),

  #withKubeStateMetrics()::

  # TODO: make this a statefulset? use a PV?
  #withK8sEvents()::

  withDefaultLogs(name='kubernetes-pods', namespace='', container_engine='')::
    (import './internal/podlogs.libsonnet')(name, namespace, container_engine)
}
