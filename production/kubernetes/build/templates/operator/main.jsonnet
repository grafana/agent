local k = import 'ksonnet-util/kausal.libsonnet';
local agent_operator = import 'grafana-agent-operator/main.libsonnet';

{
  agent_operator:
    agent_operator.new(name='grafana-agent-operator', namespace='default') +
    agent_operator.withGrafanaAgent(name='grafana-agent', namespace='default') +
    agent_operator.withMetricsInstance(name='grafana-agent-metrics', namespace='default', external_labels={cluster: 'cloud'}) +
    agent_operator.withLogsInstance(name='grafana-agent-logs', namespace='default', external_labels={cluster: 'cloud'}) +
    agent_operator.withDefaultLogs(name='kubernetes-pods', namespace='default', container_engine='docker') +
    agent_operator.withK8sMetrics(namespace='default', allowlist=true, allowlistmetrics=["test_metric_1", "test_metric_2"])
}
