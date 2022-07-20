local k = import 'ksonnet-util/kausal.libsonnet';
local op = import 'grafana-agent-operator/operator.libsonnet';
local ga = import 'grafana-agent-operator/grafana-agent.libsonnet';
local mi = import 'grafana-agent-operator/metricsinstance.libsonnet';
local li = import 'grafana-agent-operator/logsinstance.libsonnet';
local pl = import 'grafana-agent-operator/podlogs.libsonnet';
local mon = import 'grafana-agent-operator/k8smonitors.libsonnet';
local ksm = import 'grafana-agent-operator/ksm.libsonnet';
local ev = import 'grafana-agent-operator/k8s-events.libsonnet';

## TODO: pod logs & metrics relabeling to match what we currently do / enable correlation
{
  operator:
    op.new(name='grafana-agent-operator', namespace='default'),
  
  grafana_agent:
    ga.new(name='grafana-agent', namespace='default') +
    ga.withMetrics(metricsInstanceLabels={agent: 'grafana-agent'}) +
    ga.withMetricsExternalLabels(externalLabels={cluster: 'cloud'}) +
    ga.withLogs(logsInstanceLabels={agent: 'grafana-agent'}) +
    ga.withIntegration(integrationLabels={agent: 'grafana-agent'}),
  
  metrics_instance:
    mi.new(name='grafana-agent-metrics', namespace='default', metricsInstanceLabels={agent: 'grafana-agent'}) +
    mi.withRemoteWrite(namespace='default', secretName='metrics-secret', metricsUrl='test', metricsUser='test', metricsKey='test') +
    mi.withServiceMonitor({instance: 'primary'}) +
    mi.withNilServiceMonitorNamespace(),

  logs_instance:
    li.new(name='grafana-agent-logs', namespace='default', logsInstanceLabels={agent: 'grafana-agent'}) +
    li.withLogsClient(namespace='default', secretName='logs-secret', logsUrl='test', logsUser='test', logsKey='test', externalLabels={cluster: 'cloud'}) +
    li.withPodLogs({instance: 'primary'}) +
    li.withNilPodLogsNamespace(),
    
  pod_logs:
    pl.new(name='kubernetes-pods', namespace='default', podlogsLabels={instance: 'primary'}, containerEngine='docker'),

  k8s_monitors: [
    mon.newKubernetesMonitor(
      name='kubelet-monitor', 
      namespace='default', 
      monitorLabels={instance: 'primary'},
      targetNamespace='default',
      targetLabels={'app.kubernetes.io/name': 'kubelet'},
      jobLabel='integrations/kubernetes/kubelet',
      metricsPath='/metrics',
      allowlist=false, 
      allowlistMetrics=[]),
    mon.newKubernetesMonitor(
      name='cadvisor-monitor', 
      namespace='default', 
      monitorLabels={instance: 'primary'},
      targetNamespace='default',
      targetLabels={'app.kubernetes.io/name': 'kubelet'},
      jobLabel='integrations/kubernetes/cadvisor',
      metricsPath='/metrics/cadvisor',
      allowlist=false, 
      allowlistMetrics=[]),
    mon.newMonitor(
      name='ksm-monitor',
      namespace='default', 
      monitorLabels={instance: 'primary'},
      targetNamespace='default', 
      targetLabels={'app.kubernetes.io/name': 'kube-state-metrics'},
      jobLabel='integrations/kubernetes/kube-state-metrics',
      metricsPath='/metrics',
      allowlist=false, 
      allowlistMetrics=[]),
  ],
    
  kube_state_metrics:
    ksm.new(name='kube-state-metrics', namespace='default'),
  
  events:
    ev.new(name='grafana-agent-events', namespace='default', integrationLabels={agent: 'grafana-agent'}, logsInstanceName='grafana-agent', logsInstanceNamespace='default') +
    ev.withPVC('agent-eventhandler', 'default')

}
