function(name='kubernetes-pods', namespace='', container_engine='docker') {    
    local this = self,
    local gen = import 'agent-operator-gen/main.libsonnet',
    local k = import 'ksonnet-util/kausal.libsonnet',
    
    local pl = gen.monitoring.v1alpha1.podLogs,
    local li = gen.monitoring.v1alpha1.logsInstance,

    _config:: {
        podlogs_labels: {agent: 'grafana-agent-logs'},
    },

    local pipeline_stage = if container_engine =='cri' then pl.spec.pipelineStages.withCri({}) else pl.spec.pipelineStages.withDocker({}),

    resource: pl.new(name) +
        pl.metadata.withNamespace(namespace) +
        pl.metadata.withLabels(this._config.podlogs_labels) +
        pl.spec.withPipelineStages(pipeline_stage) +
        pl.spec.namespaceSelector.withAny(true) +
        pl.spec.selector.withMatchLabels({}),

}
