local k = import 'ksonnet-util/kausal.libsonnet';
local secret = k.core.v1.secret;
local pvc = k.core.v1.persistentVolumeClaim;

local gen = import 'agent-operator-gen/main.libsonnet';
local ga = gen.monitoring.v1alpha1.grafanaAgent;
local mi = gen.monitoring.v1alpha1.metricsInstance;
local li = gen.monitoring.v1alpha1.logsInstance;
local pl = gen.monitoring.v1alpha1.podLogs;
local int = gen.monitoring.v1alpha1.integration;

local op = import 'grafana-agent-operator/operator.libsonnet';
local ga_util = import 'grafana-agent-operator/util/grafana-agent.libsonnet';
local mi_util = import 'grafana-agent-operator/util/metricsinstance.libsonnet';
local li_util = import 'grafana-agent-operator/util/logsinstance.libsonnet';
local pl_util = import 'grafana-agent-operator/util/k8slogs.libsonnet';
local mon_util = import 'grafana-agent-operator/util/k8smonitors.libsonnet';
local int_util = import 'grafana-agent-operator/util/integrations.libsonnet';

local ksm = import 'kube-state-metrics/kube-state-metrics.libsonnet';

{
  local this = self,

  _images:: {
    agent: 'grafana/agent:v0.32.1',
    agent_operator: 'grafana/agent-operator:v0.32.1',
    ksm: 'registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.5.0',
  },

  _config:: {
    namespace: '${NAMESPACE}',
    metrics_url: '${METRICS_URL}',
    metrics_user: '${METRICS_USER}',
    metrics_key: '${METRICS_KEY}',
    logs_url: '${LOGS_URL}',
    logs_user: '${LOGS_USER}',
    logs_key: '${LOGS_KEY}',
    cluster_label: { cluster: '${CLUSTER}' },
    kubelet_job: 'kubelet',
    cadvisor_job: 'cadvisor',
    ksm_job: 'kube-state-metrics',
    ksm_version: '2.5.0',
  },

  operator:
    op.new(name='grafana-agent-operator', namespace=this._config.namespace, image=this._images.agent_operator, serviceAccount='grafana-agent-operator') +
    op.withRbac(name='grafana-agent-operator', namespace=this._config.namespace),

  grafana_agent:
    ga.new(name='grafana-agent') +
    ga.metadata.withNamespace(this._config.namespace) +
    ga.spec.withServiceAccountName('grafana-agent') +
    ga.spec.withImage(this._images.agent) +
    ga.spec.metrics.instanceSelector.withMatchLabels({ agent: 'grafana-agent' }) +
    ga.spec.logs.instanceSelector.withMatchLabels({ agent: 'grafana-agent' }) +
    ga.spec.integrations.selector.withMatchLabels({ agent: 'grafana-agent' }) +
    ga.spec.metrics.withExternalLabels(this._config.cluster_label),
  rbac:
    ga_util.withRbac(name='grafana-agent', namespace=this._config.namespace),

  metrics_instance:
    mi.new(name='grafana-agent-metrics') +
    mi.metadata.withNamespace(this._config.namespace) +
    mi.metadata.withLabels({ agent: 'grafana-agent' }) +
    mi.spec.serviceMonitorSelector.withMatchLabels({ instance: 'primary' }) +
    mi_util.withRemoteWrite(secretName='metrics-secret', metricsUrl=this._config.metrics_url) +
    mi_util.withNilServiceMonitorNamespace(),
  metrics_secret:
    secret.new('metrics-secret', {}) +
    secret.withStringData({
      username: this._config.metrics_user,
      password: this._config.metrics_key,
    }) + secret.mixin.metadata.withNamespace(this._config.namespace),

  logs_instance:
    li.new(name='grafana-agent-logs') +
    li.metadata.withNamespace(this._config.namespace) +
    li.metadata.withLabels({ agent: 'grafana-agent' }) +
    li.spec.podLogsSelector.withMatchLabels({ instance: 'primary' }) +
    li_util.withLogsClient(secretName='logs-secret', logsUrl=this._config.logs_url, externalLabels=this._config.cluster_label) +
    li_util.withNilPodLogsNamespace(),
  logs_secret:
    secret.new('logs-secret', {}) +
    secret.withStringData({
      username: this._config.logs_user,
      password: this._config.logs_key,
    }) + secret.mixin.metadata.withNamespace(this._config.namespace),

  pod_logs:
    pl.new('kubernetes-logs') +
    pl.metadata.withNamespace(this._config.namespace) +
    pl.metadata.withLabels({ instance: 'primary' }) +
    pl.spec.withPipelineStages(pl.spec.pipelineStages.withCri({})) +
    pl.spec.namespaceSelector.withAny(true) +
    pl.spec.selector.withMatchLabels({}) +
    pl.spec.withRelabelings(pl_util.withK8sLogsRelabeling()),

  k8s_monitors: [
    mon_util.newKubernetesMonitor(
      name='kubelet-monitor',
      namespace=this._config.namespace,
      monitorLabels={ instance: 'primary' },
      targetNamespace=this._config.namespace,
      targetLabels={ 'app.kubernetes.io/name': 'kubelet' },
      jobLabel=this._config.kubelet_job,
      metricsPath='/metrics',
      allowlist=false,
      allowlistMetrics=[]
    ),
    mon_util.newKubernetesMonitor(
      name='cadvisor-monitor',
      namespace='default',
      monitorLabels={ instance: 'primary' },
      targetNamespace=this._config.namespace,
      targetLabels={ 'app.kubernetes.io/name': 'kubelet' },
      jobLabel=this._config.cadvisor_job,
      metricsPath='/metrics/cadvisor',
      allowlist=false,
      allowlistMetrics=[]
    ),
    mon_util.newServiceMonitor(
      name='ksm-monitor',
      namespace=this._config.namespace,
      monitorLabels={ instance: 'primary' },
      targetNamespace=this._config.namespace,
      targetLabels={ 'app.kubernetes.io/name': 'kube-state-metrics' },
      jobLabel=this._config.ksm_job,
      metricsPath='/metrics',
      allowlist=false,
      allowlistMetrics=[]
    ),
  ],

  kube_state_metrics:
    ksm {
      name:: 'kube-state-metrics',
      namespace:: this._config.namespace,
      version:: this._config.ksm_version,
      image:: this._images.ksm,
    },

  events:
    int.new('agent-eventhandler') +
    int.metadata.withNamespace(this._config.namespace) +
    int.metadata.withLabels({ agent: 'grafana-agent' }) +
    int.spec.withName('eventhandler') +
    int.spec.type.withUnique(true) +
    int.spec.withConfig({
      logs_instance: this._config.namespace + '/' + 'grafana-agent-logs',
      cache_path: '/etc/eventhandler/eventhandler.cache',
    }) +
    int_util.withPVC('agent-eventhandler'),
  pvc:
    pvc.new('agent-eventhandler') +
    pvc.mixin.metadata.withNamespace(this._config.namespace) +
    pvc.mixin.spec.withAccessModes('ReadWriteOnce') +
    pvc.mixin.spec.resources.withRequests({ storage: '1Gi' }),

}
