  function(name='kubernetes-events', namespace='', logs_name='grafana-agent-logs', logs_namespace='default', config={}) {    
    local this = self,
    local gen = import 'agent-operator-gen/main.libsonnet',
    local k = import 'ksonnet-util/kausal.libsonnet',
    
    local secret = k.core.v1.secret,
    local int = gen.monitoring.v1alpha1.integration,
    local ga = gen.monitoring.v1alpha1.grafanaAgent,

    _config:: {
        eh_labels: {agent: "grafana-agent-integrations"},
    } + config,

    eventhandler_integration:
        int.new('eventhandler') +
        int.metadata.withNamespace(namespace) +
        int.metadata.withLabels(this._config.eh_labels) +
        int.spec.type.withUnique(true)+
        int.spec.withConfig({
            logs_instance: logs_namespace + '/' + logs_name,
            cache_path: '/var/lib/grafana-agent/data/eventhandler.cache',  // TODO(hjet): use PVC
        }),
  }
