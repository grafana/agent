local gen = import 'agent-operator-gen/main.libsonnet';
local pl = gen.monitoring.v1alpha1.podLogs;
local li = gen.monitoring.v1alpha1.logsInstance;

{
    new(name, namespace, podlogsLabels, containerEngine='docker'):: {    

        local pipelineStage = if containerEngine =='cri' then pl.spec.pipelineStages.withCri({}) else pl.spec.pipelineStages.withDocker({}),

        resource: pl.new(name) +
            pl.metadata.withNamespace(namespace) +
            pl.metadata.withLabels(podlogsLabels) +
            pl.spec.withPipelineStages(pipelineStage) +
            pl.spec.namespaceSelector.withAny(true) +
            pl.spec.selector.withMatchLabels({}),
    }
}
