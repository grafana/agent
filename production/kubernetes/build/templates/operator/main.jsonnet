local k = import 'ksonnet-util/kausal.libsonnet';
local agent_operator = import 'grafana-agent-operator/main.libsonnet';
local gen = import 'agent-operator-gen/main.libsonnet';
local ga = gen.monitoring.v1alpha1.grafanaAgent;
local li = gen.monitoring.v1alpha1.logsInstance;
local mi = gen.monitoring.v1alpha1.metricsInstance;

{
  operator:
    agent_operator.new(name='grafana-agent-operator', namespace='default'),
  
  grafanaAgent:
    agent_operator.withGrafanaAgent(name='grafana-agent', namespace='default') + {
      resource+:
        // add metrics external labels
        ga.spec.metrics.withExternalLabels($.metricsInstance._config.external_labels) + 
        // pick up metrics instance defined below
        ga.spec.metrics.instanceSelector.withMatchLabels($.metricsInstance._config.mi_labels) +
        // pick up logs instance defined below
        ga.spec.logs.instanceSelector.withMatchLabels($.logsInstance._config.li_labels) +
        // pick up events integration instance defined below
        ga.spec.integrations.selector.withMatchLabels($.events._config.eh_labels)
    },
  
  metricsInstance:
    agent_operator.withMetricsInstance(name='grafana-agent-metrics', namespace='default') + {
      resource+:
        // pick up svc monitors defined below
        mi.spec.serviceMonitorSelector.withMatchLabels($.monitors._config.monitor_labels) +
        // pick up svc monitors in all namespaces
        {
          spec+: {
            serviceMonitorNamespaceSelector: {}
          }
        }
    },

  logsInstance:
    agent_operator.withLogsInstance(name='grafana-agent-logs', namespace='default') + {
      resource+:
        // pick up podlogs resource defined below
        li.spec.podLogsSelector.withMatchLabels($.podLogs._config.podlogs_labels) + 
        // select podlogs across all namespaces
        {
          spec+: {
            podLogsNamespaceSelector: {}
          }
        }
    },
    
  podLogs:
    agent_operator.withDefaultLogs(name='kubernetes-pods', namespace='default', containerEngine='docker'),
    
  monitors:
    agent_operator.withK8sMetrics(namespace='default', allowlist=false, allowlistMetrics=[]),
    
  kubeStateMetrics:
    agent_operator.withKSM(name='kube-state-metrics', namespace='default'),
  
  events:
    agent_operator.withK8sEvents(name='kubernetes-events', namespace='default', logsName='grafana-agent-logs', logsNamespace='default'),
}
