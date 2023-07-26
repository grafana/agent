local gen = import 'agent-operator-gen/main.libsonnet';
local pl = gen.monitoring.v1alpha1.podLogs;
local r = pl.spec.relabelings;

{
    withK8sLogsRelabeling():: [
        r.withSourceLabels(['__meta_kubernetes_pod_node_name']) +
        r.withTargetLabel('__host__'),

        // r.withAction('replace') +
        // r.withReplacement('$1') +
        // r.withSeparator('/') +
        // r.withSourceLabels(['__meta_kubernetes_namespace', '__meta_kubernetes_pod_name']) +
        // r.withTargetLabel('job'),

        r.withAction('replace') +
        r.withSourceLabels('__meta_kubernetes_namespace') +
        r.withTargetLabel('namespace'),

        r.withAction('replace') +
        r.withSourceLabels('__meta_kubernetes_pod_name') +
        r.withTargetLabel('pod'),

        r.withAction('replace') +
        r.withSourceLabels('__meta_kubernetes_pod_container_name') +
        r.withTargetLabel('container'),

        r.withReplacement('/var/log/pods/*$1/*.log') +
        r.withSeparator('/') +
        r.withSourceLabels(['__meta_kubernetes_pod_uid', '__meta_kubernetes_pod_container_name']) +
        r.withTargetLabel('__path__')
    ]
}
