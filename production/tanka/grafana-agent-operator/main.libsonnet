{
  new(name='grafana-agent-operator', namespace='')::
    (import './internal/operator.libsonnet')(name, namespace),

  withGrafanaAgent(name='grafana-agent', namespace='')::
    (import './internal/grafana-agent.libsonnet')(name, namespace),

  withMetricsInstance(name='grafana-agent-metrics', namespace='', config={})::
    (import './internal/metricsinstance.libsonnet')(name, namespace, config),

  withLogsInstance(name='grafana-agent-logs', namespace='', config={})::
    (import './internal/logsinstance.libsonnet')(name, namespace, config),

  withK8sMetrics(namespace='', allowlist=false, allowlistMetrics=[], config={})::
    (import './internal/k8smetrics.libsonnet')(namespace, allowlist, allowlistMetrics, config),

  withKSM(name='kube-state-metrics', namespace='')::
    (import './internal/ksm.libsonnet')(name, namespace),

  withK8sEvents(name='kubernetes-events', namespace='', logsName='', logsNamespace='', config={})::
    (import './internal/k8s-events.libsonnet')(name, namespace, logsName, logsNamespace, config),

  withDefaultLogs(name='kubernetes-pods', namespace='', containerEngine='')::
    (import './internal/podlogs.libsonnet')(name, namespace, containerEngine)
}
